package ldp

import "ldpserver/util"

type Settings struct {
	dataPath string
	rootUri  string
	idFile   string
}

func SettingsNew(rootUri, datapath string) Settings {
	var sett Settings
	sett.rootUri = util.StripSlash(rootUri)
	sett.dataPath = util.PathConcat(datapath, "/")
	sett.idFile = util.PathConcat(sett.dataPath, "meta.rdf.id")
	return sett
}

func (settings Settings) DataPath() string {
	return settings.dataPath
}

func (settings Settings) RootUri() string {
	return settings.rootUri
}

func (settings Settings) IdFile() string {
	return settings.idFile
}
