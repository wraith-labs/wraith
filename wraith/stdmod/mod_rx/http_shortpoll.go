package mod_rx

import (
	"encoding/json"
	"fmt"
	"time"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/libwraith"
)

type HttpShortpollModule struct {
	w    *libwraith.Wraith
	data map[string]interface{}
}

func (m *HttpShortpollModule) WraithModuleInit(wraith *libwraith.Wraith) {
	m.w = wraith
}
func (m *HttpShortpollModule) CommsChanRxModule() {}

// On start, run a thread pushing a debug message every 2 seconds
func (m *HttpShortpollModule) StartRx() {
	fmt.Printf("DEBUG: mod_rx.DebugModule.StartRx called\n")

	// Init data map
	m.data = make(map[string]interface{})

	// Create a channel to trigger exit via the `StopRx` method
	m.data["exitTrigger"] = make(chan struct{})

	// The data to send over debug
	debugData := map[string]interface{}{
		"cmd":   `func f() string {println("Message from debug receiver"); return ""}`,
		"debug": "DEBUG!",
	}
	debugDataJson, err := json.Marshal(debugData)
	if err != nil {
		panic("Marshalling debug data failed!")
	}

	go func() {
		defer close(m.data["exitTrigger"].(chan struct{}))
		for {
			select {
			case <-m.data["exitTrigger"].(chan struct{}):
				return
			case <-time.After(2 * time.Second):
				m.w.PushRx(libwraith.RxQueueElement{Data: append([]byte("DEBUG:"), debugDataJson...)})
			}
		}
	}()
}

func (m *HttpShortpollModule) StopRx() {
	fmt.Printf("DEBUG: mod_rx.DebugModule.StopRx called\n")

	// Trigger exit
	m.data["exitTrigger"].(chan struct{}) <- struct{}{}
	// Wait until channel closed (exit confirmed)
	<-m.data["exitTrigger"].(chan struct{})
}
