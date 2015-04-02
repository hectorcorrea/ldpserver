package ldp

type Settings struct {
	dataPath       string
	rootUrl        string
	rootNodeOnDisk string
}

func SettingsNew(datapath, rootUrl string) Settings {
	var sett Settings
	sett.dataPath = PathConcat(datapath, "/")
	sett.rootUrl = StripSlash(rootUrl)
	sett.rootNodeOnDisk = PathConcat(sett.dataPath, "meta.rdf")
	return sett
}

func (settings Settings) RootUrl() string {
	return settings.rootUrl
}
