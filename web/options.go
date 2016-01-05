package web

import (
	"log"
	"net/http"
)

func handleOptions(resp http.ResponseWriter, req *http.Request) {
	logHeaders(req)
	path := safePath(req.URL.Path)
	node, err := theServer.GetNode(path)
	if err != nil {
		log.Printf("Error %s", err)
		http.Error(resp, "Could not fetch resource", http.StatusInternalServerError)
		return
	}

	for key, header := range node.Headers() {
		for _, value := range header {
			resp.Header().Add(key, value)
		}
	}
}
