package cmd

import (
	"io"
	"os/exec"
	"strings"

	"clr-installer/log"
)

type runLogger struct{}

func (rl runLogger) Write(p []byte) (n int, err error) {
	for _, curr := range strings.Split(string(p), "\n") {
		if curr == "" {
			continue
		}

		log.Out(curr)
	}
	return len(p), nil
}

// RunAndLog executes a command (similar to Run) but takes care of writing
// the output to default logger
func RunAndLog(root bool, args ...string) error {
	return Run(runLogger{}, root, args...)
}

// Run executes a command and uses writer to write both stdout and stderr
// root determines if the command must be executed as root (using sudo) and
// args are the actual command and its arguments
func Run(writer io.Writer, root bool, args ...string) error {
	var exe string
	var cmdArgs []string

	log.Debug("%s", strings.Join(args, " "))

	if root {
		exe = "sudo"
		cmdArgs = args
	} else {
		exe = args[0]
		cmdArgs = args[1:]
	}

	cmd := exec.Command(exe, cmdArgs...)

	cmd.Stdout = writer
	cmd.Stderr = writer

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
