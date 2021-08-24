package modmgr

type CommsChanRxModule struct {
	Start func()
	Stop  func()
	Data  map[string]interface{}
}

func (CommsChanRxModule) isWraithModule()

type CommsChanTxModule struct {
	Start   func()
	Stop    func()
	Trigger func(data TxQueueElement) bool
	Data    map[string]interface{}
}

func (CommsChanTxModule) isWraithModule()

type ProtoLangModule struct {
}

func (ProtoLangModule) isWraithModule()

type ProtoPartModule struct {
	Process func(hkvs *HandlerKeyValueStore, data interface{})
}

func (ProtoPartModule) isWraithModule()
