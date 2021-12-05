//go:build debug
// +build debug

package main

// This file activates the debug modules if the debug tag is set while building.
// Otherwise it does nothing and is not included.

func init() {
	// Set up debugging modules if needed
	//w.Modules.Register("debug", libwraith.ModCommsChanRx, &mod_rx.DebugModule{}, true)
	//w.Modules.Register("debug", libwraith.ModProtoLang, &mod_lang.DebugModule{}, true)
	//w.Modules.Register("debug", libwraith.ModProtoPart, &mod_part.DebugModule{}, true)
	//w.Modules.Register("debug", libwraith.ModCommsChanTx, &mod_tx.DebugModule{}, true)
}
