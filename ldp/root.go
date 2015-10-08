package ldp

import (
	"fmt"
	"ldpserver/textstore"
	"log"
)

func CreateRoot(settings Settings) {
	if textstore.Exists(settings.dataPath) {
		// nothing to do
		return
	}

	store := textstore.CreateStore(settings.dataPath)
	if store.Error() != nil {
		errorMsg := fmt.Sprintf("Could not create root store: %s", store.Error())
		panic(errorMsg)
	}

	if err := store.SaveFile("meta.rdf.id", "0"); err != nil {
		errorMsg := fmt.Sprintf("Could not create root ID file: %s", err.Error())
		panic(errorMsg)
	}

	graph := DefaultGraph(settings.rootUri)
	content := graph.String()
	if err := store.SaveFile("meta.rdf", content); err != nil {
		errorMsg := fmt.Sprintf("Could not create root file at %s.", err.Error())
		panic(errorMsg)
	}

	log.Printf("Root node created on disk at : %s\n", settings.dataPath)
}
