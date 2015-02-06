// +build !windows,!nacl,!plan9

package elevated

import (
	"fmt"
	"log/syslog"
	"os"
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

func verifyProgramSecurable() error {
	return nil
}

func ensureProgramSecure() error {
	log.Debugf("Changing ownership of %v to root", program)
	err := os.Chown(program, 0, 0)
	if err != nil {
		return fmt.Errorf("Unable to chown %v to root: %v", program, err)
	}

	log.Debugf("Chmodding %v to 0755", program)
	err = os.Chmod(program, 0755)
	if err != nil {
		return fmt.Errorf("Unable to chmod %v to 0755: %v", program, err)
	}

	return nil
}
