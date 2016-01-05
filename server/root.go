package server

import (
	"fmt"
	"ldpserver/ldp"
	"log"
)

func (server Server) createRoot() {
	_, err := server.GetHead("/")
	if err == nil {
		return
	}

	if err != ldp.NodeNotFoundError {
		panic(fmt.Sprintf("Error reading root node: %s", err.Error()))
	}

	_, err = server.CreateRdfSource("", ".", ".")
	if err != nil {
		panic(fmt.Sprintf("Could not create root node: %s", err.Error()))
	}

	log.Printf("Root node created on disk at : %s\n", server.settings.DataPath())
}
