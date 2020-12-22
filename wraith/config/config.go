package config

var Config ConfigSkeleton

func Init() {
	Config.Radio.Transmitter.DefaultURLGenerator = "package gen"
}
