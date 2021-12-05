package main

import (
	"os"
	"os/signal"
	"syscall"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/libwraith"
)

const RESPECT_EXIT_SIGNALS = true

var exitTrigger chan struct{}

func setupCloseHandler(triggerChannel chan struct{}) {
	c := make(chan os.Signal, 1)
	signal.Notify(
		c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	go func() {
		<-c
		if RESPECT_EXIT_SIGNALS {
			close(triggerChannel)
		}
	}()
}

func init() {
	exitTrigger = make(chan struct{})
	setupCloseHandler(exitTrigger)
}

// Init module list as global variable so it can be accessed by debug.go
var modules = []libwraith.WraithModule{}

func main() {
	// Create Wraith
	w := libwraith.Wraith{}

	// Start Wraith in goroutine
	go w.Spawn(
		libwraith.WraithConf{},
		modules...,
	)

	// Wait until Wraith dies or the exit trigger fires
	for {
		select {
		case <-exitTrigger:
			w.Kill()
		case <-w.IsDead:
			return
		}
	}
}
