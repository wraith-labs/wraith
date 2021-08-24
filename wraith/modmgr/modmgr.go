// This package keeps track of all modular parts of Wraith. They must register
// with this module in order to be used. They can also be de-registered and
// re-registered dynamically.

// TODO: This could really benefit from generics once they land

package modmgr

import "sync"

//
type ModuleTree struct {
	lock           sync.Mutex
	modCommsChanTx [2]map[string]*CommsChanTxModule
	modCommsChanRx [2]map[string]*CommsChanRxModule
	modProtoLang   [2]map[string]*ProtoLangModule
	modProtoPart   [2]map[string]*ProtoPartModule
}

// Automatically lock and unlock the tree mutex
func (mt *ModuleTree) mutex() func() {
	mt.lock.Lock()
	return func() {
		mt.lock.Unlock()
	}
}

// Register a module so that it can be used by Wraith

func (mt *ModuleTree) Register_CommsChanTxModule(modname string, mod *CommsChanTxModule, enabled bool) {
	defer mt.mutex()()

	index := 0
	if !enabled {
		index = 1
	}
	mt.modCommsChanTx[index][modname] = mod
}

func (mt *ModuleTree) Register_CommsChanRxModule(modname string, mod *CommsChanRxModule, enabled bool) {
	defer mt.mutex()()

	index := 0
	if !enabled {
		index = 1
	}
	mt.modCommsChanRx[index][modname] = mod
}

func (mt *ModuleTree) Register_ProtoLangModule(modname string, mod *ProtoLangModule, enabled bool) {
	defer mt.mutex()()

	index := 0
	if !enabled {
		index = 1
	}
	mt.modProtoLang[index][modname] = mod
}

func (mt *ModuleTree) Register_ProtoPartModule(modname string, mod *ProtoPartModule, enabled bool) {
	defer mt.mutex()()

	index := 0
	if !enabled {
		index = 1
	}
	mt.modProtoPart[index][modname] = mod
}

// Semi-permanently (does not survive Wraith re-starts) remove a module
// This can save memory if a module is guaranteed to not be needed anymore, but is
// very risky because the module can never be re-added without re-starting Wraith
// (if the module is built-in) or re-sending the module (if it's not)

func (mt *ModuleTree) Deregister_CommsChanTxModule(modname string) {
	defer mt.mutex()()

	// Make sure to delete both if enabled and disabled
	// delete() is a no-op when the key does not exist so it's safe not to check
	delete(mt.modCommsChanTx[0], modname)
	delete(mt.modCommsChanTx[1], modname)
}

func (mt *ModuleTree) Deregister_CommsChanRxModule(modname string) {
	defer mt.mutex()()

	// Make sure to delete both if enabled and disabled
	// delete() is a no-op when the key does not exist so it's safe not to check
	delete(mt.modCommsChanRx[0], modname)
	delete(mt.modCommsChanRx[1], modname)
}

func (mt *ModuleTree) Deregister_ProtoLangModule(modname string) {
	defer mt.mutex()()

	// Make sure to delete both if enabled and disabled
	// delete() is a no-op when the key does not exist so it's safe not to check
	delete(mt.modProtoLang[0], modname)
	delete(mt.modProtoLang[1], modname)
}

func (mt *ModuleTree) Deregister_ProtoPartModule(modname string) {
	defer mt.mutex()()

	// Make sure to delete both if enabled and disabled
	// delete() is a no-op when the key does not exist so it's safe not to check
	delete(mt.modProtoPart[0], modname)
	delete(mt.modProtoPart[1], modname)
}

// If a module is currently registered but disabled, enable it

func (mt *ModuleTree) Enable_CommsChanTxModule(modname string) {
	defer mt.mutex()()

	if mod, exists := mt.modCommsChanTx[1][modname]; exists {
		mt.modCommsChanTx[0][modname] = mod
		delete(mt.modCommsChanTx[1], modname)
	}
}

func (mt *ModuleTree) Enable_CommsChanRxModule(modname string) {
	defer mt.mutex()()

	if mod, exists := mt.modCommsChanRx[1][modname]; exists {
		mt.modCommsChanRx[0][modname] = mod
		delete(mt.modCommsChanRx[1], modname)
	}
}

func (mt *ModuleTree) Enable_ProtoLangModule(modname string) {
	defer mt.mutex()()

	if mod, exists := mt.modProtoLang[1][modname]; exists {
		mt.modProtoLang[0][modname] = mod
		delete(mt.modProtoLang[1], modname)
	}
}

