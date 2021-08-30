package types

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
