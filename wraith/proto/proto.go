package proto

func HandleData(data []byte) {
	// TODO: Check if data is signed

	// Attempt to translate data from any known format to a map[string]interface{}
	dataMap, err := DecodeGuess(data)
	if err != nil {
		// The data failed to decode (most likely didn't match any known format or the signature was invalid) so it should be discarded
		return
	}

	// The w.validity key is special - it decides whether the rest of the keys are evaluated
	// If it's present, always handle it first
	if validity, ok := dataMap["w.validity"]; ok {
		PartMap.Handle("w.validity", validity)
		delete(dataMap, "w.validity")
	}

	// Handle all else
	for key, value := range dataMap {
		PartMap.Handle(key, value)
	}
}
