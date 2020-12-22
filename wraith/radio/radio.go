package radio

import (
	"fmt"

	"github.com/TR-SLimey/wraith/config"
	"github.com/traefik/yaegi/interp"
)

func NewRadio() Radio {
	r := Radio{
		Transmitter: Antenna{
			URLGenerator: config.Config.Radio.Transmitter.DefaultURLGenerator,
		}
	}
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
	Decrypter     struct {
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

type Frequency struct {
	ProtoName string
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
