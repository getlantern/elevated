// demo provides an example of using escalated. It includes one privileged
// function that sets the mtu on en0 to a random value between 1300 and 1500
// using the networksetup utility. This is something that requires root
// permissions.
package main

import (
	"os/exec"

	"bitbucket.org/kardianos/osext"
	"github.com/getlantern/escalated"
	"github.com/getlantern/golog"
)

var (
	log = golog.LoggerFor("demo")
)

func main() {
	escalated.Export(firewallAdd)
	err := escalated.Run(9789, doMain)
	if err != nil {
		panic(err)
	}
}

func doMain() error {
	log.Debug("Running program")

	err := firewallAdd("demo_bad")
	if err != nil {
		log.Debugf("First call got an error as expected: %v", err)
	} else {
		log.Debug("First call didn't get an error, though it should have!")
	}

	err = escalated.Call(firewallAdd, "demo_good")
	if err == nil {
		log.Debug("Successfully called escalated function")
	} else {
		log.Debugf("Unexpected error calling escalated function: %v", err)
	}

	return nil
}

// firewallAdd makes a call to the netsh utility
func firewallAdd(parms ...string) error {
	name := parms[0]
	exe, err := osext.Executable()
	if err != nil {
		return err
	}
	return exec.Command("netsh", "advfirewall", "firewall", "add", "rule", "name="+name, "dir=in", "action=allow", "program="+exe, "enable=yes", "profile=any").Run()
}
