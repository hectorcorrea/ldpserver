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
	msg := err.Error()
	code := http.StatusBadRequest
	path := req.URL.Path
	slug := requestSlug(req.Header)

	switch err {
	case ldp.NodeNotFoundError:
		msg = "Parent container [" + path + "] not found."
		code = http.StatusNotFound
	case ldp.DuplicateNodeError:
		msg = fmt.Sprintf("Resource already exists. Path: %s Slug: %s", path, slug)
		code = http.StatusConflict
	case ldp.EtagMissingError:
		msg = fmt.Sprintf("Etag missing. Path: %s Slug: %s", path, slug)
		code = 428 // precondition required
	case ldp.EtagMismatchError:
		msg = fmt.Sprintf("Etag mismatch. Path: %s Slug: %s", path, slug)
		code = http.StatusPreconditionFailed
	case ldp.ServerManagedPropertyError:
		msg = fmt.Sprintf("Cannot overwrite server-managed property")
		code = http.StatusConflict
		constrainedBy := "<" + req.URL.Path + ">; rel=\"" + rdf.LdpConstrainedBy + "\""
		resp.Header().Add("Link", constrainedBy)
	}

	logReqError(req, msg, code)
	http.Error(resp, msg, code)
}
