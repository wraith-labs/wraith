package proto

import (
	"errors"
	"regexp"
)

type Format struct {
	Encode func(map[string]interface{}) []byte
	Decode func([]byte) (map[string]interface{}, error)
}

var formats map[string]*Format

// Register a format such that it can be encoded to / decoded from
func RegFormat(summary string, format *Format) {
	formats[summary] = format
}

// Get a list of formats and their descriptions
func GetFormats() map[string]*Format {
	return formats
}

// Encode data to a specific format
func Encode(data map[string]interface{}, formatSummary string) ([]byte, error) {
	if format, exists := formats[formatSummary]; exists {
		return format.Encode(data), nil
	} else {
		return []byte{}, errors.New("no such format: " + formatSummary)
	}
}

// Decode data from specific format
func Decode(data []byte, formatSummary string) (map[string]interface{}, error) {
	if format, exists := formats[formatSummary]; exists {
		return format.Decode(data)
	} else {
		return make(map[string]interface{}), errors.New("no such format: " + formatSummary)
	}
}

// Decode data but guess the format
func DecodeGuess(data []byte) (map[string]interface{}, error) {
	for formatSummary, format := range formats {
		if matched, err := regexp.Match(formatSummary, []byte(data)); matched && err == nil {
			return format.Decode(data)
		}
	}
	return make(map[string]interface{}), errors.New("format could not be guessed")
}
