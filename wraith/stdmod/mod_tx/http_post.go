package mod_tx

import (
	"git.0x1a8510f2.space/0x1a8510f2/wraith/libwraith"
)

type HttpPostModule struct {
	w *libwraith.Wraith
}

func (m *HttpPostModule) WraithModuleInit(wraith *libwraith.Wraith) {
	m.w = wraith
}
func (m *HttpPostModule) CommsChanTxModule() {}

// On start, no-op because this module only needs to do stuff when triggered
func (m *HttpPostModule) StartTx() {}

// On stop, do nothing because this module doesn't run anything in the background
func (m *HttpPostModule) StopTx() {}

// On trigger, send HTTP POST request to server, with attached data
func (m *HttpPostModule) TriggerTx(addr string, data []byte) bool {

	return true
}
