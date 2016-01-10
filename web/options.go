package web

import (
	"net/http"
)

func handleOptions(resp http.ResponseWriter, req *http.Request) {
	path := safePath(req.URL.Path)
	node, err := theServer.GetNode(path)
	if err != nil {
		handleCommonErrors(resp, req, err)
		return
	}
	setResponseHeaders(resp, node)
}
