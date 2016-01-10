package server

import (
	"errors"
	"io"
	"ldpserver/ldp"
	"ldpserver/textstore"
	"ldpserver/util"
)

// POST
func (server Server) CreateNonRdfSource(reader io.ReadCloser, parentPath, slug, triples string) (ldp.Node, error) {
	path, err := server.newPathFromSlug(parentPath, slug)
	if err != nil {
		return ldp.Node{}, err
	}

	resource := server.createResource(path)
	err = resource.Error()
	if err != nil && err != textstore.AlreadyExistsError && err != textstore.CreateDeletedError {
		return ldp.Node{}, resource.Error()
	}

	if err == textstore.AlreadyExistsError || err == textstore.CreateDeletedError {
		if slug == "" {
			// We generated a duplicate node.
			return ldp.Node{}, ldp.DuplicateNodeError
		}

		// The user provided slug is duplicated.
		// Let's try with one of our own.
		return server.CreateNonRdfSource(reader, parentPath, "", triples)
	}

	// Create new node
	node, err := ldp.NewNonRdfNode(server.settings, reader, path, triples)
	if err != nil {
		return node, err
	}

	if path != "/" {
		err = server.addNodeToContainer(node, parentPath)
	}

	return node, err
}

// PUT
func (server Server) ReplaceNonRdfSource(reader io.ReadCloser, path, etag, triples string) (ldp.Node, error) {
	if isRootPath(path) {
		return ldp.Node{}, errors.New("Cannot replace root node with an Non-RDF source")
	}

	resource := server.createResource(path)
	if resource.Error() != nil && resource.Error() != textstore.AlreadyExistsError {
		return ldp.Node{}, resource.Error()
	}

	if resource.Error() == textstore.AlreadyExistsError {
		// Replace existing node
		return ldp.ReplaceNonRdfNode(server.settings, reader, path, etag, triples)
	}

	// Create new node
	node, err := ldp.NewNonRdfNode(server.settings, reader, path, triples)
	if err != nil {
		return ldp.Node{}, err
	}

	parentPath := util.ParentUriPath(path)
	err = server.addNodeToContainer(node, parentPath)
	return node, err
}
