package web

import (
	"log"
	"net/http"
)

func handleOptions(resp http.ResponseWriter, req *http.Request) {
	path := safePath(req.URL.Path)
	node, err := theServer.GetNode(path)
	if err != nil {
		log.Printf("Error %s", err)
		http.Error(resp, "Could not fetch resource", http.StatusInternalServerError)
		return
	}

	setResponseHeaders(resp, node)
}
