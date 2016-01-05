package server

import (
	"ldpserver/ldp"
	"ldpserver/textstore"
)

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
