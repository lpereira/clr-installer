package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"path"
	"path/filepath"
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
		massinstall.New(configFile),
		tui.New(),
	}
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

	go func() {
		for _, fe := range frontEndImpls {
			if !fe.MustRun() {
				continue
			}

			err := fe.Run(rootDir)
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

	if reboot {
		if err := cmd.RunAndLog("reboot"); err != nil {
			fatal(err)
		}
	}
}
