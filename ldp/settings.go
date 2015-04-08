package ldp

type Settings struct {
	dataPath       string
	rootUri        string
	rootNodeOnDisk string
}

func SettingsNew(datapath, rootUri string) Settings {
	var sett Settings
	sett.dataPath = PathConcat(datapath, "/")
	sett.rootUri = StripSlash(rootUri)
	sett.rootNodeOnDisk = PathConcat(sett.dataPath, "meta.rdf")
	return sett
}

func (settings Settings) RootUri() string {
	return settings.rootUri
}
