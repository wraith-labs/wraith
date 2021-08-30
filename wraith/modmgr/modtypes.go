package modmgr

import "git.0x1a8510f2.space/0x1a8510f2/wraith/types"

type CommsChanTxModule interface {
	Start()
	Stop()
	Trigger(data types.TxQueueElement) bool
}

type CommsChanRxModule interface {
	Start()
	Stop()
}

type ProtoLangModule interface {
}

type ProtoPartModule interface {
	Process(hkvs *types.HandlerKeyValueStore, data interface{})
}
