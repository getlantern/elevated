package elevated

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/getlantern/elevate"
)

var (
	isInProgramfiles = regexp.MustCompile(`[a-zA-Z]:\\Program Files( \(x86\))?.*`)
)

func elevatedCommand(prompt string, program string, args ...string) *exec.Cmd {
	return elevate.Prompt(prompt, program, args...)
}

func directLogsToSyslog() error {
	return nil
}

func verifyProgramSecurable() error {
	dir, _ := filepath.Split(program)
	if !isInProgramfiles.MatchString(dir) {
		return fmt.Errorf("Program %v is not in Program Files", program)
	}

	return nil
}

func ensureProgramSecure() error {
	return verifyProgramSecurable()
}
