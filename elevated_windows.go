package elevated

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/getlantern/elevate"
)

const (
	mimicACLSOf = `C:\Program Files`
)

var (
	aclSplit = regexp.MustCompile(`([^_]+) "([^"]+)"`)
)

func elevatedCommand(prompt string, program string, args ...string) *exec.Cmd {
	return elevate.Prompt(prompt, program, args...)
}

func directLogsToSyslog() error {
	return nil
}

func ensureProgramSecure() error {
	log.Tracef(`Get ACL SDDL for %v`, mimicACLSOf)
	out, err := exec.Command("cacls", mimicACLSOf, "/S").CombinedOutput()
	if err != nil {
		return fmt.Errorf(`Unable to get ACLS for %v: %v: %v`, mimicACLSOf, out, err)
	}

	ms := aclSplit.FindStringSubmatch(string(out))
	if len(ms) != 3 {
		return fmt.Errorf("Unable to parse SDDL from %v", string(out))
	}
	sddl := ms[2]

	log.Tracef(`Set ACL of program to match ACL of %v`, mimicACLSOf)
	cmd := exec.Command("cacls", program, fmt.Sprintf("/S:%v", sddl))
	// Respond "Y" to "Are you sure?" prompt
	cmd.Stdin = bytes.NewReader([]byte("Y"))
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(`Unable to set ACLS for %v: %v: %v`, program, out, err)
	}

	return nil
}
