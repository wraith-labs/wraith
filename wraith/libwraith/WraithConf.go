package libwraith

type WraithConf struct {
	Fingerprint           string
	DefaultReturnAddr     string
	DefaultReturnEncoding string
	Custom                map[string]interface{}
}
