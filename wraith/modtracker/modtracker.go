// This package imports and keeps track of all modular parts of Wraith. They must register
// with this module in order to be used. They can also be de-registered and re-registered
// dynamically.

package modtracker

// "Enum" of the available module types
const (
	ModProtoLang = iota
	ModProtoPart
	ModCommsTx
	ModCommsRx

	// As long as this is last, it will equal the number of entries above because
	// iotas start at 0 so (last entry)+1 == total
	TotalModTypes
)

// Arrays of maps of the modules which are currently enabled or disabled
var moduleTreeEnabled [TotalModTypes]map[string]interface{}
var moduleTreeDisabled [TotalModTypes]map[string]interface{}

// Register and enable a module so that it can be used by Wraith
func RegisterModule(modtype int, modname string, mod interface{}) {
	moduleTreeEnabled[modtype][modname] = mod
}

// Semi-permanently (does not survive Wraith re-starts) remove a module
// This can save memory if a module is guaranteed to not be needed anymore, but is
// very risky because the module can never be re-added without re-starting Wraith
func DeregisterModule(modtype int, modname string) {
	// Make sure to delete both if enabled and disabled
	// delete() is a no-op when the key does not exist so it's safe not to check
	delete(moduleTreeEnabled[modtype], modname)
	delete(moduleTreeDisabled[modtype], modname)
}

// If a module is currently registered but disabled, enable it
func EnableModule(modtype int, modname string) {
	if mod, exists := moduleTreeDisabled[modtype][modname]; exists {
		moduleTreeEnabled[modtype][modname] = mod
		delete(moduleTreeDisabled[modtype], modname)
	}
}

// If a module is currently registered and enabled, disable it
func DisableModule(modtype int, modname string) {
	if mod, exists := moduleTreeEnabled[modtype][modname]; exists {
		moduleTreeDisabled[modtype][modname] = mod
		delete(moduleTreeEnabled[modtype], modname)
	}
}

// Get all enabled modules of a certain type as an array (nameless)
func GetEnabledNameless(modtype int) []interface{} {
	mods := []interface{}{}
	for _, mod := range moduleTreeEnabled[modtype] {
		mods = append(mods, mod)
	}
	return mods
}

// Get the names of all enabled modules of a certain type as an array
func GetEnabledNameOnly(modtype int) []string {
	modNames := []string{}
	for modName := range moduleTreeEnabled[modtype] {
		modNames = append(modNames, modName)
	}
	return modNames
}

// Get all enabled modules of a certain type as a map (named)
func GetEnabledNamed(modtype int) map[string]interface{} {
	return moduleTreeEnabled[modtype]
}

// Get all enabled modules of all types
func GetAllEnabled() [TotalModTypes]map[string]interface{} {
	return moduleTreeEnabled
}

// Get all disabled modules of a certain type as an array (nameless)
func GetDisabledNameless(modtype int) []interface{} {
	mods := []interface{}{}
	for _, mod := range moduleTreeDisabled {
		mods = append(mods, mod)
	}
	return mods
}

// Get the names of all disabled modules of a certain type as an array
func GetDisabledNameOnly(modtype int) []string {
	modNames := []string{}
	for modName := range moduleTreeDisabled[modtype] {
		modNames = append(modNames, modName)
	}
	return modNames
}

// Get all disabled modules of a certain type as a map (named)
func GetDisabledNamed(modtype int) map[string]interface{} {
	return moduleTreeDisabled[modtype]
}

// Get all disabled modules of all types
func GetAllDisabled() [TotalModTypes]map[string]interface{} {
	return moduleTreeDisabled
}

// Prepare all module stuff
func init() {
	// Init storages
	for i := TotalModTypes - 1; i >= 0; i-- {
		moduleTreeEnabled[i] = make(map[string]interface{})
		moduleTreeDisabled[i] = make(map[string]interface{})
	}
}
