package libwraith

import "net/url"

type Comms struct {
	wraith      *Wraith
	exitTrigger chan struct{}
}

// Infinite loop managing data transmission
// This should run in a thread and only a single instance should run at a time
func (c *Comms) Start() {
	// Always stop transmitters and receivers before exiting
	defer func() {
		if transmitters, ok := c.wraith.Modules.GetEnabled(ModCommsChanTx).(map[string]CommsChanTxModule); ok {
			for _, transmitter := range transmitters {
				transmitter.StopTx()
			}
		}
		if receivers, ok := c.wraith.Modules.GetEnabled(ModCommsChanRx).(map[string]CommsChanRxModule); ok {
			for _, receiver := range receivers {
				receiver.StopRx()
			}
		}
		close(c.exitTrigger)
	}()

	// Start transmitters and receivers
	if transmitters, ok := c.wraith.Modules.GetEnabled(ModCommsChanTx).(map[string]CommsChanTxModule); ok {
		for _, transmitter := range transmitters {
			transmitter.StartTx()
		}
	}
	if receivers, ok := c.wraith.Modules.GetEnabled(ModCommsChanRx).(map[string]CommsChanRxModule); ok {
		for _, receiver := range receivers {
			receiver.StartRx()
		}
	}

	for {
		select {
		case <-c.exitTrigger:
			return
		case transmission := <-c.wraith.unifiedTxQueue:
			txaddr, err := url.Parse(transmission.Addr)
			// If there was an error parsing the URL, the whole txdata should be dropped as there's nothing more we can do
			if err == nil {
				// ...same in case of a non-existent transmitter
				if transmitters, ok := c.wraith.Modules.GetEnabled(ModCommsChanTx).(map[string]CommsChanTxModule); ok {
					if transmitter, exists := transmitters[txaddr.Scheme]; exists {
						transmitter.TriggerTx(transmission)
					}
				}
			}
		}
	}
}

func (c *Comms) Stop() {
	close(c.exitTrigger)
}

func (c *Comms) Init(wraith *Wraith) {
	// Save Wraith pointer
	c.wraith = wraith

	// Init exit trigger channel
	c.exitTrigger = make(chan struct{})
}
