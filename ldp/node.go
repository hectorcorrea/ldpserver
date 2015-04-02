package ldp

import "time"
import "ldpserver/fileio"
import "io"
import "os"
import "errors"
import "log"

const NodeNotFound = "Not Found"

type Node struct {
	IsRdf   bool
	Uri     string
	Headers map[string]string
	Graph   RdfGraph
	Binary  string // should be []byte or reader
}

func (node Node) Content() string {
	if node.IsRdf {
		return node.Graph.String()
	}
	return node.Binary
}

func (node Node) String() string {
	return node.Uri
}

func CreateRdfSource(settings Settings, triples string, parentPath string) (Node, error) {
	parentUri, err := getContainerUri(settings, parentPath)
	if err != nil {
		return Node{}, err
	}

	node, err := createRdfNode(settings, triples, parentUri)
	if err != nil {
		return node, err
	}
	addChildToContainer(settings, node.Uri, parentUri)
	return node, nil
}

func CreateNonRdfSource(settings Settings, reader io.ReadCloser, parentPath string) (Node, error) {
	parentUri, err := getContainerUri(settings, parentPath)
	if err != nil {
		return Node{}, err
	}

	node, err := createNonRdfNode(settings, reader, parentUri)
	if err != nil {
		return node, err
	}
	addChildToContainer(settings, node.Uri, parentUri)
	return node, nil
}

func GetNode(settings Settings, path string) (Node, error) {
	metaOnDisk, dataOnDisk := FileNamesForPath(settings, path)
	log.Printf("Reading %s", metaOnDisk)
	if !fileio.FileExists(metaOnDisk) {
		var emptyNode Node
		return emptyNode, errors.New(NodeNotFound)
	}

	uri := UriConcat(settings.rootUrl, path)
	meta, err := fileio.ReadFile(metaOnDisk)
	if err != nil {
		var emptyNode Node
		return emptyNode, err
	}

	graph, err := StringToGraph(meta, uri)
	if err != nil {
		var emptyNode Node
		return emptyNode, err
	}

	// RDF
	if graph.IsRdfSource(uri) {
		node := newRdfNode(uri, graph)
		return node, nil
	}

	// non-RDF
	content, err := fileio.ReadFile(dataOnDisk)
	if err != nil {
		var emptyNode Node
		return emptyNode, err
	}

	node := newNonRdfNode(uri, graph, content)
	return node, nil
}

func PatchNode(settings Settings, path string, triples string) (Node, error) {
	node, err := GetNode(settings, path)
	if err != nil {
		return node, err
	}

	if !node.IsRdf {
		return node, errors.New("Cannot PATCH non-RDF Source")
	}

	if len(triples) == 0 {
		// Should we return a different status (e.g. nothing done)
		return node, nil
	}

	graph, err := StringToGraph(triples, node.Uri)
	if err != nil {
		return node, err
	}

	// This is pretty useless as-is since it does not allow to update
	// a triple. It always adds triples.
	node.Graph.Append(graph)

	// write it to disk
	if err := node.writeToDisk(settings, nil); err != nil {
		return node, err
	}

	return node, nil
}

func GetHead(settings Settings, path string) (Node, error) {
	metaOnDisk, _ := FileNamesForPath(settings, path)
	if !fileio.FileExists(metaOnDisk) {
		var emptyNode Node
		return emptyNode, errors.New(NodeNotFound)
	}

	uri := UriConcat(settings.rootUrl, path)
	meta, err := fileio.ReadFile(metaOnDisk)
	if err != nil {
		var emptyNode Node
		return emptyNode, err
	}

	// RDF
	graph, _ := StringToGraph(meta, uri)
	if graph.IsRdfSource(uri) {
		node := newRdfNode(uri, graph)
		return node, nil
	}

	// non-RDF
	node := newNonRdfNode(uri, graph, "")
	return node, nil
}

func getContainerUri(settings Settings, parentPath string) (string, error) {
	if parentPath == "" || parentPath == "/" {
		return settings.rootUrl, nil
	}

	// Make sure the parent node exists and it's a container
	parentNode, err := GetNode(settings, parentPath)
	if err != nil {
		return "", err
	} else if !parentNode.Graph.IsBasicContainer(parentNode.Uri) {
		return "", errors.New("Parent is not a container")
	}
	return parentNode.Uri, nil
}

