package server

import "io"
import "errors"
import "ldpserver/ldp"

// type Server struct {
// 	settings Settings
// }

// func New(settings ldp.Settings) {
// 	return Server{Settings: settings}
// }

func GetNode(settings ldp.Settings, path string) (ldp.Node, error) {
	return ldp.GetNode(settings, path)
}

func GetHead(settings ldp.Settings, path string) (ldp.Node, error) {
	return ldp.GetHead(settings, path)
}

func CreateRdfSource(settings ldp.Settings, triples string, parentPath string) (ldp.Node, error) {
	parentUri, err := getContainerUri(settings, parentPath)
	if err != nil {
		return ldp.Node{}, err
	}

	node, err := ldp.NewRdfNode(settings, triples, parentUri)
	if err != nil {
		return node, err
	}

	ldp.AddChildToContainer(settings, node.Uri, parentUri)
	return node, nil
}

func CreateNonRdfSource(settings ldp.Settings, reader io.ReadCloser, parentPath string) (ldp.Node, error) {
	parentUri, err := getContainerUri(settings, parentPath)
	if err != nil {
		return ldp.Node{}, err
	}

	node, err := ldp.NewNonRdfNode(settings, reader, parentUri)
	if err != nil {
		return node, err
	}
	ldp.AddChildToContainer(settings, node.Uri, parentUri)
	return node, nil
}

func PatchNode(settings ldp.Settings, path string, triples string) error {
	node, err := ldp.GetNode(settings, path)
	if err != nil {
		return err
	}
	return node.Patch(settings, triples)
}

func getContainerUri(settings ldp.Settings, parentPath string) (string, error) {
	if parentPath == "" || parentPath == "/" {
		return settings.RootUrl(), nil
	}

	// Make sure the parent node exists and it's a container
	parentNode, err := ldp.GetNode(settings, parentPath)
	if err != nil {
		return "", err
	} else if !parentNode.IsBasicContainer() {
		return "", errors.New("Parent is not a container")
	}
	return parentNode.Uri, nil
}
