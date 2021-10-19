package libwraith

import (
	"net/url"
	"time"
)

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
	// Check if the outbound data has hit the retransmission cap
	if outbound.TransmissionFailCount >= h.wraith.Conf.RetransmissionCap-1 {
		// Drop it if so
		return
	}

	// Check if the outbound data has failed transmission recently and wait
	// until the retransmission delay expires if so
	if retransmitDelay := time.Until(outbound.TransmissionFailTime.Add(h.wraith.Conf.RetransmissionDelay)); !outbound.TransmissionFailTime.IsZero() && retransmitDelay > 0 {
		time.Sleep(retransmitDelay)
	}

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
					if !transmitter.TriggerTx(outbound.Addr, data) {
						// The sending could have failed though, so we will need to re-attempt it
						outbound.TransmissionFailCount += 1
						outbound.TransmissionFailTime = time.Now()
						h.wraith.PushTx(outbound)
					}
				}
			}
		}
	}
}
