package ldp

import (
	"fmt"
	"ldpserver/rdf"
	"ldpserver/textstore"
	"log"
	"time"
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

	graph := defaultRootRdfGraph(settings.rootUri)
	content := graph.String()
	if err := store.SaveFile("meta.rdf", content); err != nil {
		errorMsg := fmt.Sprintf("Could not create root file at %s.", err.Error())
		panic(errorMsg)
	}

	log.Printf("Root node created on disk at : %s\n", settings.dataPath)
}

func defaultRootRdfGraph(uri string) rdf.RdfGraph {
	subject := "<" + uri + ">"
	// define the triples
	resource := rdf.NewTriple(subject, "<"+rdf.RdfTypeUri+">", "<"+rdf.LdpResourceUri+">")
	rdfSource := rdf.NewTriple(subject, "<"+rdf.RdfTypeUri+">", "<"+rdf.LdpRdfSourceUri+">")
	basicContainer := rdf.NewTriple(subject, "<"+rdf.RdfTypeUri+">", "<"+rdf.LdpBasicContainerUri+">")
	title := rdf.NewTriple(subject, "<"+rdf.DcTitleUri+">", "\"Root node\"")
	nowString := "\"" + time.Now().Format(time.RFC3339) + "\""
	created := rdf.NewTriple(subject, "<"+rdf.DcCreatedUri+">", nowString)

	// create the graph
	graph := rdf.RdfGraph{resource, rdfSource, basicContainer, title, created}
	return graph
}
