package modmgr

import "git.0x1a8510f2.space/0x1a8510f2/wraith/types"

// Constants for describing module types to the module tree
const (
	ModCommsChanTx modtype = iota
	ModCommsChanRx
	ModProtoLang
	ModProtoPart
)

type modtype int

// Every module must implement this interface to make sure
// it's meant to be used as a Wraith module
type GenericModule interface {
	WraithModule()
}

type CommsChanTxModule interface {
	GenericModule
	StartTx()
	StopTx()
	TriggerTx(data types.TxQueueElement) bool
	CommsChanTxModule()
}

type CommsChanRxModule interface {
	GenericModule
	StartRx()
	StopRx()
	CommsChanRxModule()
}

type ProtoLangModule interface {
	GenericModule
	Encode(map[string]interface{}) ([]byte, error)
	Decode([]byte) (map[string]interface{}, error)
	Identify([]byte) bool
	ProtoLangModule()
}

type ProtoPartModule interface {
	GenericModule
	ProcessProtoPart(hkvs *types.HandlerKeyValueStore, data interface{})
	ProtoPartModule()
}
