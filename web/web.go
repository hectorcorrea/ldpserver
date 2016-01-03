package web

import (
	"bufio"
	"errors"
	"fmt"
	"ldpserver/fileio"
	"ldpserver/ldp"
	"ldpserver/rdf"
	"ldpserver/server"
	"ldpserver/util"
	"log"
	"net/http"
	"os"
	"strings"
)

var stdin *bufio.Reader
var theServer server.Server

func Start(address, dataPath string) {
	theServer = server.NewServer("http://"+address, dataPath)
	stdin = bufio.NewReader(os.Stdin)
	log.Printf("Listening for requests at %s\n", "http://"+address)
	log.Printf("Data folder: %s\n", dataPath)
	http.HandleFunc("/", homePage)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatal("Failed to start the web server: ", err)
	}
}

func homePage(resp http.ResponseWriter, req *http.Request) {
	readline()
	if req.Method == "GET" {
		handleGet(true, resp, req)
	} else if req.Method == "HEAD" {
		handleGet(false, resp, req)
	} else if req.Method == "POST" {
		handlePost(resp, req)
	} else if req.Method == "PUT" {
		handlePut(resp, req)
	} else if req.Method == "PATCH" {
		handlePatch(resp, req)
	} else if req.Method == "OPTIONS" {
		handleOptions(resp, req)
	} else {
		log.Printf("Unknown request type %s", req.Method)
	}
}

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

	for key, header := range node.Headers() {
		for _, value := range header {
			resp.Header().Add(key, value)
		}
	}

	if etag := requestIfNoneMatch(req.Header); etag != "" {
		if etag == node.Etag() {
			resp.WriteHeader(http.StatusNotModified)
			return
		}
	}

	fmt.Fprint(resp, node.Content())
}

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

func handlePatch(resp http.ResponseWriter, req *http.Request) {
	if !isRdfContentType(req.Header) {
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
		errorMsg := err.Error()
		if err == ldp.NodeNotFoundError {
			logReqError(req, errorMsg, http.StatusNotFound)
			http.NotFound(resp, req)
		} else {
			logReqError(req, errorMsg, http.StatusInternalServerError)
			http.Error(resp, errorMsg, http.StatusInternalServerError)
		}
		return
	}

	fmt.Fprint(resp, req.URL.Path)
}

func handlePost(resp http.ResponseWriter, req *http.Request) {
	logHeaders(req)

	slug := getSlug(req.Header)
	path := safePath(req.URL.Path)
	node, err := doPost(resp, req, path, slug)
	if err != nil {
		handlePostPutError(resp, req, err)
		return
	}

	handlePostPutSuccess(resp, node)
}

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
	slug := getSlug(req.Header)
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
	}
	logReqError(req, errorMsg, errorCode)
	http.Error(resp, errorMsg, errorCode)
}

func handlePut(resp http.ResponseWriter, req *http.Request) {
	logHeaders(req)

	node, err := doPut(resp, req)
	if err != nil {
		handlePostPutError(resp, req, err)
		return
	}

	handlePostPutSuccess(resp, node)
}

func doPost(resp http.ResponseWriter, req *http.Request, path string, slug string) (ldp.Node, error) {
	if isNonRdfPost(req.Header) {
		// We should pass some hints too
		// (e.g. application type, file name)
		log.Printf("Creating Non-RDF Source at %s", path)
		return theServer.CreateNonRdfSource(req.Body, path, slug)
	}

	log.Printf("Creating RDF Source %s at %s", slug, path)
	triples, err := fileio.ReaderToString(req.Body)
	if err != nil {
		return ldp.Node{}, err
	}
	return theServer.CreateRdfSource(triples, path, slug)
}

func doPut(resp http.ResponseWriter, req *http.Request) (ldp.Node, error) {
	if getSlug(req.Header) != "" {
		return ldp.Node{}, errors.New("Slug is not accepted on PUT requests")
	}

	etag := requestIfMatch(req.Header)

	if isNonRdfPost(req.Header) {
		// We should pass some hints too
		// (e.g. application type, file name)
		path := req.URL.Path
		log.Printf("Creating Non-RDF Source at %s", path)
		return theServer.ReplaceNonRdfSource(req.Body, path, etag)
	}

	path, slug := util.DirBasePath(safePath(req.URL.Path))
	log.Printf("Creating RDF Source %s at %s", slug, path)
	triples, err := fileio.ReaderToString(req.Body)
	if err != nil {
		return ldp.Node{}, errors.New("Invalid request body received")
	}
	return theServer.ReplaceRdfSource(triples, path, slug, etag)
}

func isNonRdfPost(header http.Header) bool {
	return !isRdfContentType(header)
}

func safePath(rawPath string) string {
	if strings.HasSuffix(rawPath, "/") {
		return rawPath
	}
	return rawPath + "/"
}

func getSlug(header http.Header) string {
	for _, value := range header["Slug"] {
		return value
	}
	return ""
}

func requestContentType(header http.Header) string {
	for _, value := range header["Content-Type"] {
		return value
	}
	return rdf.TurtleContentType
}

func requestIfNoneMatch(header http.Header) string {
	for _, value := range header["If-None-Match"] {
		return value
	}
	return ""
}

func requestIfMatch(header http.Header) string {
	for _, value := range header["If-Match"] {
		return value
	}
	return ""
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

func readline() {
	return
	log.Print("Hit [ENTER]")
	stdin.ReadString('\n')
}
