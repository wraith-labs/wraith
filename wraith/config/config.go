package config

var Config ConfigSkeleton

func Init() {
	Config.Radio.Transmitter.URLGenerator = "package gen\nfunc Gen(old string) string {\nreturn 'http://localhost:8080'\n}"
	Config.Radio.Transmitter.Trigger = "package trigger\nfunc Trigger(curTimestamp int) bool {\nif curTimestamp % 5 == 0 {\nreturn true\n} else {\nreturn false\n}\n}"
	Config.Radio.Transmitter.TriggerCheckInterval = 1
	Config.Radio.Transmitter.Encryption.Enabled = false
	Config.Radio.Transmitter.Encryption.Type = 0
	Config.Radio.Transmitter.Encryption.Key = ""

	Config.Radio.Receiver.URLGenerator = "package gen\nfunc Gen(old string) string {\nreturn 'http://localhost:8080'\n}"
	Config.Radio.Receiver.Trigger = "package trigger\nfunc Trigger(curTimestamp int) bool {\nif curTimestamp % 6 == 0 {\nreturn true\n} else {\nreturn false\n}\n}"
	Config.Radio.Receiver.TriggerCheckInterval = 1
	Config.Radio.Receiver.Encryption.Enabled = false
	Config.Radio.Receiver.Encryption.Type = 0
	Config.Radio.Receiver.Encryption.Key = ""

}
