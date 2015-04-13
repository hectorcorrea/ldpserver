package web

import "fmt"
import "log"
import "net/http"
import "strings"
import "ldpserver/rdf"
import "ldpserver/ldp"
import "ldpserver/fileio"
import "ldpserver/server"

var sett ldp.Settings
var minter chan string

func Start(address, dataPath string) {
	sett = ldp.SettingsNew(dataPath, "http://"+address)
	ldp.CreateRoot(sett)
	log.Printf("Listening for requests at %s\n", "http://"+address)
	log.Printf("Data folder: %s\n", dataPath)

	minter = ldp.CreateMinter(sett)

	http.HandleFunc("/", homePage)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatal("Failed to start the web server: ", err)
	}
}

func homePage(resp http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		handleGet(sett, true, resp, req)
	} else if req.Method == "POST" {
		handlePost(sett, resp, req)
	} else if req.Method == "HEAD" {
		handleGet(sett, false, resp, req)
	} else {
		log.Print("UNK request type")
	}
}

func handleGet(sett ldp.Settings, includeBody bool, resp http.ResponseWriter, req *http.Request) {
	var node ldp.Node
	var err error
	path := safePath(req.URL.Path)
	log.Printf("GET request %s", path)
	if includeBody {
		node, err = ldp.GetNode(sett, path)
	} else {
		node, err = ldp.GetHead(sett, path)
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

func handlePost(sett ldp.Settings, resp http.ResponseWriter, req *http.Request) {
	var node ldp.Node
	var triples string
	var err error
	path := safePath(req.URL.Path)

	if isNonRdfPost(req.Header) {
		// We should pass some hints too
		// (e.g. application type, file name)
		log.Printf("Creating Non-RDF Source")
		node, err = server.CreateNonRdfSource(sett, req.Body, path, minter)
	} else {
		log.Printf("Creating RDF Source")
		triples, err = fileio.ReaderToString(req.Body)
		if err != nil {
			http.Error(resp, "Invalid request body received", 400)
			log.Printf(err.Error())
			return
		}
		node, err = server.CreateRdfSource(sett, triples, path, minter)
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
