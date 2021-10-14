//go:build debug
// +build debug

package mod_tx

import (
	"fmt"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/libwraith"
)

type DebugModule struct {
	w    *libwraith.Wraith
	data map[string]interface{}
}

func (m *DebugModule) WraithModuleInit(wraith *libwraith.Wraith) {
	fmt.Printf("DEBUG: mod_tx.DebugModule.WraithModuleInit called\n")

	m.w = wraith
}
func (m *DebugModule) CommsChanTxModule() {}

// On start, no-op
func (m *DebugModule) StartTx() {
	fmt.Printf("DEBUG: mod_tx.DebugModule.StartRx called\n")
}

func (m *DebugModule) StopTx() {
	fmt.Printf("DEBUG: mod_tx.DebugModule.StopRx called\n")
}

func (m *DebugModule) TriggerTx(addr string, data []byte) bool {
	fmt.Printf("DEBUG: mod_tx.DebugModule.TriggerTx called with params: %v | %v\n", addr, data)

	return true
}
