package proto

import (
	"fmt"

	mm "git.0x1a8510f2.space/0x1a8510f2/wraith/modmgr"
)

// Encode data to a specific format
func Encode(data map[string]interface{}, formatName string) ([]byte, error) {
	if formats, ok := mm.Modules.GetEnabled(mm.ModProtoLang).(map[string]mm.ProtoLangModule); ok {
		if format, exists := formats[formatName]; exists {
			return format.Encode(data)
		}
	}
	return nil, fmt.Errorf("no such format: %s", formatName)
}

// Decode data from specific format
func Decode(data []byte, formatName string) (map[string]interface{}, error) {
	if formats, ok := mm.Modules.GetEnabled(mm.ModProtoLang).(map[string]mm.ProtoLangModule); ok {
		if format, exists := formats[formatName]; exists {
			return format.Decode(data)
		}
	}
	return nil, fmt.Errorf("no such format: %s", formatName)
}

// Decode data but guess the format
func DecodeGuess(data []byte) (map[string]interface{}, error) {
	if formats, ok := mm.Modules.GetEnabled(mm.ModProtoLang).(map[string]mm.ProtoLangModule); ok {
		for _, format := range formats {
			if format.Identify(data) {
				return format.Decode(data)
			}
		}
	}
	return nil, fmt.Errorf("format could not be guessed")
}
