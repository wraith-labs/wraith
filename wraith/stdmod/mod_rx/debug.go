/*
This is a debug receiver for testing purposes. It only gets included in debug builds.
When included, it "receives" a print command every 2 seconds.
*/

package mod_rx

import (
	"encoding/json"
	"time"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/libwraith"
)

type debugReceiveChannel struct {
	data map[string]interface{}
}

func (c debugReceiveChannel) WraithModule()      {}
func (c debugReceiveChannel) CommsChanRxModule() {}

// On start, run a thread pushing a debug message every 2 seconds
func (c debugReceiveChannel) StartRx() {
	// Init data map
	c.data = make(map[string]interface{})

	// Create a channel to trigger exit via the `StopRx` method
	c.data["exitTrigger"] = make(chan struct{})

	// The data to send over debug
	debugData := map[string]interface{}{
		"w.cmd": `func wcmd() string {println("Message from debug receiver"); return ""}`,
	}
	debugDataJson, err := json.Marshal(debugData)
	if err != nil {
		panic("Marshalling debug data failed!")
	}

	go func() {
		defer close(c.data["exitTrigger"].(chan struct{}))
		for {
			select {
			case <-c.data["exitTrigger"].(chan struct{}):
				return
			case <-time.After(2 * time.Second):
				c.data["queue"].(libwraith.RxQueue) <- libwraith.RxQueueElement{Data: debugDataJson}
			}
		}
	}()
}

func (c debugReceiveChannel) StopRx() {
	// Trigger exit
	c.data["exitTrigger"].(chan struct{}) <- struct{}{}
	// Wait until channel closed (exit confirmed)
	<-c.data["exitTrigger"].(chan struct{})
}
