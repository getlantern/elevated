// package elevated supports exporting certain functions for execution with
// elevated privileges. It does this by installing the current program as a
// privileged system service (or Launchd Daemon on OS X). This service exposes a
// REST API that allows any process to execute the elevated functions by making
// an HTTP call to the service. Installing this system service requires
// prompting the user for elevation (e.g. UAC) once, but after that it can be
// used to execute elevated functions without prompting the user.
//
// Since any process on the local machine can call this REST API, it's best not
// to make the elevated functions too open-ended.
//
// TODO - on OS X, use smjobbless with XPC for even better security. See
// http://stackoverflow.com/questions/9134841/writing-a-privileged-helper-tool-with-smjobbless
// for discussion.
package elevated

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"bitbucket.org/kardianos/osext"
	"github.com/getlantern/golog"
	"github.com/getlantern/service"
	"github.com/getlantern/waitforserver"
)

const (
	flagInstall = "-install"
	flagelevate = "-elevate"
)

var (
	log = golog.LoggerFor("elevated")
)

var (
	program     string
	programFile string
	name        string
	allArgs     []string
	svcCfg      service.Config
	svc         service.Service
)

var (
	port             = 0
	addr             string
	exportedFns      = make(map[int]func(parms ...string) error)
	exportedFnsIndex = 0
	exportedFnsMutex sync.Mutex
)

// Export exports the given function as an elevated function.
func Export(fn func(parms ...string) error) {
	exportedFnsMutex.Lock()
	exportedFns[exportedFnsIndex] = fn
	exportedFnsIndex += 1
	exportedFnsMutex.Unlock()
}

// Run gives control to elevated. If this program was invoked as the elevated
// service, Run will start the REST server for handling invocations of
// elevated functions. Otherwise, it will make sure that the service is
// installed and then run the given main function.
func Run(exportPort int, main func() error) error {
	port = exportPort
	addr = fmt.Sprintf("localhost:%d", port)

	var err error
	program, err = osext.Executable()
	if err != nil {
		return fmt.Errorf("Unable to determine program: %v", err)
	}

	_, programFile = filepath.Split(program)
	name = programFile + ".elevated"
	allArgs = make([]string, 0, len(os.Args))
	for _, arg := range os.Args[1:] {
		if arg != flagInstall {
			allArgs = append(allArgs, arg)
		}
	}
	allArgs = append(allArgs, flagelevate)
	svcCfg = service.Config{
		Name:      name,
		Arguments: allArgs,
		Start: func() error {
			go runelevated()
			return nil
		},
	}

	svc, err = service.New(svcCfg)
	if err != nil {
		return fmt.Errorf("Unable to construct service: %v", err)
	}

	if hasFlag(flagelevate) {
		return svc.Run()
	} else if hasFlag(flagInstall) {
		return runInstall()
	} else {
		return runMain(main)
	}
}

func runMain(main func() error) error {
	log.Debug("Running main")

	err := verifyProgramSecurable()
	if err != nil {
		return fmt.Errorf("Program at %v cannot be secured for use as a service: %v", program, err)
	}

	err = waitforserver.WaitForServer("tcp", addr, 100*time.Millisecond)
	needsUpdate := err != nil

	if needsUpdate {
		log.Debug("Installing as a service")

		prompt := fmt.Sprintf("%v needs to install itself as a service", programFile)
		out, err := elevatedCommand(prompt, program, flagInstall).CombinedOutput()
		if err != nil {
			return fmt.Errorf("Unable to install service: %v (%v)", string(out), err)
		}
		log.Debug("Installed service")
	}

	err = waitforserver.WaitForServer("tcp", addr, 30*time.Second)
	if err != nil {
		return fmt.Errorf("elevated server not found")
	}

	return main()
}

func runInstall() error {
	log.Debug("Installing service")

	err := ensureProgramSecure()
	if err != nil {
		return fmt.Errorf("Program at %v not secure, can't install as service: %v", program, err)
	}

	updated, err := svc.InstallOrUpdate()
	if err != nil {
		return fmt.Errorf("Unable to install or update service: %v", err)
	}
	if updated {
		log.Debug("Stopping service")
		err = svc.Stop()
		if err != nil {
			return fmt.Errorf("Unable to stop service: %v", err)
		}
	}

	log.Debug("Making sure service is started")
	return svc.Start()
}

func runelevated() error {
	err := directLogsToSyslog()
	if err != nil {
		return err
	}

	log.Debugf("Running as elevated service at %v", addr)
	s := &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(handleelevatedCall),
	}
	return s.ListenAndServe()
}

func hasFlag(flag string) bool {
	for _, arg := range os.Args {
		if arg == flag {
			return true
		}
	}
	return false
}

func Call(fn func(parms ...string) error, parms ...string) error {
	for i, efn := range exportedFns {
		if fmt.Sprint(efn) == fmt.Sprint(fn) {
			c := &call{i, parms}
			data, err := json.Marshal(c)
			if err != nil {
				return fmt.Errorf("Unable to marshall call: %v", err)
			}
			resp, err := http.Post(fmt.Sprintf("http://%v", addr), "application/json", bytes.NewReader(data))
			if err != nil {
				return fmt.Errorf("Unable to send call: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				b, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return fmt.Errorf("Unable to read failed response body: %v", err)
				}
				return fmt.Errorf("Bad response status %d: %v", resp.StatusCode, string(b))
			}
			return nil
		}
	}

	return fmt.Errorf("Called function is not exported")
}

func handleelevatedCall(resp http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(resp, "Unable to read call data: %v", err)
	}

	c := &call{}
	err = json.Unmarshal(data, c)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(resp, "Unable to unmarshall call data: %v", err)
	}

	fn := exportedFns[c.Fnid]
	if fn == nil {
		resp.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(resp, "Unknown function id: %d", c.Fnid)
	}

	err = fn(c.Parms...)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(resp, "Error making call: %v", err)
	}

	resp.WriteHeader(http.StatusOK)
}

type call struct {
	Fnid  int
	Parms []string
}
