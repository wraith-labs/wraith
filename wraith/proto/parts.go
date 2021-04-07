package proto

var PartMap ProtoPartsMap

type HandlerKeyValueStore struct {
	data map[string]interface{}
}

func (hkvs *HandlerKeyValueStore) Init() {
	if hkvs.data == nil {
		hkvs.data = make(map[string]interface{})
	}
}

type ProtoPartsMap struct {
	data map[string]func(HandlerKeyValueStore, interface{})
	hkvs HandlerKeyValueStore
}

func (ppm *ProtoPartsMap) Init() {
	if ppm.data == nil {
		ppm.data = make(map[string]func(HandlerKeyValueStore, interface{}))
	}
	ppm.hkvs.Init()
}

func (ppm *ProtoPartsMap) Add(target string, handler func(HandlerKeyValueStore, interface{})) {
	ppm.data[target] = handler
}

func (ppm *ProtoPartsMap) Handle(target string, data interface{}) {
	// If a handler for the key exists, execute the handler - otherwise do nothing
	if handler, ok := ppm.data[target]; ok {
		handler(ppm.hkvs, data)
	}
}

func init() {
	PartMap = ProtoPartsMap{}
	PartMap.Init()
}
