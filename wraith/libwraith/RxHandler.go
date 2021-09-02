package libwraith

type RxHandler struct {
	wraith *Wraith
	hkvs   HandlerKeyValueStore
}

func (h *RxHandler) handlePart(name string, data interface{}) {
	if handlers, ok := h.wraith.Modules.GetEnabled(ModCommsChanRx).(map[string]ProtoPartModule); ok {
		if handler, ok := handlers[name]; ok {
			handler.ProcessProtoPart(&h.hkvs, data)
		}
	}
}

func (h *RxHandler) HandleData(data []byte) []string {
	// TODO: Check if data is signed

	// Attempt to translate data from any known format to a map[string]interface{}
	dataMap, err := h.wraith.translator.DecodeGuess(data)
	if err != nil {
		// The data failed to decode (most likely didn't match any known format or the signature was invalid) so it should be discarded
		return []string{}
	}

	// Set up HandlerKeyValueStore with special data
	h.hkvs.Set("w.msg.results", []string{}) // Array of all "output" from handlers
	h.hkvs.Set("w.validity.isValid", true)  // Initially should be valid in case `w.validity` not specified

	// The w.validity data key is special - it decides whether the rest of the keys are evaluated
	// If it's present, always handle it first so other handlers don't have to wait
	if validity, ok := dataMap["w.validity"]; ok {
		h.handlePart("w.validity", validity)
		delete(dataMap, "w.validity")
	}

	// Handle all else
	for key, value := range dataMap {
		h.handlePart(key, value)
	}

	// Return the results (we set the key above so it'll definitely exist)
	results, _ := h.hkvs.Get("w.msg.results")
	// It's best to make sure that it's a []string though
	if resultsStrArr, ok := results.([]string); ok {
		return resultsStrArr
	} else {
		return []string{"HandleData was unable to fetch the results array - a handler module is likely broken"}
	}
}

func (h *RxHandler) Init(wraith *Wraith) {
	// Save our pointer to the Wraith
	h.wraith = wraith

	// Init HKVS
	h.hkvs.Init()
}
