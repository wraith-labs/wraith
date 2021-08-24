package proto

import (
	"git.0x1a8510f2.space/0x1a8510f2/wraith/hooks"
	"git.0x1a8510f2.space/0x1a8510f2/wraith/proto/parts"
)

func HandleData(data []byte) []string {
	// TODO: Check if data is signed

	// Attempt to translate data from any known format to a map[string]interface{}
	dataMap, err := DecodeGuess(data)
	if err != nil {
		// The data failed to decode (most likely didn't match any known format or the signature was invalid) so it should be discarded
		return []string{}
	}

	// Set up HandlerKeyValueStore with special data
	parts.PartMap.GetHKVS().Set("w.msg.results", []string{}) // Array of all "output" from handlers
	parts.PartMap.GetHKVS().Set("w.validity.isValid", true)  // Initially should be valid in case `w.validity` not specified

	// The w.validity data key is special - it decides whether the rest of the keys are evaluated
	// If it's present, always handle it first so other handlers don't have to wait
	if validity, ok := dataMap["w.validity"]; ok {
		parts.PartMap.Handle("w.validity", validity)
		delete(dataMap, "w.validity")
	}

	// Handle all else
	for key, value := range dataMap {
		parts.PartMap.Handle(key, value)
	}

	// Return the results (we set the key above so it'll definitely exist)
	results, _ := parts.PartMap.GetHKVS().Get("w.msg.results")
	// It's best to make sure that it's a []string though
	if resultsStrArr, ok := results.([]string); ok {
		return resultsStrArr
	} else {
		return []string{"HandleData was unable to fetch the results array - a handler module is likely broken"}
	}
}

func init() {
	// Register HandleData as an OnRx hook
	hooks.OnRx.Add(HandleData)
}
