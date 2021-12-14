package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/libwraith"
	"git.0x1a8510f2.space/0x1a8510f2/wraith/stdmod"
)

const RESPECT_EXIT_SIGNALS = true

var exitTrigger chan struct{}

func setupCloseHandler(triggerChannel chan struct{}) {
	c := make(chan os.Signal, 2)
	signal.Notify(
		c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	if RESPECT_EXIT_SIGNALS {
		go func() {
			for range c {
				triggerChannel <- struct{}{}
			}
		}()
	}
}

func init() {
	exitTrigger = make(chan struct{})
	setupCloseHandler(exitTrigger)
}

func main() {
	// Create Wraith
	w := libwraith.Wraith{}

	// Create a channel to watch for Wraith status updates
	statusUpdates, _ := w.SHMWatch(libwraith.SHM_WRAITH_STATUS)

	// Start Wraith in goroutine
	go w.Spawn(
		libwraith.WraithConf{
			FingerprintGenerator: func() string { return "" },
		},
		&stdmod.DefaultJWTCommsManager{},
	)

	// Wait until Wraith starts up or time out
waitloop:
	for {
		select {
		case status := <-statusUpdates:
			if status == libwraith.WSTATUS_ACTIVE {
				break waitloop
			}
		case <-time.After(2 * time.Second):
			panic("Wraith failed to start within 2 seconds")
		}
	}

	// Wait until Wraith dies or the exit trigger fires
	for {
		select {
		case <-exitTrigger:
			// TODO: Check success
			w.Kill(30 * time.Second)
		case status := <-statusUpdates:
			if status == libwraith.WSTATUS_INACTIVE {
				return
			} else if status == libwraith.WSTATUS_ERROR {
				panic("Wraith exited with error status")
			}
		}
	}
}