func (mt *ModuleTree) Enable_ProtoPartModule(modname string) {
	defer mt.mutex()()

	if mod, exists := mt.modProtoPart[1][modname]; exists {
		mt.modProtoPart[0][modname] = mod
		delete(mt.modProtoPart[1], modname)
	}
}

// If a module is currently registered and enabled, disable it

func (mt *ModuleTree) Disable_CommsChanTxModule(modname string) {
	defer mt.mutex()()

	if mod, exists := mt.modCommsChanTx[0][modname]; exists {
		mt.modCommsChanTx[1][modname] = mod
		delete(mt.modCommsChanTx[0], modname)
	}
}

func (mt *ModuleTree) Disable_CommsChanRxModule(modname string) {
	defer mt.mutex()()

	if mod, exists := mt.modCommsChanRx[0][modname]; exists {
		mt.modCommsChanRx[1][modname] = mod
		delete(mt.modCommsChanRx[0], modname)
	}
}

func (mt *ModuleTree) Disable_ProtoLangModule(modname string) {
	defer mt.mutex()()

	if mod, exists := mt.modProtoLang[0][modname]; exists {
		mt.modProtoLang[1][modname] = mod
		delete(mt.modProtoLang[0], modname)
	}
}

func (mt *ModuleTree) Disable_ProtoPartModule(modname string) {
	defer mt.mutex()()

	if mod, exists := mt.modProtoPart[0][modname]; exists {
		mt.modProtoPart[1][modname] = mod
		delete(mt.modProtoPart[0], modname)
	}
}

// Get all enabled modules of a certain type as an array (nameless)
// This is less CPU-efficient than GetEnabledNamed but more memory-efficient

func (mt *ModuleTree) GetEnabledNameless_CommsChanTxModule() []*CommsChanTxModule {
	defer mt.mutex()()

	mods := []*CommsChanTxModule{}
	for _, mod := range mt.modCommsChanTx[0] {
		mods = append(mods, mod)
	}
	return mods
}

func (mt *ModuleTree) GetEnabledNameless_CommsChanRxModule() []*CommsChanRxModule {
	defer mt.mutex()()

	mods := []*CommsChanRxModule{}
	for _, mod := range mt.modCommsChanRx[0] {
		mods = append(mods, mod)
	}
	return mods
}

func (mt *ModuleTree) GetEnabledNameless_ProtoLangModule() []*ProtoLangModule {
	defer mt.mutex()()

	mods := []*ProtoLangModule{}
	for _, mod := range mt.modProtoLang[0] {
		mods = append(mods, mod)
	}
	return mods
}

func (mt *ModuleTree) GetEnabledNameless_ProtoPartModule() []*ProtoPartModule {
	defer mt.mutex()()

	mods := []*ProtoPartModule{}
	for _, mod := range mt.modProtoPart[0] {
		mods = append(mods, mod)
	}
	return mods
}

// Get the names of all enabled modules of a certain type as an array
// This is less CPU-efficient than GetEnabledNamed but more memory-efficient

func (mt *ModuleTree) GetEnabledNames_CommsChanTxModule() []string {
	defer mt.mutex()()

	modNames := []string{}
	for modName := range mt.modCommsChanTx[0] {
		modNames = append(modNames, modName)
	}
	return modNames
}

func (mt *ModuleTree) GetEnabledNames_CommsChanRxModule() []string {
	defer mt.mutex()()

	modNames := []string{}
	for modName := range mt.modCommsChanRx[0] {
		modNames = append(modNames, modName)
	}
	return modNames
}

func (mt *ModuleTree) GetEnabledNames_ProtoLangModule() []string {
	defer mt.mutex()()

	modNames := []string{}
	for modName := range mt.modProtoLang[0] {
		modNames = append(modNames, modName)
	}
	return modNames
}

func (mt *ModuleTree) GetEnabledNames_ProtoPartModule() []string {
	defer mt.mutex()()

	modNames := []string{}
	for modName := range mt.modProtoPart[0] {
		modNames = append(modNames, modName)
	}
	return modNames
}

// Get all enabled modules of a certain type as a map (named)

func (mt *ModuleTree) GetEnabledNamed_CommsChanTxModule() map[string]*CommsChanTxModule {
	defer mt.mutex()()

	return mt.modCommsChanTx[0]
}

