package main

import "github.com/traefik/yaegi/interp"

type radio struct {
	CheckURLGenerator  string
	NextCheckSeconds   int
	LastCheckTimestamp int
	LastCheckURL       string
	CheckVerifPkey     string
	CheckVerif         bool
	CheckCryptKey      string
}

func (r *radio) GenerateCheckURL() (string, error) {
	// Use yaegi to run function to generate the next check URL
	i := interp.New(interp.Options{})
	_, err := i.Eval(r.CheckURLGenerator)
	if err != nil {
		return "", err
	}
	v, err := i.Eval("gen.Gen")
	if err != nil {
		return "", err
	}
	gen := v.Interface().(func(string) string)
	result := gen(r.LastCheckURL)
	return result, nil
}

func MakeDNSRequest(dnsname, record string) string {
	return "a"
}

func (r *radio) Check() error {
	return nil
}
