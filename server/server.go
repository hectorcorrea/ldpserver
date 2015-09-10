package server

import "fmt"
import "errors"
import "io"
import "ldpserver/ldp"
import "ldpserver/util"
import "ldpserver/textstore"

const defaultSlug string = "node"

type Server struct {
	settings ldp.Settings
	minter   chan string
	nextBag  chan textstore.Store
}

func NewServer(rootUri string, dataPath string) Server {
	var server Server
	server.settings = ldp.SettingsNew(rootUri, dataPath)
	ldp.CreateRoot(server.settings)
	server.minter = CreateMinter(server.settings.IdFile())
	server.nextBag = make(chan textstore.Store)
	return server
}

func (server Server) GetNode(path string) (ldp.Node, error) {
	return ldp.GetNode(server.settings, path)
}

func (server Server) GetHead(path string) (ldp.Node, error) {
	return ldp.GetHead(server.settings, path)
}

func (server Server) getNewPath(slug string) (string, error) {
	if slug == "" {
		// Generate a new server URI (e.g. node34)
		return MintNextUri(defaultSlug, server.minter), nil
	}

	if !util.IsValidSlug(slug) {
		errorMsg := fmt.Sprintf("Invalid Slug received (%s). Slug must not include special characters.", slug)
		return "", errors.New(errorMsg)
	}
	return slug, nil
}

func (server Server) createBag(parentPath string, newPath string) textstore.Store {
	// Queue up the creation of a new bag
	path := util.UriConcat(parentPath, newPath)
	fullPath := util.PathConcat(server.settings.DataPath(), path)
	go func(fullPath string) {
		server.nextBag <- textstore.CreateStore(fullPath)
	}(fullPath)

	// Wait for the new bag to be available.
	bag := <-server.nextBag
	return bag
}

func (server Server) CreateRdfSource(triples string, parentPath string, slug string) (ldp.Node, error) {
	container, err := server.getContainer(parentPath)
	if err != nil {
		return ldp.Node{}, err
	}

	newPath, err := server.getNewPath(slug)
	if err != nil {
		return ldp.Node{}, err
	}

	bag := server.createBag(parentPath, newPath)
	if bag.Error() != nil {
		return ldp.Node{}, bag.Error()
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

	newPath, err := server.getNewPath(slug)
	if err != nil {
		return ldp.Node{}, err
	}

	bag := server.createBag(parentPath, newPath)
	if bag.Error() != nil {
		return ldp.Node{}, bag.Error()
	}

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
