package ldp

import "ldpserver/fileio"
import "strconv"

func CreateMinter(settings Settings) chan string {
	nextId := make(chan string)
	go func(settings Settings) {
		for {
			nextId <- mintNextId(settings)
		}
	}(settings)
	return nextId
}

func MintNextUri(slug string, minter chan string) string {
	nextId := <-minter
	return slug + nextId
}

func mintNextId(settings Settings) string {
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
