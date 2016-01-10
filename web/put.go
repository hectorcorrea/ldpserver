package web

import (
	"errors"
	"ldpserver/fileio"
	"ldpserver/ldp"
	"ldpserver/util"
	"log"
	"net/http"
)

func handlePut(resp http.ResponseWriter, req *http.Request) {
	logHeaders(req)

	node, err := doPut(resp, req)
	if err != nil {
		handlePostPutError(resp, req, err)
		return
	}

	handlePostPutSuccess(resp, node)
}

func doPut(resp http.ResponseWriter, req *http.Request) (ldp.Node, error) {
	if requestSlug(req.Header) != "" {
		return ldp.Node{}, errors.New("Slug is not accepted on PUT requests")
	}

	etag := requestIfMatch(req.Header)

	if isNonRdfRequest(req.Header) {
		path := req.URL.Path
		log.Printf("Creating Non-RDF Source at %s", path)
		triples := defaultNonRdfTriples(req.Header)
		return theServer.ReplaceNonRdfSource(req.Body, path, etag, triples)
	}

	path, slug := util.DirBasePath(safePath(req.URL.Path))
	log.Printf("Creating RDF Source %s at %s", slug, path)
	triples, err := fileio.ReaderToString(req.Body)
	if err != nil {
		return ldp.Node{}, errors.New("Invalid request body received")
	}
	return theServer.ReplaceRdfSource(triples, path, slug, etag)
}
