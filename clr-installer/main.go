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

	"github.com/clearlinux/clr-installer/cmd"
	"github.com/clearlinux/clr-installer/conf"
	"github.com/clearlinux/clr-installer/crypt"
	"github.com/clearlinux/clr-installer/frontend"
	"github.com/clearlinux/clr-installer/keyboard"
	"github.com/clearlinux/clr-installer/log"
	"github.com/clearlinux/clr-installer/massinstall"
	"github.com/clearlinux/clr-installer/model"
	"github.com/clearlinux/clr-installer/tui"
	flag "github.com/spf13/pflag"
)

var (
	frontEndImpls []frontend.Frontend
	args          frontend.Args
	genPwd        string
)

func init() {
	flag.BoolVar(
		&args.Version, "version", false, "Version of the Installer",
	)

	flag.BoolVar(
		&args.Reboot, "reboot", true, "Reboot after finishing",
	)

	flag.BoolVar(
		&args.ForceTUI, "tui", false, "Use TUI frontend",
	)

	flag.StringVar(
		&args.ConfigFile, "config", "", "Installation configuration file",
	)

	flag.StringVar(
		&genPwd, "genpwd", "", "Generates a PAM compatible password hash based on the provided string",
	)

	flag.IntVar(
		&args.LogLevel,
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
	flag.StringVar(&args.LogFile, "log-file", defaultLogFile, "The log file path")
}

func fatal(err error) {
	log.ErrorError(err)
	panic(err)
}

func initFrontendList() {
	frontEndImpls = []frontend.Frontend{
		massinstall.New(),
		tui.New(),
	}
}

func main() {
	flag.Parse()

	if err := log.SetLogLevel(args.LogLevel); err != nil {
		fatal(err)
	}

	if genPwd != "" {
		hashed, err := crypt.Crypt(genPwd)
		if err != nil {
			panic(err)
		}

		fmt.Println(hashed)
		return
	}

	if args.Version {
		fmt.Println(path.Base(os.Args[0]) + ": " + model.Version)
		return
	}

	f, err := os.OpenFile(args.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
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
	cf := args.ConfigFile

	if args.ConfigFile == "" {
		if cf, err = conf.LookupDefaultConfig(); err != nil {
			fatal(err)
		}
	}

	log.Debug("Loading config file: %s", cf)
	if md, err = model.LoadFile(cf); err != nil {
		fatal(err)
	}

	if md.Keyboard != nil {
		if err := keyboard.Apply(md.Keyboard); err != nil {
			fatal(err)
		}
	}

	installReboot := false

	go func() {
		for _, fe := range frontEndImpls {
			if !fe.MustRun(&args) {
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

	if args.Reboot && installReboot {
		if err := cmd.RunAndLog("reboot"); err != nil {
			fatal(err)
		}
	}
}
