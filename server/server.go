package server

import (
	"errors"
	"fmt"
	"ldpserver/ldp"
	"ldpserver/textstore"
	"ldpserver/util"
	// "log"
)

const defaultSlug string = "node"

type Server struct {
	settings ldp.Settings
	minter   chan string
	// this should use an interface so it's not tied to "textStore"
	nextResource chan textstore.Store
}

func NewServer(rootUri string, dataPath string) Server {
	settings := ldp.SettingsNew(rootUri, dataPath)
	var server Server
	server.settings = settings
	server.createIdFile()
	server.minter = CreateMinter(server.settings.IdFile())
	server.nextResource = make(chan textstore.Store)
	server.createRoot()
	return server
}

func (server Server) GetNode(path string, pref ldp.PreferTriples) (ldp.Node, error) {
	return ldp.GetNode(server.settings, path, pref)
}

func (server Server) GetHead(path string) (ldp.Node, error) {
	return ldp.GetHead(server.settings, path)
}

func (server Server) PatchNode(path string, triples string) error {
	node, err := ldp.GetNode(server.settings, path, ldp.PreferTriples{})
	if err != nil {
		return err
	}
	return node.Patch(triples)
}

func (server Server) DeleteNode(path string) error {
	if isRootPath(path) {
		return errors.New("Cannot delete root node")
	}

	node, err := ldp.GetNode(server.settings, path, ldp.PreferTriples{})
	if err != nil {
		return err
	}

	parentPath := util.ParentUriPath(path)
	parent, err := server.getContainer(parentPath)
	if err != nil {
		return err
	}

	// First remove the reference to the node to be deleted...
	err = parent.RemoveContainsUri("<" + node.Uri() + ">")
	if err != nil {
		return err
	}

	// ...then delete the requested node
	return node.Delete()
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

	node, err := ldp.GetNode(server.settings, path, ldp.PreferTriples{})
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
	parentNode, err := ldp.GetNode(server.settings, parentPath, ldp.PreferTriples{})
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
