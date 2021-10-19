package libwraith

type RxHandler struct {
	wraith     *Wraith
	hkvs       HandlerKeyValueStore
	translator Translator
}

func (h *RxHandler) Init(wraith *Wraith) {
	// Save our pointer to the Wraith
	h.wraith = wraith

	// Init HKVS
	h.hkvs.Init()

	// Init translator
	h.translator.Init(wraith)
}

func (h *RxHandler) handlePart(name string, data interface{}) {
	if handlers, ok := h.wraith.Modules.GetEnabled(ModProtoPart).(map[string]ProtoPartModule); ok {
		if handler, ok := handlers[name]; ok {
			handler.ProcessProtoPart(&h.hkvs, data)
		}
	}
}

func (h *RxHandler) ConstValidityModule() string {
	return "validity"
}

func (h *RxHandler) ConstValidityKey() string {
	return "validity.valid"
}

func (h *RxHandler) ConstReturnModule() string {
	return "return"
}

func (h *RxHandler) ConstReturnAddrKey() string {
	return "return.addr"
}

func (h *RxHandler) ConstReturnEncodeKey() string {
	return "return.encode"
}

func (h *RxHandler) Handle(inbound RxQueueElement) {
	data := inbound.Data

	// Attempt to translate data from any known format to a map[string]interface{}
	dataMap, err := h.translator.DecodeGuess(data)
	if err != nil {
		// The data failed to decode (most likely didn't match any known format or
		// the signature was invalid) so it should be discarded
		return
	}

	// Set up HandlerKeyValueStore with special data
	h.hkvs.Set(h.ConstValidityKey(), true)                                    // Initially should be valid in case `w.validity` not specified
	h.hkvs.Set(h.ConstReturnAddrKey(), h.wraith.Conf.DefaultReturnAddr)       // Initially should be the default return addr
	h.hkvs.Set(h.ConstReturnEncodeKey(), h.wraith.Conf.DefaultReturnEncoding) // Initially should be the default return encoding

	// The w.validity data key is special - it decides whether the rest of the keys are evaluated
	// If it's present, always handle it first so other handlers don't have to wait
	if validity, ok := dataMap[h.ConstValidityModule()]; ok {
		h.handlePart(h.ConstValidityModule(), validity)
		delete(dataMap, h.ConstValidityModule())
	}

	// The w.return data key is also special - it decides how data is returned
	// If it's present, always handle it second so other handlers don't have to wait
	if returnMethod, ok := dataMap[h.ConstReturnModule()]; ok {
		h.handlePart(h.ConstReturnModule(), returnMethod)
		delete(dataMap, h.ConstReturnModule())
	}

	// Handle all else
	for key, value := range dataMap {
		h.handlePart(key, value)
	}
}
