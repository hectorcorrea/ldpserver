package ldp

import (
	"fmt"
	"ldpserver/fileio"
	"ldpserver/rdf"
	"log"
	"time"
)

func CreateRoot(settings Settings) {
	if fileio.FileExists(settings.rootNodeOnDisk) {
		// nothing to do
		return
	}

	if err := fileio.WriteFile(settings.idFile, "0"); err != nil {
		errorMsg := fmt.Sprintf("Could not create root ID file at %s. %s", settings.idFile, err.Error())
		panic(errorMsg)
	}

	graph := defaultRootRdfGraph(settings.rootUri)
	content := graph.String()
	if err := fileio.WriteFile(settings.rootNodeOnDisk, content); err != nil {
		errorMsg := fmt.Sprintf("Could not create root file at %s.", settings.rootNodeOnDisk)
		panic(errorMsg)
	}
	log.Printf("Root node created on disk at : %s\n", settings.rootNodeOnDisk)
}

func defaultRootRdfGraph(subject string) rdf.RdfGraph {
	// define the triples
	resource := rdf.NewTriple(subject, rdf.RdfTypeUri, rdf.LdpResourceUri)
	rdfSource := rdf.NewTriple(subject, rdf.RdfTypeUri, rdf.LdpRdfSourceUri)
	basicContainer := rdf.NewTriple(subject, rdf.RdfTypeUri, rdf.LdpBasicContainerUri)
	title := rdf.NewTriple(subject, rdf.DcTitleUri, "Root node")
	nowString := time.Now().Format(time.RFC3339)
	created := rdf.NewTriple(subject, rdf.DcCreatedUri, nowString)
	// create the graph
	graph := rdf.RdfGraph{resource, rdfSource, basicContainer, title, created}
	return graph
}
