package parts

var PartMap ProtoPartsMap

// A key-value storage allowing for communication between handlers
type HandlerKeyValueStore struct {
	data map[string]interface{}
}

func (hkvs *HandlerKeyValueStore) Init() {
	if hkvs.data == nil {
		hkvs.data = make(map[string]interface{})
	}
}

func (hkvs *HandlerKeyValueStore) Set(key string, value interface{}) {
	hkvs.data[key] = value
}

func (hkvs *HandlerKeyValueStore) Get(key string) (interface{}, bool) {
	data, ok := hkvs.data[key]
	return data, ok
}

// A structure mapping protocol keys to functions which handle them
type ProtoPartsMap struct {
	data map[string]func(*HandlerKeyValueStore, interface{})
	hkvs HandlerKeyValueStore
}

func (ppm *ProtoPartsMap) Init() {
	if ppm.data == nil {
		ppm.data = make(map[string]func(*HandlerKeyValueStore, interface{}))
	}
	ppm.hkvs.Init()
}

func (ppm *ProtoPartsMap) Add(target string, handler func(*HandlerKeyValueStore, interface{})) {
	ppm.data[target] = handler
}

func (ppm *ProtoPartsMap) Handle(target string, data interface{}) {
	// If a handler for the key exists, execute the handler - otherwise do nothing (ignore the key)
	if handler, ok := ppm.data[target]; ok {
		handler(&ppm.hkvs, data)
	}
}

// Get a pointer to the internal HandlerKeyValueStore. This should not really be used unless necessary.
func (ppm *ProtoPartsMap) GetHKVS() *HandlerKeyValueStore {
	return &ppm.hkvs
}

func init() {
	PartMap = ProtoPartsMap{}
	PartMap.Init()
}
