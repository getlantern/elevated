// demo provides an example of using elevated. It includes one privileged
// function that sets the mtu on en0 to a random value between 1300 and 1500
// using the networksetup utility. This is something that requires root
// permissions.
package main

import (
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/getlantern/elevated"
	"github.com/getlantern/golog"
)

var (
	log = golog.LoggerFor("demo")
)

func main() {
	rand.Seed(time.Now().UnixNano())
	elevated.Export(setMtu)
	err := elevated.Run(9789, doMain)
	if err != nil {
		panic(err)
	}
}

func doMain() error {
	log.Debug("Running program")

	err := setMtu("en0", "1400")
	if err != nil {
		log.Debugf("First call got an error as expected: %v", err)
	} else {
		log.Debug("First call didn't get an error, though it should have!")
	}

	newMtu := strconv.Itoa(1300 + rand.Intn(200))
	err = elevated.Call(setMtu, "en0", newMtu)
	if err == nil {
		log.Debug("Successfully called elevated function")
	} else {
		log.Debugf("Unexpected error calling elevated function: %v", err)
	}

	if mtuMatches("en0", newMtu) {
		log.Debugf("mtu was successfully updated to %v", newMtu)
	} else {
		log.Debug("mtu wasn't updated!")
	}

	err = elevated.Call(setMtu, "en0", "1500")
	if err == nil {
		log.Debug("mtu was successfully set back to 1500")
	} else {
		log.Debugf("Unable to set mtu back to 1500: %v", err)
	}

	return nil
}

// setMTU makes a call to the networksetup utility that requires root
// permissions
func setMtu(parms ...string) error {
	intf := parms[0]
	mtu := parms[1]
	log.Debugf("Setting MTU for %v to %v", intf, mtu)
	cmd := exec.Command("networksetup", "-setMTU", intf, mtu)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: 0,
			Gid: 0,
		},
	}
	return cmd.Run()
}

// mtuMatches checks that the current mtu matches the expected value. This does
// not require root permissions.
func mtuMatches(intf string, mtu string) bool {
	out, err := exec.Command("networksetup", "-getMTU", intf).CombinedOutput()
	if err != nil {
		log.Errorf("Unable to get MTU: %v", err)
		return false
	}
	return strings.Contains(string(out), mtu)
}
