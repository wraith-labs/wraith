package proto

import "github.com/0x1a8510f2/wraith/hooks"

func HandleData(data []byte) []string {
	// TODO: Check if data is signed

	// Attempt to translate data from any known format to a map[string]interface{}
	dataMap, err := DecodeGuess(data)
	if err != nil {
		// The data failed to decode (most likely didn't match any known format or the signature was invalid) so it should be discarded
		return []string{}
	}

	// Set up HandlerKeyValueStore with special data
	PartMap.GetHKVS().data["w.msg.results"] = []string{} // Array of all "output" from handlers
	PartMap.GetHKVS().data["w.validity.isValid"] = true  // Initially should be valid in case `w.validity` not specified

	// The w.validity data key is special - it decides whether the rest of the keys are evaluated
	// If it's present, always handle it first so other handlers don't have to wait
	if validity, ok := dataMap["w.validity"]; ok {
		PartMap.Handle("w.validity", validity)
		delete(dataMap, "w.validity")
	}

	// Handle all else
	for key, value := range dataMap {
		PartMap.Handle(key, value)
	}

	return PartMap.GetHKVS().data["w.msg.results"].([]string)
}

func init() {
	// Register HandleData as an OnRx hook
	hooks.OnRx.Add(HandleData)
}
