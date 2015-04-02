package ldp

import "ldpserver/fileio"
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

	content := defaultRootRdfGraph(settings.rootUrl).String()
	if err := fileio.WriteFile(settings.rootNodeOnDisk, content); err != nil {
		panic("Could not create root file at " + settings.rootNodeOnDisk)
	}
	log.Printf("Root node created on disk at : %s\n", settings.rootNodeOnDisk)
}

func defaultRootRdfGraph(subject string) RdfGraph {
	// define the triples
	resource := NewTriple(subject, RdfTypeUri, LdpResourceUri)
	rdfSource := NewTriple(subject, RdfTypeUri, LdpRdfSourceUri)
	basicContainer := NewTriple(subject, RdfTypeUri, LdpBasicContainerUri)
	title := NewTriple(subject, DcTitleUri, "Root node")
	nowString := time.Now().Format(time.RFC3339)
	created := NewTriple(subject, DcCreatedUri, nowString)
	// create the graph
	graph := RdfGraph{resource, rdfSource, basicContainer, title, created}
	return graph
}
