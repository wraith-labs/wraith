package frequencies

type DNS struct{}

func (freq DNS) Transmit(url string, data []byte) error {
	return nil
}

func (freq DNS) Receive(url string) ([]byte, error) {
	return []byte{}, nil
}
