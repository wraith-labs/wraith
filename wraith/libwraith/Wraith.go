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

	// Public objects

	Conf    WraithConf
	Modules ModuleTree
}

func (w *Wraith) Init() {
	w.initTime = time.Now()
	w.exitTrigger = make(chan struct{})
	w.unifiedTxQueue = make(TxQueue, 5)
	w.unifiedRxQueue = make(RxQueue, 5)

	// Init modules
	if transmitters, ok := w.Modules.GetEnabled(ModCommsChanTx).(map[string]CommsChanTxModule); ok {
		for _, transmitter := range transmitters {
			transmitter.WraithModuleInit(w)
		}
	}
	if receivers, ok := w.Modules.GetEnabled(ModCommsChanRx).(map[string]CommsChanRxModule); ok {
		for _, receiver := range receivers {
			receiver.WraithModuleInit(w)
		}
	}
	if langs, ok := w.Modules.GetEnabled(ModProtoLang).(map[string]ProtoLangModule); ok {
		for _, lang := range langs {
			lang.WraithModuleInit(w)
		}
	}
	if parts, ok := w.Modules.GetEnabled(ModProtoPart).(map[string]ProtoPartModule); ok {
		for _, part := range parts {
			part.WraithModuleInit(w)
		}
	}
}

func (w *Wraith) PushTx(tx TxQueueElement) {
	w.unifiedTxQueue <- tx
}

func (w *Wraith) PushRx(rx RxQueueElement) {
	w.unifiedRxQueue <- rx
}

func (w *Wraith) Run() {
	// Always stop transmitters and receivers before exiting
	defer func() {
		if transmitters, ok := w.Modules.GetEnabled(ModCommsChanTx).(map[string]CommsChanTxModule); ok {
			for _, transmitter := range transmitters {
				transmitter.StopTx()
			}
		}
		if receivers, ok := w.Modules.GetEnabled(ModCommsChanRx).(map[string]CommsChanRxModule); ok {
			for _, receiver := range receivers {
				receiver.StopRx()
			}
		}
	}()

	// Start transmitters and receivers
	if transmitters, ok := w.Modules.GetEnabled(ModCommsChanTx).(map[string]CommsChanTxModule); ok {
		for _, transmitter := range transmitters {
			transmitter.StartTx()
		}
	}
	if receivers, ok := w.Modules.GetEnabled(ModCommsChanRx).(map[string]CommsChanRxModule); ok {
		for _, receiver := range receivers {
			receiver.StartRx()
		}
	}

	// Spawn and init Tx and Rx handlers
	txh := TxHandler{}
	txh.Init(w)

	rxh := RxHandler{}
	rxh.Init(w)

	// Mainloop: transmit, receive and process stuff
	for {
		select {
		case <-w.exitTrigger:
			return
		case outbound := <-w.unifiedTxQueue:
			go txh.Handle(outbound)
		case inbound := <-w.unifiedRxQueue:
			go rxh.Handle(inbound)
		}
	}
}

func (w *Wraith) Shutdown() {
	// Trigger exit of mainloop
	close(w.exitTrigger)
}
