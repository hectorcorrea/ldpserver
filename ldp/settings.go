package ldp

type Settings struct {
	dataPath       string
	rootUri        string
	rootNodeOnDisk string
	idFile         string
}

func SettingsNew(rootUri, datapath string) Settings {
	var sett Settings
	sett.rootUri = StripSlash(rootUri)
	sett.dataPath = PathConcat(datapath, "/")
	sett.rootNodeOnDisk = PathConcat(sett.dataPath, "meta.rdf")
	sett.idFile = sett.rootNodeOnDisk + ".id"
	return sett
}

func (settings Settings) RootUri() string {
	return settings.rootUri
}

func (settings Settings) IdFile() string {
	return settings.idFile
}
