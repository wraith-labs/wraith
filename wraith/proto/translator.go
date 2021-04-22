package proto

import (
	"errors"
)

/*type Format struct { // TODO: Interface or struct
	Encode      func(map[string]interface{}) []byte
	Decode      func([]byte) (map[string]interface{}, error)
	Identify    func([]byte) bool
	DecodeCache struct {
		Id    string
		Value map[string]interface{}
	}
}*/

type Format interface {
	Encode(map[string]interface{}) []byte
	Decode([]byte) (map[string]interface{}, error)
	Identify([]byte) bool
}

var formats map[string]*Format

// Register a format such that it can be encoded to / decoded from / identified
func RegFormat(name string, format *Format) {
	formats[name] = format
}

// Get a list of formats and their names
func GetFormats() map[string]*Format {
	return formats
}

// Encode data to a specific format
func Encode(data map[string]interface{}, formatName string) ([]byte, error) {
	if format, exists := formats[formatName]; exists {
		return (*format).Encode(data), nil
	} else {
		return []byte{}, errors.New("no such format: " + formatName)
	}
}

// Decode data from specific format
func Decode(data []byte, formatName string) (map[string]interface{}, error) {
	if format, exists := formats[formatName]; exists {
		return (*format).Decode(data)
	} else {
		return make(map[string]interface{}), errors.New("no such format: " + formatName)
	}
}

// Decode data but guess the format
func DecodeGuess(data []byte) (map[string]interface{}, error) {
	for _, format := range formats {
		if (*format).Identify(data) {
			return (*format).Decode(data)
		}
	}
	return make(map[string]interface{}), errors.New("format could not be guessed")
}
