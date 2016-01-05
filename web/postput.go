package web

import (
	"fmt"
	"ldpserver/ldp"
	"ldpserver/rdf"
	"log"
	"net/http"
)

func handlePostPutSuccess(resp http.ResponseWriter, node ldp.Node) {
	resp.Header().Add("Location", node.Uri())
	resp.WriteHeader(http.StatusCreated)
	log.Printf("Resource created at %s", node.Uri())
	fmt.Fprint(resp, node.Uri())
}

func handlePostPutError(resp http.ResponseWriter, req *http.Request, err error) {
	errorMsg := err.Error()
	errorCode := http.StatusBadRequest
	path := req.URL.Path
	slug := requestSlug(req.Header)
	if err == ldp.NodeNotFoundError {
		errorMsg = "Parent container [" + path + "] not found."
		errorCode = http.StatusNotFound
	} else if err == ldp.DuplicateNodeError {
		errorMsg = fmt.Sprintf("Resource already exists. Path: %s Slug: %s", path, slug)
		errorCode = http.StatusConflict
	} else if err == ldp.EtagMissingError {
		errorMsg = fmt.Sprintf("Etag missing. Path: %s Slug: %s", path, slug)
		errorCode = 428 // precondition required
	} else if err == ldp.EtagMismatchError {
		errorMsg = fmt.Sprintf("Etag mismatch. Path: %s Slug: %s", path, slug)
		errorCode = http.StatusPreconditionFailed
	} else if err == ldp.ServerManagedPropertyError {
		errorMsg = fmt.Sprintf("Cannot overwrite server-managed property")
		errorCode = http.StatusConflict
		constrainedBy := "<" + req.URL.Path + ">; rel=\"" + rdf.LdpConstrainedBy + "\""
		resp.Header().Add("Link", constrainedBy)
	}
	logReqError(req, errorMsg, errorCode)
	http.Error(resp, errorMsg, errorCode)
}