func (mt *ModuleTree) GetEnabledNamed_CommsChanRxModule() map[string]*CommsChanRxModule {
	defer mt.mutex()()

	return mt.modCommsChanRx[0]
}

func (mt *ModuleTree) GetEnabledNamed_ProtoLangModule() map[string]*ProtoLangModule {
	defer mt.mutex()()

	return mt.modProtoLang[0]
}

func (mt *ModuleTree) GetEnabledNamed_ProtoPartModule() map[string]*ProtoPartModule {
	defer mt.mutex()()

	return mt.modProtoPart[0]
}

// Get all disabled modules of a certain type as an array (nameless)
// This is less CPU-efficient than GetDisabledNamed but more memory-efficient

func (mt *ModuleTree) GetDisabledNameless_CommsChanTxModule() []*CommsChanTxModule {
	defer mt.mutex()()

	mods := []*CommsChanTxModule{}
	for _, mod := range mt.modCommsChanTx[1] {
		mods = append(mods, mod)
	}
	return mods
}

func (mt *ModuleTree) GetDisabledNameless_CommsChanRxModule() []*CommsChanRxModule {
	defer mt.mutex()()

	mods := []*CommsChanRxModule{}
	for _, mod := range mt.modCommsChanRx[1] {
		mods = append(mods, mod)
	}
	return mods
}

func (mt *ModuleTree) GetDisabledNameless_ProtoLangModule() []*ProtoLangModule {
	defer mt.mutex()()

	mods := []*ProtoLangModule{}
	for _, mod := range mt.modProtoLang[1] {
		mods = append(mods, mod)
	}
	return mods
}

func (mt *ModuleTree) GetDisabledNameless_ProtoPartModule() []*ProtoPartModule {
	defer mt.mutex()()

	mods := []*ProtoPartModule{}
	for _, mod := range mt.modProtoPart[1] {
		mods = append(mods, mod)
	}
	return mods
}

// Get the names of all disabled modules of a certain type as an array
// This is less CPU-efficient than GetDisabledNamed but more memory-efficient

func (mt *ModuleTree) GetDisabledNames_CommsChanTxModule() []string {
	defer mt.mutex()()

	modNames := []string{}
	for modName := range mt.modCommsChanTx[1] {
		modNames = append(modNames, modName)
	}
	return modNames
}

func (mt *ModuleTree) GetDisabledNames_CommsChanRxModule() []string {
	defer mt.mutex()()

	modNames := []string{}
	for modName := range mt.modCommsChanRx[1] {
		modNames = append(modNames, modName)
	}
	return modNames
}

func (mt *ModuleTree) GetDisabledNames_ProtoLangModule() []string {
	defer mt.mutex()()

	modNames := []string{}
	for modName := range mt.modProtoLang[1] {
		modNames = append(modNames, modName)
	}
	return modNames
}

func (mt *ModuleTree) GetDisabledNames_ProtoPartModule() []string {
	defer mt.mutex()()

	modNames := []string{}
	for modName := range mt.modProtoPart[1] {
		modNames = append(modNames, modName)
	}
	return modNames
}

// Get all disabled modules of a certain type as a map (named)

func (mt *ModuleTree) GetDisabledNamed_CommsChanTxModule(modtype int) map[string]*CommsChanTxModule {
	defer mt.mutex()()

	return mt.modCommsChanTx[1]
}

func (mt *ModuleTree) GetDisabledNamed_CommsChanRxModule(modtype int) map[string]*CommsChanRxModule {
	defer mt.mutex()()

	return mt.modCommsChanRx[1]
}

func (mt *ModuleTree) GetDisabledNamed_ProtoLangModule(modtype int) map[string]*ProtoLangModule {
	defer mt.mutex()()

	return mt.modProtoLang[1]
}

func (mt *ModuleTree) GetDisabledNamed_ProtoPartModule(modtype int) map[string]*ProtoPartModule {
	defer mt.mutex()()

	return mt.modProtoPart[1]
}

// Prepare ModuleTree stuff
func (mt *ModuleTree) Init() {
	defer mt.mutex()()

	// Init storages
	mt.modCommsChanTx = [2]map[string]*CommsChanTxModule{}
	mt.modCommsChanRx = [2]map[string]*CommsChanRxModule{}
	mt.modProtoLang = [2]map[string]*ProtoLangModule{}
}

// Global module tree
var Modules ModuleTree
