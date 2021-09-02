package libwraith

import (
	"time"
)

type Wraith struct {
	// Internal information

	initTime time.Time

	// Internal communication

	exitTrigger    chan struct{}
	unifiedTxQueue TxQueue
	unifiedRxQueue RxQueue

	// Internal objects

	translator Translator

	// Public objects

	Conf    WraithConf
	Modules ModuleTree
}

func (w *Wraith) Start() {
	// Mainloop: Transmit, receive and process stuff
	for {
		// TODO: Find what is concurrent and what is not to catch points where Wraith can break/stall
		select {
		case <-w.unifiedRxQueue:
			// When data is received, run the OnRx handlers
			// TODO: Do this
		case <-w.exitTrigger:
			return
		}
	}
}

func (w *Wraith) Stop() {
	close(w.exitTrigger)
}

func (w *Wraith) Init() {
	w.initTime = time.Now()
	w.exitTrigger = make(chan struct{})

	w.translator.Init(w)
}
