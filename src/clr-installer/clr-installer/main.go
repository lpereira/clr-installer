package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"

	"clr-installer/controller"
	"clr-installer/log"
	"clr-installer/model"

	flag "github.com/spf13/pflag"
)

var (
	logFile    string
	configFile string
	logLevel   int
)

func init() {
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
	os.Exit(1)
}

func main() {
	flag.Parse()

	if err := log.SetLogLevel(logLevel); err != nil {
		fatal(err)
	}

	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	rootDir, err := ioutil.TempDir("", "install-")
	if err != nil {
		fatal(err)
	}

	go func() {
		if configFile != "" {
			log.Debug("Loading config file: %s", configFile)
			model, err := model.LoadFile(configFile)
			if err != nil {
				fatal(err)
			}

			log.Debug("Starting install")
			err = controller.Install(rootDir, model)
			if err != nil {
				if controller.Cleanup(rootDir) != nil {
					log.ErrorError(err)
				}
				fatal(err)
			}
		}

		done <- true
	}()

	go func() {
		<-sigs
		fmt.Println("Leaving...")
		done <- true
	}()

	<-done
	if err := controller.Cleanup(rootDir); err != nil {
		log.ErrorError(err)
	}
}
