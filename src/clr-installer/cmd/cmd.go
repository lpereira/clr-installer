package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"clr-installer/log"
)

type runLogger struct{}

var (
	httpsProxy string
)

// SetHTTPSProxy defines the HTTPS_PROXY env var value for all the cmd executions
func SetHTTPSProxy(addr string) {
	httpsProxy = addr
}

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
func RunAndLog(args ...string) error {
	return Run(runLogger{}, args...)
}

// Run executes a command and uses writer to write both stdout and stderr
// args are the actual command and its arguments
func Run(writer io.Writer, args ...string) error {
	var exe string
	var cmdArgs []string

	log.Debug("%s", strings.Join(args, " "))

	exe = args[0]
	cmdArgs = args[1:]

	cmd := exec.Command(exe, cmdArgs...)

	if httpsProxy != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("https_proxy=%s", httpsProxy))
	}

	cmd.Stdout = writer
	cmd.Stderr = writer

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