func createRdfNode(settings Settings, triples string, parentUri string) (Node, error) {
	newUri := MintNextUri(settings, "blog")
	fullUri := UriConcat(parentUri, newUri)
	userGraph, err := StringToGraph(triples, fullUri)
	if err != nil {
		return Node{}, err
	}
	graph := defaultGraph(fullUri)
	graph.Append(userGraph)
	node := newRdfNode(fullUri, graph)
	if err := node.writeToDisk(settings, nil); err != nil {
		return node, err
	}
	return node, nil
}

func createNonRdfNode(settings Settings, reader io.ReadCloser, parentUri string) (Node, error) {
	newUri := MintNextUri(settings, "blog")
	fullUri := UriConcat(parentUri, newUri)
	graph := defaultNonRdfGraph(fullUri)
	node := newNonRdfNode(fullUri, graph, "")
	if err := node.writeToDisk(settings, reader); err != nil {
		return node, err
	}
	return node, nil
}

func (node Node) writeToDisk(settings Settings, reader io.ReadCloser) error {
	// Write the RDF metadata
	metaOnDisk, dataOnDisk := FileNamesForUri(settings, node.Uri)
	err := fileio.WriteFile(metaOnDisk, node.Graph.String())
	if err != nil {
		return err
	}

	if node.IsRdf {
		return nil
	}

	// Write the binary
	out, err := os.Create(dataOnDisk)
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(out, reader)
	return out.Close()
}

func defaultGraph(subject string) RdfGraph {
	// define the triples
	resource := NewTriple(subject, RdfTypeUri, LdpResourceUri)
	rdfSource := NewTriple(subject, RdfTypeUri, LdpRdfSourceUri)
	// TODO: Not all RDFs resources should be containers
	basicContainer := NewTriple(subject, RdfTypeUri, LdpBasicContainerUri)
	title := NewTriple(subject, DcTitleUri, "This is a new entry")
	nowString := time.Now().Format(time.RFC3339)
	created := NewTriple(subject, DcCreatedUri, nowString)
	// create the graph
	graph := RdfGraph{resource, rdfSource, basicContainer, title, created}
	return graph
}

func defaultNonRdfGraph(subject string) RdfGraph {
	// define the triples
	resource := NewTriple(subject, RdfTypeUri, LdpResourceUri)
	nonRdfSource := NewTriple(subject, RdfTypeUri, LdpNonRdfSourceUri)
	title := NewTriple(subject, DcTitleUri, "This is a new entry")
	nowString := time.Now().Format(time.RFC3339)
	created := NewTriple(subject, DcCreatedUri, nowString)
	// create the graph
	graph := RdfGraph{resource, nonRdfSource, title, created}
	return graph
}

func newRdfNode(uri string, graph RdfGraph) Node {
	var node Node
	node.IsRdf = true
	node.Uri = uri
	node.Graph = graph
	node.Headers = make(map[string]string)
	node.Headers["Link"] = LdpResourceLink
	node.Headers["Content-Type"] = "text/plain"
	if graph.IsBasicContainer(uri) {
		node.Headers["Link"] = LdpContainerLink
		node.Headers["Link"] = LdpBasicContainerLink
		node.Headers["Allow"] = "GET, HEAD, POST"
	} else {
		node.Headers["Allow"] = "GET, HEAD"
	}

	return node
}

func newNonRdfNode(uri string, graph RdfGraph, binary string) Node {
	// TODO Figure out a way to pass the binary as a stream
	var node Node
	node.IsRdf = false
	node.Uri = uri
	node.Graph = graph
	node.Binary = binary
	node.Headers = make(map[string]string)
	node.Headers["Link"] = LdpNonRdfSourceLink
	node.Headers["Allow"] = "GET, HEAD"
	// TODO: guess the content-type from meta
	return node
}

func addChildToContainer(settings Settings, childUri string, parentUri string) {
	triple := NewTriple(parentUri, LdpContainsUri, childUri)
	metaOnDisk, _ := FileNamesForUri(settings, parentUri)
	err := fileio.AppendToFile(metaOnDisk, triple.StringLn())
	if err != nil {
		log.Printf("%s", err)
		panic("Could not append child to container " + parentUri)
	}
}
