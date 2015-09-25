package web

import (
	"fmt"
	"ldpserver/fileio"
	"ldpserver/ldp"
	"ldpserver/rdf"
	"ldpserver/server"
	"ldpserver/util"
	"log"
	"net/http"
	"strings"
)

var theServer server.Server

func Start(address, dataPath string) {
	theServer = server.NewServer("http://"+address, dataPath)
	log.Printf("Listening for requests at %s\n", "http://"+address)
	log.Printf("Data folder: %s\n", dataPath)
	http.HandleFunc("/", homePage)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatal("Failed to start the web server: ", err)
	}
}

func homePage(resp http.ResponseWriter, req *http.Request) {
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
	} else {
		log.Printf("Unknown request type %s", req.Method)
	}
}

func handleGet(includeBody bool, resp http.ResponseWriter, req *http.Request) {
	var node ldp.Node
	var err error

	path := safePath(req.URL.Path)
	if includeBody {
		log.Printf("GET request %s", path)
		node, err = theServer.GetNode(path)
	} else {
		log.Printf("HEAD request %s", path)
		node, err = theServer.GetHead(path)
	}
	if err != nil {
		if err.Error() == ldp.NodeNotFound {
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
	fmt.Fprint(resp, node.Content())
}

func handlePost(resp http.ResponseWriter, req *http.Request) {
	slug := getSlug(req.Header)
	path := safePath(req.URL.Path)
	doPostPut(resp, req, path, slug)
}

func handlePut(resp http.ResponseWriter, req *http.Request) {
	// In PUT requests we don't expect a slug in the headers...
	if getSlug(req.Header) != "" {
		logReqError(req, "Unexpected client provided Slug in PUT request", http.StatusBadRequest)
		http.Error(resp, "Slug is not accepted on PUT requests", http.StatusBadRequest)
		return
	}

	// ...insted we use the last segment of the path as the
	// slug (i.e. the ID of the resource to write.)
	path, slug := util.DirBasePath(safePath(req.URL.Path))
	if path == "." || slug == "." {
		errorMsg := fmt.Sprintf("Invalid Path (%s) or Slug received (%s) in PUT request", path, slug)
		logReqError(req, errorMsg, http.StatusBadRequest)
		http.Error(resp, "Invalid path or Slug indicated on PUT requests", http.StatusBadRequest)
	}

	doPostPut(resp, req, path, slug)
}

func doPostPut(resp http.ResponseWriter, req *http.Request, path string, slug string) {
	var node ldp.Node
	var triples string
	var err error

	if isNonRdfPost(req.Header) {
		// We should pass some hints too
		// (e.g. application type, file name)
		log.Printf("Creating Non-RDF Source at %s", path)
		node, err = theServer.CreateNonRdfSource(req.Body, path, slug)
	} else {
		log.Printf("Creating RDF Source %s at %s", slug, path)
		triples, err = fileio.ReaderToString(req.Body)
		if err != nil {
			logReqError(req, err.Error(), http.StatusBadRequest)
			http.Error(resp, "Invalid request body received", http.StatusBadRequest)
			return
		}
		node, err = theServer.CreateRdfSource(triples, path, slug)
	}

	if err == nil {
		resp.WriteHeader(http.StatusCreated)
	} else {
		errorMsg := err.Error()
		errorCode := http.StatusBadRequest
		if errorMsg == ldp.NodeNotFound {
			errorMsg = "Parent container [" + path + "] not found."
			errorCode = http.StatusNotFound
		} else if errorMsg == ldp.DuplicateNode {
			errorMsg = fmt.Sprintf("Resource already exists. Path: %s Slug: %s", path, slug)
			errorCode = http.StatusConflict
		}
		logReqError(req, errorMsg, errorCode)
		http.Error(resp, errorMsg, errorCode)
		return
	}

	fmt.Fprint(resp, node.Uri())
}

func handlePatch(resp http.ResponseWriter, req *http.Request) {

	path := safePath(req.URL.Path)
	log.Printf("Patching %s", path)

	triples, err := fileio.ReaderToString(req.Body)
	if err != nil {
		http.Error(resp, "Invalid request body received", http.StatusBadRequest)
		log.Printf(err.Error())
		return
	}

	err = theServer.PatchNode(path, triples)
	if err != nil {
		errorMsg := err.Error()
		if errorMsg == ldp.NodeNotFound {
			log.Printf("Not found %s", path)
			http.NotFound(resp, req)
		} else {
			http.Error(resp, errorMsg, http.StatusInternalServerError)
		}
		return
	}

	fmt.Fprint(resp, req.URL.Path)
}

func isNonRdfPost(header http.Header) bool {
	for _, value := range header["Link"] {
		if strings.Contains(value, rdf.LdpNonRdfSourceUri) {
			return true
		}
	}
	return false
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

func logReqError(req *http.Request, message string, code int) {
	log.Printf("Error %d on %s %s: %s", code, req.Method, req.URL.Path, message)
}
