// +build !windows,!nacl,!plan9

package elevated

import (
	"fmt"
	"log/syslog"
	"os/exec"

	"github.com/getlantern/elevate"
	"github.com/getlantern/golog"
)

func elevatedCommand(prompt string, program string, args ...string) *exec.Cmd {
	return elevate.Prompt(prompt, program, args...)
}

func directLogsToSyslog() error {
	errorOut, err := syslog.New(syslog.LOG_ERR, name)
	if err != nil {
		return fmt.Errorf("Unable to get syslog for errors: %v", err)
	}
	defer errorOut.Close()

	debugOut, err := syslog.New(syslog.LOG_INFO, name)
	if err != nil {
		return fmt.Errorf("Unable to get syslog for debug: %v", err)
	}
	defer debugOut.Close()

	debugOut.Write([]byte("Updating logs"))
	golog.SetOutputs(errorOut, debugOut)
	debugOut.Write([]byte("Updated logs"))

	return nil
}
