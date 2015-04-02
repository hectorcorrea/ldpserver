package ldp

import "ldpserver/fileio"
import "strconv"

func mintNextId(settings Settings) string {
	// TODO: handle concurrency if more than one call
	// come at the same time
	lastText, err := fileio.ReadFile(settings.rootNodeOnDisk + ".id")
	if err != nil {
		panic("Could not read last id")
	}

	lastId, err := strconv.ParseInt(lastText, 10, 0)
	if err != nil {
		panic("Could not calculate last id")
	}

	nextId := strconv.Itoa(int(lastId + 1))
	err = fileio.WriteFile(settings.rootNodeOnDisk+".id", nextId)
	if err != nil {
		panic("Error writting next id")
	}
	return nextId
}

func MintNextUri(settings Settings, slug string) string {
	return slug + mintNextId(settings)
}
