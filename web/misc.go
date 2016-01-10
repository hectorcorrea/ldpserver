package web

import (
	"ldpserver/ldp"
	"ldpserver/rdf"
	"log"
	"net/http"
	"strings"
)

func handleCommonErrors(resp http.ResponseWriter, req *http.Request, err error) {
	if err == nil {
		panic("No error to handle")
	}

	if err == ldp.NodeNotFoundError {
		log.Printf("Not found %s", req.URL.Path)
		http.NotFound(resp, req)
		return
	}

	log.Printf("Error %s", err)
	http.Error(resp, "Error processing request", http.StatusInternalServerError)
}

func isRdfRequest(header http.Header) bool {
	contentType := requestContentType(header)
	if contentType == "" {
		return true
	}
	return strings.HasPrefix(contentType, rdf.TurtleContentType)
}

func isNonRdfRequest(header http.Header) bool {
	return !isRdfRequest(header)
}

func safePath(rawPath string) string {
	if strings.HasSuffix(rawPath, "/") {
		return rawPath
	}
	return rawPath + "/"
}

func setResponseHeaders(resp http.ResponseWriter, node ldp.Node) {
	for key, header := range node.Headers() {
		for _, value := range header {
			resp.Header().Add(key, value)
		}
	}
}

func setResponseHeadersMetadataOnly(resp http.ResponseWriter, node ldp.Node) {
	resp.Header().Add("Content-Type", rdf.TurtleContentType)
	resp.Header().Add("Allow", "GET")
	resp.Header().Add("Allow", "HEAD")
	resp.Header().Add("Etag", node.Etag())
}

func requestSlug(header http.Header) string {
	return headerValue(header, "Slug")
}

func requestContentType(header http.Header) string {
	value := headerValue(header, "Content-Type")
	// TODO: remove this horrible hack
	if strings.HasSuffix(value, "; charset=ISO-8859-1") {
		return strings.Replace(value, "; charset=ISO-8859-1", "", 1)
	}
	return value
}

func requestIfNoneMatch(header http.Header) string {
	return headerValue(header, "If-None-Match")
}

func requestIfMatch(header http.Header) string {
	return headerValue(header, "If-Match")
}

func headerValue(header http.Header, name string) string {
	for _, value := range header[name] {
		return value
	}
	return ""
}

func isNonRdfMetadataOnlyRequest(req *http.Request) bool {
	return req.URL.Query().Get("metadata") == "yes"
}

func defaultNonRdfTriples(header http.Header) string {
	triples := ""
	contentType := requestContentType(header)
	if contentType != "" {
		triples = "<> <" + rdf.ServerContentTypeUri + "> \"" + contentType + "\" ."
	}
	// TODO: We should also try to read the file name from the header (if available)
	return triples
}

func logHeaders(req *http.Request) {
	log.Printf("==> HTTP Headers %s %s", req.Method, req.URL.Path)
	for header, values := range req.Header {
		for _, value := range values {
			log.Printf("\t\t %s %s", header, value)
		}
	}
}

func logReqError(req *http.Request, message string, code int) {
	log.Printf("Error %d on %s %s: %s", code, req.Method, req.URL.Path, message)
}
