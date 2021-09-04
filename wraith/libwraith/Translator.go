package libwraith

import "fmt"

type Translator struct {
	wraith *Wraith
}

func (t *Translator) Init(wraith *Wraith) {
	t.wraith = wraith
}

// Encode data to a specific format
func (t *Translator) Encode(data map[string]interface{}, formatName string) ([]byte, error) {
	if formats, ok := t.wraith.Modules.GetEnabled(ModProtoLang).(map[string]ProtoLangModule); ok {
		if format, exists := formats[formatName]; exists {
			return format.Encode(data)
		}
	}
	return nil, fmt.Errorf("no such format: %s", formatName)
}

// Decode data from specific format
func (t *Translator) Decode(data []byte, formatName string) (map[string]interface{}, error) {
	if formats, ok := t.wraith.Modules.GetEnabled(ModProtoLang).(map[string]ProtoLangModule); ok {
		if format, exists := formats[formatName]; exists {
			return format.Decode(data)
		}
	}
	return nil, fmt.Errorf("no such format: %s", formatName)
}

// Decode data but guess the format
func (t *Translator) DecodeGuess(data []byte) (map[string]interface{}, error) {
	if formats, ok := t.wraith.Modules.GetEnabled(ModProtoLang).(map[string]ProtoLangModule); ok {
		for _, format := range formats {
			if format.Identify(data) {
				return format.Decode(data)
			}
		}
	}
	return nil, fmt.Errorf("format could not be guessed")
}
