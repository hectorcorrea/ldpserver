package web

import (
	"ldpserver/rdf"
	"log"
	"net/http"
	"strings"
)

func isNonRdfPost(header http.Header) bool {
	return !isRdfContentType(header)
}

func safePath(rawPath string) string {
	if strings.HasSuffix(rawPath, "/") {
		return rawPath
	}
	return rawPath + "/"
}

func requestSlug(header http.Header) string {
	return headerValue(header, "Slug", "")
}

func requestContentType(header http.Header) string {
	return headerValue(header, "Content-Type", rdf.TurtleContentType)
}

func requestIfNoneMatch(header http.Header) string {
	return headerValue(header, "If-None-Match", "")
}

func requestIfMatch(header http.Header) string {
	return headerValue(header, "If-Match", "")
}

func headerValue(header http.Header, name, defaultValue string) string {
	for _, value := range header[name] {
		return value
	}
	return defaultValue
}

func isRdfContentType(header http.Header) bool {
	contentType := requestContentType(header)
	return strings.HasPrefix(contentType, rdf.TurtleContentType)
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
