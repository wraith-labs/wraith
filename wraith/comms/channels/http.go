package frequencies

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type HTTP struct{}

// TODO: Set some headers to make Wrait less detectable

func (freq HTTP) Transmit(url string, data []byte) error {
	resp, err := http.Post(url, "application/octet-stream", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Don't bother reading the response, there's not a lot we can do about errors anyway
	return nil
}

func (freq HTTP) Receive(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
