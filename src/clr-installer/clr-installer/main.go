package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"clr-installer/cmd"
	"clr-installer/frontend"
	"clr-installer/log"
	"clr-installer/massinstall"
	"clr-installer/model"
	"clr-installer/tui"
	flag "github.com/spf13/pflag"
)

var (
	version    bool
	reboot     bool
	logFile    string
	configFile string
	logLevel   int

	frontEndImpls []frontend.Frontend
)

func init() {
	flag.BoolVar(
		&version, "version", false, "Version of the Installer",
	)

	flag.BoolVar(
		&reboot, "reboot", true, "Reboot after finishing",
	)

	flag.StringVar(
		&configFile, "config", "", "Installation configuration file",
	)

	flag.IntVar(
		&logLevel,
		"log-level",
		log.LogLevelDebug,
		fmt.Sprintf("%d (debug), %d (info), %d (warning), %d (error)",
			log.LogLevelDebug, log.LogLevelInfo, log.LogLevelWarning, log.LogLevelError),
	)

	usr, err := user.Current()
	if err != nil {
		fatal(err)
	}

	defaultLogFile := filepath.Join(usr.HomeDir, "clr-installer.log")
	flag.StringVar(&logFile, "log-file", defaultLogFile, "The log file path")
}

func fatal(err error) {
	log.ErrorError(err)
	panic(err)
}

func initFrontendList() {
	frontEndImpls = []frontend.Frontend{
		massinstall.New(configFile != ""),
		tui.New(),
	}
}

func lookupDefaultConfig() (string, error) {
	config := "/usr/share/defaults/clr-installer/clr-installer.yaml"

	src, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}

	// use the config from source code's etc dir
	if strings.Contains(src, "/clr-installer/bin") {
		config = filepath.Join(strings.Replace(src, "bin", "etc", 1), "clr-installer.yaml")
	}

	return config, nil
}

func main() {
	flag.Parse()

	if err := log.SetLogLevel(logLevel); err != nil {
		fatal(err)
	}

	if version {
		fmt.Println(path.Base(os.Args[0]) + ": " + model.Version)
		return
	}

	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fatal(err)
	}
	defer func() {
		_ = f.Close()
	}()
	log.SetOutput(f)

	initFrontendList()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	rootDir, err := ioutil.TempDir("", "install-")
	if err != nil {
		fatal(err)
	}

	var md *model.SystemInstall
	cf := configFile

	if configFile == "" {
		if cf, err = lookupDefaultConfig(); err != nil {
			fatal(err)
		}
	}

	log.Debug("Loading config file: %s", cf)
	if md, err = model.LoadFile(cf); err != nil {
		fatal(err)
	}

	installReboot := false

	go func() {
		for _, fe := range frontEndImpls {
			if !fe.MustRun() {
				continue
			}

			installReboot, err = fe.Run(md, rootDir)
			if err != nil {
				fatal(err)
			}

			break
		}

		done <- true
	}()

	go func() {
		<-sigs
		fmt.Println("Leaving...")
		done <- true
	}()

	<-done

	if reboot && installReboot {
		if err := cmd.RunAndLog("reboot"); err != nil {
			fatal(err)
		}
	}
}
