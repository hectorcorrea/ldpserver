package ldp

import (
	"fmt"
	"ldpserver/fileio"
	"ldpserver/rdf"
	"ldpserver/bagit"
	"log"
	"time"
)

func CreateRoot(settings Settings) {
	if fileio.FileExists(settings.rootBagOnDisk) {
		log.Printf("Root node already exists")
		// nothing to do
		return
	}

	bag, err := bagit.CreateBag(settings.dataPath)
	if err != nil {
		errorMsg := fmt.Sprintf("Could not create root bag: %s", err.Error())
		panic(errorMsg)
	}

	if err := bag.SaveFile("meta.rdf.id", "0"); err != nil {
		errorMsg := fmt.Sprintf("Could not create root ID file: %s", err.Error())
		panic(errorMsg)
	}

	graph := defaultRootRdfGraph(settings.rootUri)
	content := graph.String()
	if err := bag.SaveFile("meta.rdf", content); err != nil {
		errorMsg := fmt.Sprintf("Could not create root file at %s.", settings.rootNodeOnDisk)
		panic(errorMsg)
	}

	log.Printf("Root node created on disk at : %s\n", settings.dataPath)
}

func defaultRootRdfGraph(subject string) rdf.RdfGraph {
	// define the triples
	resource := rdf.NewTripleUri(subject, rdf.RdfTypeUri, rdf.LdpResourceUri)
	rdfSource := rdf.NewTripleUri(subject, rdf.RdfTypeUri, rdf.LdpRdfSourceUri)
	basicContainer := rdf.NewTripleUri(subject, rdf.RdfTypeUri, rdf.LdpBasicContainerUri)
	title := rdf.NewTripleLit(subject, rdf.DcTitleUri, "Root node")
	nowString := time.Now().Format(time.RFC3339)
	created := rdf.NewTripleLit(subject, rdf.DcCreatedUri, nowString)

	// create the graph
	graph := rdf.RdfGraph{resource, rdfSource, basicContainer, title, created}
	return graph
}
