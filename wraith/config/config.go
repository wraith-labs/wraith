package config

var Config ConfigSkeleton

func Init() {
	Config.Radio.Transmitter.DefaultURLGenerator = "package gen\nfunc Gen(old string) string {\nreturn 'http://localhost:8080'\n}"
	Config.Radio.Transmitter.DefaultIntervalSeconds = 2
	Config.Radio.Transmitter.Encryption.Enabled = false
	Config.Radio.Transmitter.Encryption.Type = 0
	Config.Radio.Transmitter.Encryption.Key = ""

	Config.Radio.Receiver.DefaultURLGenerator = "package gen\nfunc Gen(old string) string {\nreturn 'http://localhost:8080'\n}"
	Config.Radio.Receiver.DefaultIntervalSeconds = 2
	Config.Radio.Receiver.Encryption.Enabled = false
	Config.Radio.Receiver.Encryption.Type = 0
	Config.Radio.Receiver.Encryption.Key = ""
}
