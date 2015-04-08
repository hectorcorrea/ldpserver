package ldp

import "ldpserver/fileio"
import "ldpserver/rdf"
import "log"
import "time"

func CreateRoot(settings Settings) {
	if fileio.FileExists(settings.rootNodeOnDisk) {
		// nothing to do
		return
	}

	if err := fileio.WriteFile(settings.rootNodeOnDisk+".id", "0"); err != nil {
		panic("Could not create root ID file at " + settings.rootNodeOnDisk + ".id " + err.Error())
	}

	content := defaultRootRdfGraph(settings.rootUri).String()
	if err := fileio.WriteFile(settings.rootNodeOnDisk, content); err != nil {
		panic("Could not create root file at " + settings.rootNodeOnDisk)
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
