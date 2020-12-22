package settings

var Settings SettingsSkeleton

func Init() {
	Settings.Radio.Transmitter.DefaultURLGenerator = "package gen"
}
