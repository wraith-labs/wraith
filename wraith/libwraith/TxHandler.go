package libwraith

import "net/url"

type TxHandler struct {
	wraith     *Wraith
	translator Translator
}

func (h *TxHandler) Init(wraith *Wraith) {
	// Save our pointer to the Wraith
	h.wraith = wraith

	// Init translator
	h.translator.Init(wraith)
}

func (h *TxHandler) Handle(outbound TxQueueElement) {
	txaddr, err := url.Parse(outbound.Addr)
	// If there was an error parsing the URL, the whole txdata should be dropped as there's nothing more we can do
	if err == nil {
		// ...same in case of a non-existent transmitter
		if transmitters, ok := h.wraith.Modules.GetEnabled(ModCommsChanTx).(map[string]CommsChanTxModule); ok {
			if transmitter, exists := transmitters[txaddr.Scheme]; exists {
				// ...but if all went well, try to encode the payload
				data, err := h.translator.Encode(outbound.Data, outbound.Encoding)
				// If that failed, we can't do anything so skip
				if err == nil {
					// ...but if it did not, send away!
					transmitter.TriggerTx(outbound.Addr, data)
				}
			}
		}
	}
}
