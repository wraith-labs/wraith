package config

type ConfigSkeleton struct {
	Radio struct {
		Transmitter struct {
			URLGenerator         string
			Trigger              string
			TriggerCheckInterval int
			Encryption           struct {
				Enabled bool
				Type    int
				Key     string
			}
		}
		Receiver struct {
			URLGenerator         string
			Trigger              string
			TriggerCheckInterval int
			Encryption           struct {
				Enabled bool
				Type    int
				Key     string
			}
		}
	}
}
