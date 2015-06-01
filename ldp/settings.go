package ldp

import "ldpserver/util"

type Settings struct {
	dataPath string
	rootUri  string
	// rootBagOnDisk  string
	// rootNodeOnDisk string
	idFile string
}

func SettingsNew(rootUri, datapath string) Settings {
	var sett Settings
	sett.rootUri = util.StripSlash(rootUri)
	sett.dataPath = util.PathConcat(datapath, "/")
	// sett.rootBagOnDisk = util.PathConcat(sett.dataPath, "bagit.txt")
	// sett.rootNodeOnDisk = util.PathConcat(sett.dataPath, "data/meta.rdf")
	sett.idFile = util.PathConcat(sett.dataPath, "data/meta.rdf.id")
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
