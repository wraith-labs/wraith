package config

type ConfigSkeleton struct {
	Radio struct {
		Transmitter struct {
			DefaultURLGenerator    string
			DefaultIntervalSeconds int
			Encryption             struct {
				Enabled bool
				Type    int
				Key     string
			}
		}
		Receiver struct {
			DefaultURLGenerator    string
			DefaultIntervalSeconds int
			Encryption             struct {
				Enabled bool
				Type    int
				Key     string
			}
		}
	}
}
