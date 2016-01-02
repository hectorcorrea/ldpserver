package server

import (
	"errors"
	"fmt"
	"io"
	"ldpserver/ldp"
	"ldpserver/textstore"
	"ldpserver/util"
	// "log"
)

const defaultSlug string = "node"

type Server struct {
	settings ldp.Settings
	minter   chan string
	// this should use an interface to it's not tied to "textStore"
	nextResource chan textstore.Store
}

func NewServer(rootUri string, dataPath string) Server {
	var server Server
	server.settings = ldp.SettingsNew(rootUri, dataPath)
	ldp.CreateRoot(server.settings)
	server.minter = CreateMinter(server.settings.IdFile())
	server.nextResource = make(chan textstore.Store)
	return server
}

func (server Server) GetNode(path string) (ldp.Node, error) {
	return ldp.GetNode(server.settings, path)
}

func (server Server) GetHead(path string) (ldp.Node, error) {
	return ldp.GetHead(server.settings, path)
}

// PUT
func (server Server) ReplaceRdfSource(triples string, parentPath string, slug string, etag string) (ldp.Node, error) {
	path, err := server.newPathFromSlug(parentPath, slug)
	if err != nil {
		return ldp.Node{}, err
	}

	resource := server.createResource(path)
	if resource.Error() != nil && resource.Error() != textstore.AlreadyExistsError {
		return ldp.Node{}, resource.Error()
	}

	if resource.Error() == textstore.AlreadyExistsError {
		// Replace existing node
		return ldp.ReplaceRdfNode(server.settings, triples, path, etag)
	}

	// Create new node
	node, err := ldp.NewRdfNode(server.settings, triples, path)
	if err != nil {
		return ldp.Node{}, err
	}

	if path != "/" {
		err = server.addNodeToContainer(node, parentPath)
	}

	return node, err
}

// POST
func (server Server) CreateRdfSource(triples string, parentPath string, slug string) (ldp.Node, error) {
	path, err := server.newPathFromSlug(parentPath, slug)
	if err != nil {
		return ldp.Node{}, err
	}

	resource := server.createResource(path)
	if resource.Error() != nil && resource.Error() != textstore.AlreadyExistsError {
		return ldp.Node{}, resource.Error()
	}

	if resource.Error() == textstore.AlreadyExistsError {
		if slug == "" {
			// We generated a duplicate node.
			return ldp.Node{}, ldp.DuplicateNodeError
		}

		// The user provided slug is duplicated.
		// Let's try with one of our own.
		return server.CreateRdfSource(triples, parentPath, "")
	}

	// Create new node
	node, err := ldp.NewRdfNode(server.settings, triples, path)
	if err != nil {
		return ldp.Node{}, err
	}

	if path != "/" {
		err = server.addNodeToContainer(node, parentPath)
	}

	return node, err
}

// PUT
// TODO: this should receive a full path
func (server Server) ReplaceNonRdfSource(reader io.ReadCloser, parentPath string, path000 string, etag string) (ldp.Node, error) {
	if path000 == "" {
		return ldp.Node{}, errors.New("Cannot replace without a path")
	}

	path := util.UriConcat(parentPath, path000)
	resource := server.createResource(path)
	if resource.Error() != nil && resource.Error() != textstore.AlreadyExistsError {
		return ldp.Node{}, resource.Error()
	}

	if resource.Error() == textstore.AlreadyExistsError {
		// Replace existing node
		return ldp.ReplaceNonRdfNode(server.settings, reader, path, etag)
	}

	// Create new node
	node, err := ldp.NewNonRdfNode(server.settings, reader, path)
	if err != nil {
		return ldp.Node{}, err
	}

	if path != "/" {
		err = server.addNodeToContainer(node, parentPath)
	}

	return node, err
}

// POST
func (server Server) CreateNonRdfSource(reader io.ReadCloser, parentPath string, slug string) (ldp.Node, error) {
	path, err := server.newPathFromSlug(parentPath, slug)
	if err != nil {
		return ldp.Node{}, err
	}

	resource := server.createResource(path)
	if resource.Error() != nil && resource.Error() != textstore.AlreadyExistsError {
		return ldp.Node{}, resource.Error()
	}

	if resource.Error() == textstore.AlreadyExistsError {
		if slug == "" {
			// We generated a duplicate node.
			return ldp.Node{}, ldp.DuplicateNodeError
		}

		// The user provided slug is duplicated.
		// Let's try with one of our own.
		return server.CreateNonRdfSource(reader, parentPath, "")
	}

	// Create new node
	node, err := ldp.NewNonRdfNode(server.settings, reader, path)
	if err != nil {
		return node, err
	}

	if path != "/" {
		err = server.addNodeToContainer(node, parentPath)
	}

	return node, err
}

func (server Server) PatchNode(path string, triples string) error {
	node, err := ldp.GetNode(server.settings, path)
	if err != nil {
		return err
	}
	return node.Patch(triples)
}

func (server Server) addNodeToContainer(node ldp.Node, path string) error {
	container, err := server.getContainer(path)
	if err != nil {
		return err
	}
	return container.AddChild(node)
}

func (server Server) newPathFromSlug(parentPath string, slug string) (string, error) {
	isRootNode := (parentPath == ".") && (slug == ".")
	if isRootNode {
		return "/", nil // special case
	}

	if slug == "" {
		// Generate a new server URI (e.g. node34)
		slug = MintNextUri(defaultSlug, server.minter)
	}

	if !util.IsValidSlug(slug) {
		return "", fmt.Errorf("Invalid Slug (%s)", slug)
	}
	return util.UriConcat(parentPath, slug), nil
}

func (server Server) createResource(path string) textstore.Store {
	pathOnDisk := util.PathConcat(server.settings.DataPath(), path)
	// Queue up the creation of a new resource
	go func(pathOnDisk string) {
		server.nextResource <- textstore.CreateStore(pathOnDisk)
	}(pathOnDisk)

	// Wait for the new resource to be available.
	resource := <-server.nextResource
	return resource
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
