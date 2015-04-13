package ldp

var dataPath = "/Users/hector/dev/gotest/src/ldpserver/data_test"
var rootUrl = "http://localhost:9001/"
var settings Settings

func init() {
	settings = SettingsNew(rootUrl, dataPath)
	CreateRoot(settings)
}
