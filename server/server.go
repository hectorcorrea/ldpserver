package server

import "io"
import "errors"
import "ldpserver/ldp"
import "fmt"

func NewServer(rootUri, dataPath string) (ldp.Settings, chan string) {
	sett := ldp.SettingsNew(rootUri, dataPath)
	ldp.CreateRoot(sett)
	minter := CreateMinter(sett.IdFile())
	return sett, minter
}

func GetNode(settings ldp.Settings, path string) (ldp.Node, error) {
	return ldp.GetNode(settings, path)
}

func GetHead(settings ldp.Settings, path string) (ldp.Node, error) {
	return ldp.GetHead(settings, path)
}

func CreateRdfSource(settings ldp.Settings, triples string, parentPath string, minter chan string) (ldp.Node, error) {
	container, err := getContainer(settings, parentPath)
	if err != nil {
		return ldp.Node{}, err
	}

	newPath := MintNextUri("blog", minter)
	node, err := ldp.NewRdfNode(settings, triples, parentPath, newPath)
	if err != nil {
		return ldp.Node{}, err
	}

	if err := container.AddChild(node); err != nil {
		return ldp.Node{}, err
	}
	return node, nil
}

func CreateNonRdfSource(settings ldp.Settings, reader io.ReadCloser, parentPath string, minter chan string) (ldp.Node, error) {
	container, err := getContainer(settings, parentPath)
	if err != nil {
		return ldp.Node{}, err
	}

	newPath := MintNextUri("blog", minter)
	node, err := ldp.NewNonRdfNode(settings, reader, parentPath, newPath)
	if err != nil {
		return node, err
	}

	if err := container.AddChild(node); err != nil {
		return node, err
	}
	return node, nil
}

func PatchNode(settings ldp.Settings, path string, triples string) error {
	node, err := ldp.GetNode(settings, path)
	if err != nil {
		return err
	}
	return node.Patch(triples)
}

func getContainer(settings ldp.Settings, path string) (ldp.Node, error) {
	if path == "" || path == "/" {
		// Shortcut since we know for sure this is a container
		return ldp.GetHead(settings, "/")
	}

	node, err := ldp.GetNode(settings, path)
	if err != nil {
		return node, err
	} else if !node.IsBasicContainer() {
		errorMsg := fmt.Sprintf("%s is not a container", path)
		return node, errors.New(errorMsg)
	}
	return node, nil
}

func getContainerUri(settings ldp.Settings, parentPath string) (string, error) {
	if parentPath == "" || parentPath == "/" {
		return settings.RootUri(), nil
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
