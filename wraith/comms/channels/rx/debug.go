// +build debug

/*
This is a debug receiver for testing purposes. It only gets included in debug builds.
When included, it "receives" a print command every 2 seconds.
*/

package rx

import (
	"time"

	mm "git.0x1a8510f2.space/0x1a8510f2/wraith/modmgr"
)

func init() {
	var debug mm.CommsChanRxModule

	// Create a channel to trigger exit via the `Stop` method
	debug.Data["exitTrigger"] = make(chan struct{})
	// On start, run a thread pushing a debug message every 2 seconds
	debug.Start = func() {
		go func() {
			defer close(debug.Data["exitTrigger"].(chan struct{}))
			for {
				select {
				case <-debug.Data["exitTrigger"].(chan struct{}):
					return
				case <-time.After(2 * time.Second):
					debug.Data["queue"].(RxQueue) <- RxQueueElement{Data: []byte{}} /*RxQueueElement{Data: map[string]interface{}{
						"w.cmd": `func wcmd() string {println("Message from debug receiver"); return ""}`,
					}}*/
				}
			}
		}()
	}
	// On stop
	debug.Stop = func() {
		// Trigger exit
		debug.Data["exitTrigger"].(chan struct{}) <- struct{}{}
		// Wait until channel closed (exit confirmed)
		<-debug.Data["exitTrigger"].(chan struct{})
	}
	// Register handler for the debug:// URL scheme (which is never really used)
	mm.Modules.Register_CommsChanRxModule("debug", &debug, true)
}
