package web

import (
	"fmt"
	"ldpserver/ldp"
	"log"
	"net/http"
)

func handleGet(includeBody bool, resp http.ResponseWriter, req *http.Request) {
	var node ldp.Node
	var err error

	logHeaders(req)
	path := safePath(req.URL.Path)
	if includeBody {
		log.Printf("GET request %s", path)
		node, err = theServer.GetNode(path)
	} else {
		log.Printf("HEAD request %s", path)
		node, err = theServer.GetHead(path)
	}
	if err != nil {
		if err == ldp.NodeNotFoundError {
			log.Printf("Not found %s", path)
			http.NotFound(resp, req)
		} else {
			log.Printf("Error %s", err)
			http.Error(resp, "Could not fetch resource", http.StatusInternalServerError)
		}
		return
	}

	setResponseHeaders(resp, node)

	if etag := requestIfNoneMatch(req.Header); etag != "" {
		if etag == node.Etag() {
			resp.WriteHeader(http.StatusNotModified)
			return
		}
	}

	fmt.Fprint(resp, node.Content())
}
