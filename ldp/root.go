package ldp

import (
	"fmt"
	"ldpserver/textstore"
	"log"
)

func CreateRoot(settings Settings) {
	_, err := GetHead(settings, "/")
	if err == nil {
		return
	}

	if err != NodeNotFoundError {
		panic(fmt.Sprintf("Error reading root node: %s", err.Error()))
	}

	_, err = NewRdfNode(settings, "", "/")
	if err != nil {
		panic(fmt.Sprintf("Could not create root node: %s", err.Error()))
	}

	log.Printf("Root node created on disk at : %s\n", settings.dataPath)
	createRootIdFile(settings.dataPath)
}

func createRootIdFile(path string) {
	// TODO: This code should not depend on the textstore
	store := textstore.NewStore(path)
	if err := store.SaveFile("meta.rdf.id", "0"); err != nil {
		panic(fmt.Sprintf("Could not create root ID file: %s", err.Error()))
	}
}
