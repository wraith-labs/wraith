package mod_lang

import (
	"encoding/json"
	"fmt"

	"git.0x1a8510f2.space/0x1a8510f2/wraith/libwraith"
)

type DebugModule struct{}

func (m *DebugModule) WraithModuleInit(wraith *libwraith.Wraith) {
	fmt.Printf("DEBUG: mod_lang.DebugModule.WraithModuleInit called\n")
}
func (m *DebugModule) ProtoLangModule() {}

func (m *DebugModule) Encode(data map[string]interface{}) ([]byte, error) {
	fmt.Printf("DEBUG: mod_lang.DebugModule.Encode called with params: %v\n", data)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	stringData := append([]byte("DEBUG:"), jsonData...)
	return stringData, nil
}

func (m *DebugModule) Decode(data []byte) (map[string]interface{}, error) {
	fmt.Printf("DEBUG: mod_lang.DebugModule.Decode called with params: %v\n", data)

	if string(data[0:6]) != "DEBUG:" {
		return nil, fmt.Errorf("not debug data")
	}

	data = data[6:]

	var dataDict map[string]interface{}
	err := json.Unmarshal(data, &dataDict)
	if err != nil {
		return nil, err
	}

	return dataDict, nil
}

func (m *DebugModule) Identify(data []byte) bool {
	fmt.Printf("DEBUG: mod_lang.DebugModule.Identify called with params: %v\n", data)

	if string(data[0:6]) != "DEBUG:" {
		return false
	}

	data = data[6:]

	var dataDict map[string]interface{}
	err := json.Unmarshal(data, &dataDict)

	return err == nil
}
