package web

import "fmt"
import "log"
import "net/http"
import "strings"
import "ldpserver/rdf"
import "ldpserver/ldp"
import "ldpserver/fileio"
import "ldpserver/server"

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
	} else if req.Method == "POST" {
		handlePost(resp, req)
	} else if req.Method == "HEAD" {
		handleGet(false, resp, req)
	} else {
		log.Print("UNK request type")
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
		if err.Error() == "Not found" {
			log.Printf("Not found %s", path)
			http.NotFound(resp, req)
		} else {
			log.Printf("Error %s", err)
			http.Error(resp, "Could not fetch resource", 500)
		}
		return
	}

	for k, v := range node.Headers {
		resp.Header().Add(k, v)
	}
	fmt.Fprint(resp, node.Content())
}

func handlePost(resp http.ResponseWriter, req *http.Request) {
	var node ldp.Node
	var triples string
	var err error
	path := safePath(req.URL.Path)

	if isNonRdfPost(req.Header) {
		// We should pass some hints too
		// (e.g. application type, file name)
		log.Printf("Creating Non-RDF Source")
		node, err = theServer.CreateNonRdfSource(req.Body, path)
	} else {
		log.Printf("Creating RDF Source")
		triples, err = fileio.ReaderToString(req.Body)
		if err != nil {
			http.Error(resp, "Invalid request body received", 400)
			log.Printf(err.Error())
			return
		}
		node, err = theServer.CreateRdfSource(triples, path)
	}

	if err != nil {
		errorMsg := err.Error()
		if errorMsg == ldp.NodeNotFound {
			errorMsg = "Parent container [" + path + "] not found"
			http.Error(resp, errorMsg, 400)
		} else {
			http.Error(resp, errorMsg, 500)
		}
		return
	}

	fmt.Fprint(resp, node.Uri)
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
