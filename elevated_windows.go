package elevated

import (
	"os/exec"

	"github.com/getlantern/elevate"
)

func elevatedCommand(prompt string, program string, args ...string) *exec.Cmd {
	return elevate.Prompt(prompt, program, args...)
}

func directLogsToSyslog() error {
	return nil
}
