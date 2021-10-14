//go:build debug
// +build debug

package main

// This file activates the debug modules if the debug tag is set while building.
// Otherwise it does nothing and is not included.

import (
	"git.0x1a8510f2.space/0x1a8510f2/wraith/libwraith"
	"git.0x1a8510f2.space/0x1a8510f2/wraith/stdmod/mod_lang"
	"git.0x1a8510f2.space/0x1a8510f2/wraith/stdmod/mod_part"
	"git.0x1a8510f2.space/0x1a8510f2/wraith/stdmod/mod_rx"
	"git.0x1a8510f2.space/0x1a8510f2/wraith/stdmod/mod_tx"
)

func init() {
	// Set up debugging modules if needed
	w.Modules.Register("w.debug", libwraith.ModCommsChanRx, &mod_rx.DebugModule{}, true)
	w.Modules.Register("w.debug", libwraith.ModProtoLang, &mod_lang.DebugModule{}, true)
	w.Modules.Register("w.debug", libwraith.ModProtoPart, &mod_part.DebugModule{}, true)
	w.Modules.Register("w.debug", libwraith.ModCommsChanTx, &mod_tx.DebugModule{}, true)
}
