package web

import (
	"fmt"
	"ldpserver/fileio"
	"log"
	"net/http"
)

func handlePatch(resp http.ResponseWriter, req *http.Request) {
	if !isRdfRequest(req.Header) {
		errorMsg := fmt.Sprintf("Invalid Content-Type (%s) received", requestContentType(req.Header))
		logReqError(req, errorMsg, http.StatusBadRequest)
		http.Error(resp, errorMsg, http.StatusBadRequest)
		return
	}

	path := safePath(req.URL.Path)
	log.Printf("Patching %s", path)

	triples, err := fileio.ReaderToString(req.Body)
	if err != nil {
		errorMsg := fmt.Sprintf("Invalid body received. Error: %s", err.Error())
		logReqError(req, errorMsg, http.StatusBadRequest)
		http.Error(resp, errorMsg, http.StatusBadRequest)
		return
	}

	err = theServer.PatchNode(path, triples)
	if err != nil {
		handleCommonErrors(resp, req, err)
		return
	}

	fmt.Fprint(resp, req.URL.Path)
}
