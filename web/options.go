package web

import (
	"ldpserver/ldp"
	"net/http"
)

func handleOptions(resp http.ResponseWriter, req *http.Request) {
	path := safePath(req.URL.Path)
	node, err := theServer.GetNode(path, ldp.PreferTriples{})
	if err != nil {
		handleCommonErrors(resp, req, err)
		return
	}
	setResponseHeaders(resp, node)
}
