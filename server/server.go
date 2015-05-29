package server

import "io"
import "errors"
import "ldpserver/ldp"
import "fmt"

type Server struct {
	settings ldp.Settings
	minter   chan string
	nextNode chan ldp.Node
}

// type PlaceholderNode struct {
// 	Node Node
// 	Err  error
// }

func NewServer(rootUri string, dataPath string) Server {
	var server Server
	server.settings = ldp.SettingsNew(rootUri, dataPath)
	ldp.CreateRoot(server.settings)
	server.minter = CreateMinter(server.settings.IdFile())
	server.nextNode = make(chan ldp.Node)
	return server
}

func (server Server) GetNode(path string) (ldp.Node, error) {
	return ldp.GetNode(server.settings, path)
}

func (server Server) GetHead(path string) (ldp.Node, error) {
	return ldp.GetHead(server.settings, path)
}

func (server Server) createNewNode(parentPath string, newPath string) {
	// Create a new palceholder node and put it in the nextNode channel.
	server.nextNode <- ldp.NewPlaceholderNode(server.settings, parentPath, newPath)
}

func (server Server) CreateRdfSource(triples string, parentPath string, slug string) (ldp.Node, error) {
	container, err := server.getContainer(parentPath)
	if err != nil {
		return ldp.Node{}, err
	}

	var newPath string
	if slug == "blog" {
		newPath = MintNextUri(slug, server.minter)
	} else {
		newPath = slug
	}

	// Queue the creation of the new node by way
	// of the server.nextNode channel.
	go server.createNewNode(parentPath, newPath)

	// Pick up the node created from the channel.
	n := <-server.nextNode
	if n.Uri() == "error" {
		err := errors.New(fmt.Sprintf("error creating new node %s", newPath))
		return ldp.Node{}, err
	}

	node, err := ldp.NewRdfNode(server.settings, triples, parentPath, newPath)
	if err != nil {
		return ldp.Node{}, err
	}

	if err := container.AddChild(node); err != nil {
		return ldp.Node{}, err
	}
	return node, nil
}

func (server Server) CreateNonRdfSource(reader io.ReadCloser, parentPath string, slug string) (ldp.Node, error) {
	container, err := server.getContainer(parentPath)
	if err != nil {
		return ldp.Node{}, err
	}

	newPath := MintNextUri(slug, server.minter)
	node, err := ldp.NewNonRdfNode(server.settings, reader, parentPath, newPath)
	if err != nil {
		return node, err
	}

	if err := container.AddChild(node); err != nil {
		return node, err
	}
	return node, nil
}

func (server Server) PatchNode(path string, triples string) error {
	node, err := ldp.GetNode(server.settings, path)
	if err != nil {
		return err
	}
	return node.Patch(triples)
}

func (server Server) getContainer(path string) (ldp.Node, error) {
	if isRootPath(path) {
		// Shortcut. We know for sure this is a container
		return ldp.GetHead(server.settings, "/")
	}

	node, err := ldp.GetNode(server.settings, path)
	if err != nil {
		return node, err
	} else if !node.IsBasicContainer() {
		errorMsg := fmt.Sprintf("%s is not a container", path)
		return node, errors.New(errorMsg)
	}
	return node, nil
}

func (server Server) getContainerUri(parentPath string) (string, error) {
	if isRootPath(parentPath) {
		return server.settings.RootUri(), nil
	}

	// Make sure the parent node exists and it's a container
	parentNode, err := ldp.GetNode(server.settings, parentPath)
	if err != nil {
		return "", err
	} else if !parentNode.IsBasicContainer() {
		return "", errors.New("Parent is not a container")
	}
	return parentNode.Uri(), nil
}

func isRootPath(path string) bool {
	return path == "" || path == "/"
}
