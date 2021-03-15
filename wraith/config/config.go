package config

var Config ConfigSkeleton

func init() {
	Config.Process.RespectExitSignals = true

	Config.Comms.Transmitter.URLGenerator = "package gen\nfunc Gen(old string) string {\nreturn \"http://localhost:8080\"\n}"
	Config.Comms.Transmitter.Trigger = "package trigger\nfunc Trigger(curTimestamp int) bool {\nif curTimestamp % 5 == 0 {\nreturn true\n} else {\nreturn false\n}\n}"
	Config.Comms.Transmitter.TriggerCheckInterval = 1
	Config.Comms.Transmitter.Encryption.Enabled = false
	Config.Comms.Transmitter.Encryption.Type = 0
	Config.Comms.Transmitter.Encryption.Key = ""

	Config.Comms.Receiver.URLGenerator = "package gen\nfunc Gen(old string) string {\nreturn \"http://localhost:8080\"\n}"
	Config.Comms.Receiver.Trigger = "package trigger\nfunc Trigger(curTimestamp int) bool {\nif curTimestamp % 6 == 0 {\nreturn true\n} else {\nreturn false\n}\n}"
	Config.Comms.Receiver.TriggerCheckInterval = 1
	Config.Comms.Receiver.Encryption.Enabled = false
	Config.Comms.Receiver.Encryption.Type = 0
	Config.Comms.Receiver.Encryption.Key = ""
}
