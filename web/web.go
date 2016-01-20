package web

import (
	"ldpserver/server"
	"log"
	"net/http"
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
	logHeaders(req)
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
	} else if req.Method == "xxDELETE" {
		handleDelete(resp, req)
	} else {
		log.Printf("Unknown request type %s", req.Method)
	}
}
