package server

import "ldpserver/fileio"
import "strconv"

func CreateMinter(idFile string) chan string {
	nextId := make(chan string)
	go func(idFile string) {
		for {
			nextId <- mintNextId(idFile)
		}
	}(idFile)
	return nextId
}

// Uses a synchronous channel to force sequential process
// of this code.
func MintNextUri(slug string, minter chan string) string {
	nextId := <-minter
	return slug + nextId
}

func mintNextId(idFile string) string {
	lastText, err := fileio.ReadFile(idFile)
	if err != nil {
		panic("Could not read last id")
	}

	lastId, err := strconv.ParseInt(lastText, 10, 0)
	if err != nil {
		panic("Could not calculate last id")
	}

	nextId := strconv.Itoa(int(lastId + 1))
	err = fileio.WriteFile(idFile, nextId)
	if err != nil {
		panic("Error writting next id")
	}
	return nextId
}
