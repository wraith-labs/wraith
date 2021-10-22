package libwraith

import "sync"

// A structure holding lists of modules
type ModuleTree struct {
	modCommsChanTx [2]map[string]CommsChanTxModule
	modCommsChanRx [2]map[string]CommsChanRxModule
	modProtoLang   [2]map[string]ProtoLangModule
	modProtoPart   [2]map[string]ProtoPartModule
	lock           sync.Mutex
}

// Automatically lock and unlock the tree mutex
func (mt *ModuleTree) wrapper() func() {
	// Lock struct to make it thread-safe
	mt.lock.Lock()

	// Make sure all maps are initialised
	if mt.modCommsChanTx[0] == nil {
		mt.modCommsChanTx[0] = map[string]CommsChanTxModule{}
	}
	if mt.modCommsChanTx[1] == nil {
		mt.modCommsChanTx[1] = map[string]CommsChanTxModule{}
	}
	if mt.modCommsChanRx[0] == nil {
		mt.modCommsChanRx[0] = map[string]CommsChanRxModule{}
	}
	if mt.modCommsChanRx[1] == nil {
		mt.modCommsChanRx[1] = map[string]CommsChanRxModule{}
	}
	if mt.modProtoLang[0] == nil {
		mt.modProtoLang[0] = map[string]ProtoLangModule{}
	}
	if mt.modProtoLang[1] == nil {
		mt.modProtoLang[1] = map[string]ProtoLangModule{}
	}
	if mt.modProtoPart[0] == nil {
		mt.modProtoPart[0] = map[string]ProtoPartModule{}
	}
	if mt.modProtoPart[1] == nil {
		mt.modProtoPart[1] = map[string]ProtoPartModule{}
	}

	// Make sure to unlock the struct at the end
	return func() {
		mt.lock.Unlock()
	}
}

// Register a module so that it can be used by Wraith (can be registered under multiple names)
func (mt *ModuleTree) Register(mtype modtype, mod GenericModule, enabled bool, mnames ...string) {
	defer mt.wrapper()()

	index := 0
	if !enabled {
		index = 1
	}

	switch mtype {
	case ModCommsChanTx:
		if commsChanTxMod, ok := mod.(CommsChanTxModule); ok {
			for _, mname := range mnames {
				mt.modCommsChanTx[index][mname] = commsChanTxMod
			}
		}
	case ModCommsChanRx:
		if commsChanRxMod, ok := mod.(CommsChanRxModule); ok {
			for _, mname := range mnames {
				mt.modCommsChanRx[index][mname] = commsChanRxMod
			}
		}
	case ModProtoLang:
		if protoLangMod, ok := mod.(ProtoLangModule); ok {
			for _, mname := range mnames {
				mt.modProtoLang[index][mname] = protoLangMod
			}
		}
	case ModProtoPart:
		if protoPartMod, ok := mod.(ProtoPartModule); ok {
			for _, mname := range mnames {
				mt.modProtoPart[index][mname] = protoPartMod
			}
		}
	}
}

// Semi-permanently (does not survive Wraith re-starts) remove given modules
// This can save memory if a module is guaranteed to not be needed anymore, but is
// very risky because the module can never be re-added without re-starting Wraith
// (if the module is built-in) or re-sending the module (if it's not)
func (mt *ModuleTree) Deregister(mtype modtype, mnames ...string) {
	defer mt.wrapper()()

	// Make sure to delete both if enabled and disabled
	// delete() is a no-op when the key does not exist so it's safe not to check
	switch mtype {
	case ModCommsChanTx:
		for _, mname := range mnames {
			delete(mt.modCommsChanTx[0], mname)
			delete(mt.modCommsChanTx[1], mname)
		}
	case ModCommsChanRx:
		for _, mname := range mnames {
			delete(mt.modCommsChanRx[0], mname)
			delete(mt.modCommsChanRx[1], mname)
		}
	case ModProtoLang:
		for _, mname := range mnames {
			delete(mt.modProtoLang[0], mname)
			delete(mt.modProtoLang[1], mname)
		}
	case ModProtoPart:
		for _, mname := range mnames {
			delete(mt.modProtoPart[0], mname)
			delete(mt.modProtoPart[1], mname)
		}
	}
}

// If given modules are currently registered but disabled, enable them
func (mt *ModuleTree) Enable(mtype modtype, mnames ...string) {
	defer mt.wrapper()()

	switch mtype {
	case ModCommsChanTx:
		for _, mname := range mnames {
			if mod, exists := mt.modCommsChanTx[1][mname]; exists {
				mt.modCommsChanTx[0][mname] = mod
				delete(mt.modCommsChanTx[1], mname)
			}
		}
	case ModCommsChanRx:
		for _, mname := range mnames {
			if mod, exists := mt.modCommsChanRx[1][mname]; exists {
				mt.modCommsChanRx[0][mname] = mod
				delete(mt.modCommsChanRx[1], mname)
			}
		}
	case ModProtoLang:
		for _, mname := range mnames {
			if mod, exists := mt.modProtoLang[1][mname]; exists {
				mt.modProtoLang[0][mname] = mod
				delete(mt.modProtoLang[1], mname)
			}
		}
	case ModProtoPart:
		for _, mname := range mnames {
			if mod, exists := mt.modProtoPart[1][mname]; exists {
				mt.modProtoPart[0][mname] = mod
				delete(mt.modProtoPart[1], mname)
			}
		}
	}
}

// If given modules are currently registered and enabled, disable them
func (mt *ModuleTree) Disable(mtype modtype, mnames ...string) {
	defer mt.wrapper()()

	switch mtype {
	case ModCommsChanTx:
		for _, mname := range mnames {
			if mod, exists := mt.modCommsChanTx[0][mname]; exists {
				mt.modCommsChanTx[1][mname] = mod
				delete(mt.modCommsChanTx[0], mname)
			}
		}
	case ModCommsChanRx:
		for _, mname := range mnames {
			if mod, exists := mt.modCommsChanRx[0][mname]; exists {
				mt.modCommsChanRx[1][mname] = mod
				delete(mt.modCommsChanRx[0], mname)
			}
		}
	case ModProtoLang:
		for _, mname := range mnames {
			if mod, exists := mt.modProtoLang[0][mname]; exists {
				mt.modProtoLang[1][mname] = mod
				delete(mt.modProtoLang[0], mname)
			}
		}
	case ModProtoPart:
		for _, mname := range mnames {
			if mod, exists := mt.modProtoPart[0][mname]; exists {
				mt.modProtoPart[1][mname] = mod
				delete(mt.modProtoPart[0], mname)
			}
		}
	}
}

// Get all enabled modules of a certain type as a map (named)
func (mt *ModuleTree) GetEnabled(mtype modtype) interface{} {
	defer mt.wrapper()()

	switch mtype {
	case ModCommsChanTx:
		return mt.modCommsChanTx[0]
	case ModCommsChanRx:
		return mt.modCommsChanRx[0]
	case ModProtoLang:
		return mt.modProtoLang[0]
	case ModProtoPart:
		return mt.modProtoPart[0]
	default:
		return map[string]GenericModule{}
	}
}

// Get all disabled modules of a certain type as a map (named)
func (mt *ModuleTree) GetDisabled(mtype modtype) interface{} {
	defer mt.wrapper()()

	switch mtype {
	case ModCommsChanTx:
		return mt.modCommsChanTx[0]
	case ModCommsChanRx:
		return mt.modCommsChanRx[0]
	case ModProtoLang:
		return mt.modProtoLang[0]
	case ModProtoPart:
		return mt.modProtoPart[0]
	default:
		return map[string]GenericModule{}
	}
}
