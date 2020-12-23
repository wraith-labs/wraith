package config

import "time"

type ConfigSkeleton struct {
	Radio struct {
		Transmitter struct {
			DefaultURLGenerator    string
			DefaultIntervalSeconds time.Duration
			Encryption             struct {
				Enabled bool
				Type    int
				Key     string
			}
		}
		Receiver struct {
			DefaultURLGenerator    string
			DefaultIntervalSeconds time.Duration
			Encryption             struct {
				Enabled bool
				Type    int
				Key     string
			}
		}
	}
}
