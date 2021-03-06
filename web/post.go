package web

import (
	"ldpserver/fileio"
	"ldpserver/ldp"
	"log"
	"net/http"
)

func handlePost(resp http.ResponseWriter, req *http.Request) {
	slug := requestSlug(req.Header)
	path := safePath(req.URL.Path)
	node, err := doPost(resp, req, path, slug)
	if err != nil {
		handlePostPutError(resp, req, err)
		return
	}

	handlePostPutSuccess(resp, node)
}

func doPost(resp http.ResponseWriter, req *http.Request, path string, slug string) (ldp.Node, error) {
	if isNonRdfRequest(req.Header) {
		log.Printf("Creating Non-RDF Source at %s", path)
		triples := defaultNonRdfTriples(req.Header)
		return theServer.CreateNonRdfSource(req.Body, path, slug, triples)
	}

	log.Printf("Creating RDF Source %s at %s", slug, path)
	triples, err := fileio.ReaderToString(req.Body)
	if err != nil {
		return ldp.Node{}, err
	}
	return theServer.CreateRdfSource(triples, path, slug)
}
