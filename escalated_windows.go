package escalated

import (
	"os/exec"

	"github.com/getlantern/escalate"
)

func escalatedCommand(prompt string, program string, args ...string) *exec.Cmd {
	return escalate.Prompt(prompt, program, args...)
}

func directLogsToSyslog() error {
	return nil
}
