package libwraith

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

// A pre-made function for generating Wraith fingerprints based on information
// about both the host and about the Wraith process. This functions generates
// fingerprints of varying length. Standard length can be achieved by hashing
// the output of this function. This function is useful if you don't mind
// fingerprints changing when Wraith is restarted, but would still like some
// indication of which Wraith runs on which host. If multiple Wraiths run on
// one host, this also allows for differentiation between them.
func FingerprintGenerator_HostAndProcess() string {
	// Gather fingerprint data
	hostname, hostnameErr := os.Hostname()
	if hostnameErr != nil {
		hostname = "(UNKNOWN)"
	}
	var hwaddrs string
	ifaces, ifacesErr := net.Interfaces()
	if ifacesErr == nil {
		for _, iface := range ifaces {
			hwaddrs += iface.HardwareAddr.String() + ";"
		}
	}
	uid := os.Getuid()
	gid := os.Getegid()
	pid := os.Getpid()
	ppid := os.Getppid()

	// Format fingerprint data
	return fmt.Sprintf("<host[%s] hwaddrs[%s] uid[%d] gid[%d] pid[%d] ppid[%d]>", hostname, hwaddrs, uid, gid, pid, ppid)
}

// A pre-made function for generating Wraith fingerprints based on information
// only about the host. This functions generates fingerprints of varying length.
// Standard length can be achieved by hashing the output of this function. This
// function is useful if you don't want Wraith fingerprint to change at all on
// restart. If multiple Wraiths run on one host, they will both have the same
// fingerprint.
func FingerprintGenerator_HostOnly() string {
	// Gather fingerprint data
	hostname, hostnameErr := os.Hostname()
	if hostnameErr != nil {
		hostname = "(UNKNOWN)"
	}
	var hwaddrs string
	ifaces, ifacesErr := net.Interfaces()
	if ifacesErr == nil {
		for _, iface := range ifaces {
			hwaddrs += iface.HardwareAddr.String() + ";"
		}
	}

	// Format fingerprint data
	return fmt.Sprintf("<host[%s] hwaddrs[%s]>", hostname, hwaddrs)
}

// A pre-made function for generating Wraith fingerprints which are random
// and not based on any information whatsoever. Useful if you don't really
// care about the fingerprints and just need them to uniquely identify a
// Wraith, but they can change on restart.
func FingerprintGenerator_Random() string {
	// Apparently this is hyper-optimised
	// Might as well ¯\_(ツ)_/¯
	// Stolen from: https://stackoverflow.com/a/31832326/8623347

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	const length = 32
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)

	var src = rand.NewSource(time.Now().UnixNano())

	sb := strings.Builder{}
	sb.Grow(length)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := length-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(charset) {
			sb.WriteByte(charset[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}
