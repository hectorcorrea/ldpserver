package server

import "fmt"
import "strconv"
import "ldpserver/fileio"

// TODO: Move the handling of the IdFile to its own class.
func (server Server) createIdFile() {
	idFile := server.settings.IdFile()
	if fileio.FileExists(idFile) {
		return
	}

	err := fileio.CreateFile(idFile, "0")
	if err != nil {
		panic(fmt.Sprintf("Could not create ID file: %s", err.Error()))
	}
}

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
		errorMsg := fmt.Sprintf("Could not read last id from [%s]. Error: %s", idFile, err)
		panic(errorMsg)
	}

	lastId, err := strconv.ParseInt(lastText, 10, 0)
	if err != nil {
		errorMsg := fmt.Sprintf("Could not calculate last id from [%s]", idFile)
		panic(errorMsg)
	}

	nextId := strconv.Itoa(int(lastId + 1))
	err = fileio.WriteFile(idFile, nextId)
	if err != nil {
		errorMsg := fmt.Sprintf("Error writing next id to [%s]", idFile)
		panic(errorMsg)
	}
	return nextId
}
