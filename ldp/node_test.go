package ldp

import (
	"os"
	"path/filepath"
)

var dataPath string
var rootUrl = "http://localhost:9001/"
var settings Settings

func init() {
	dataPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	settings = SettingsNew(rootUrl, dataPath)
	CreateRoot(settings)
}
