package ldp

import "time"
import "ldpserver/fileio"
import "ldpserver/rdf"
import "io"
import "os"
import "errors"
import "log"

const NodeNotFound = "Not Found"

type Node struct {
	IsRdf   bool
	Uri     string
	Headers map[string]string
	Graph   rdf.RdfGraph // TODO: should this be an embedded type?
	Binary  string       // should be []byte or reader
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

func (node Node) IsBasicContainer() bool {
	return node.Graph.IsBasicContainer(node.Uri)
}

func GetNode(settings Settings, path string) (Node, error) {
	return getNode(settings, path, true)
}

func GetHead(settings Settings, path string) (Node, error) {
	return getNode(settings, path, false)
}

func (node *Node) Patch(settings Settings, triples string) error {
	if !node.IsRdf {
		return errors.New("Cannot PATCH non-RDF Source")
	}

	graph, err := rdf.StringToGraph(triples, node.Uri)
	if err != nil {
		return err
	}

	// This is pretty useless as-is since it does not allow to update
	// a triple. It always adds triples.
	node.Graph.Append(graph)

	// write it to disk
	if err := node.writeToDisk(settings, nil); err != nil {
		return err
	}

	return nil
}

// TODO: move disk operations to a repo type
func getNode(settings Settings, path string, isIncludeBody bool) (Node, error) {
	node, err := getMeta(settings, path)
	if err != nil {
		return Node{}, err
	}

	if node.IsRdf || isIncludeBody == false {
		return node, nil
	}

	_, dataOnDisk := FileNamesForPath(settings, path)
	node.Binary, err = fileio.ReadFile(dataOnDisk)
	if err != nil {
		return Node{}, err
	}
	return node, nil
}

// TODO: move disk operations to a repo type
func getMeta(settings Settings, path string) (Node, error) {
	metaOnDisk, _ := FileNamesForPath(settings, path)
	log.Printf("Reading %s", metaOnDisk)
	if !fileio.FileExists(metaOnDisk) {
		return Node{}, errors.New(NodeNotFound)
	}

	uri := UriConcat(settings.rootUrl, path)
	meta, err := fileio.ReadFile(metaOnDisk)
	if err != nil {
		return Node{}, err
	}

	graph, err := rdf.StringToGraph(meta, uri)
	if err != nil {
		return Node{}, err
	}

	var node Node
	if graph.IsRdfSource(uri) {
		node = newRdfNodeFromGraph(uri, graph)
	} else {
		node = newNonRdfNode(uri, graph, "")
	}
	return node, nil
}

func NewRdfNode(settings Settings, triples string, parentUri string) (Node, error) {
	newUri := MintNextUri(settings, "blog")
	fullUri := UriConcat(parentUri, newUri)
	userGraph, err := rdf.StringToGraph(triples, fullUri)
	if err != nil {
		return Node{}, err
	}
	graph := defaultGraph(fullUri)
	graph.Append(userGraph)
	node := newRdfNodeFromGraph(fullUri, graph)
	if err := node.writeToDisk(settings, nil); err != nil {
		return node, err
	}
	return node, nil
}

func NewNonRdfNode(settings Settings, reader io.ReadCloser, parentUri string) (Node, error) {
	newUri := MintNextUri(settings, "blog")
	fullUri := UriConcat(parentUri, newUri)
	graph := defaultNonRdfGraph(fullUri)
	node := newNonRdfNode(fullUri, graph, "")
	if err := node.writeToDisk(settings, reader); err != nil {
		return node, err
	}
	return node, nil
}

func AddChildToContainer(settings Settings, childUri string, parentUri string) {
	triple := rdf.NewTriple(parentUri, rdf.LdpContainsUri, childUri)
	metaOnDisk, _ := FileNamesForUri(settings, parentUri)
	err := fileio.AppendToFile(metaOnDisk, triple.StringLn())
	if err != nil {
		log.Printf("%s", err)
		panic("Could not append child to container " + parentUri)
	}
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

func defaultGraph(subject string) rdf.RdfGraph {
	// define the triples
	resource := rdf.NewTriple(subject, rdf.RdfTypeUri, rdf.LdpResourceUri)
	rdfSource := rdf.NewTriple(subject, rdf.RdfTypeUri, rdf.LdpRdfSourceUri)
	// TODO: Not all RDFs resources should be containers
	basicContainer := rdf.NewTriple(subject, rdf.RdfTypeUri, rdf.LdpBasicContainerUri)
	title := rdf.NewTriple(subject, rdf.DcTitleUri, "This is a new entry")
	nowString := time.Now().Format(time.RFC3339)
	created := rdf.NewTriple(subject, rdf.DcCreatedUri, nowString)
	// create the graph
	graph := rdf.RdfGraph{resource, rdfSource, basicContainer, title, created}
	return graph
}

func defaultNonRdfGraph(subject string) rdf.RdfGraph {
	// define the triples
	resource := rdf.NewTriple(subject, rdf.RdfTypeUri, rdf.LdpResourceUri)
	nonRdfSource := rdf.NewTriple(subject, rdf.RdfTypeUri, rdf.LdpNonRdfSourceUri)
	title := rdf.NewTriple(subject, rdf.DcTitleUri, "This is a new entry")
	nowString := time.Now().Format(time.RFC3339)
	created := rdf.NewTriple(subject, rdf.DcCreatedUri, nowString)
	// create the graph
	graph := rdf.RdfGraph{resource, nonRdfSource, title, created}
	return graph
}

func newRdfNodeFromGraph(uri string, graph rdf.RdfGraph) Node {
	var node Node
	node.IsRdf = true
	node.Uri = uri
	node.Graph = graph
	node.Headers = make(map[string]string)
	node.Headers["Link"] = rdf.LdpResourceLink
	node.Headers["Content-Type"] = "text/plain"
	if graph.IsBasicContainer(uri) {
		node.Headers["Link"] = rdf.LdpContainerLink
		node.Headers["Link"] = rdf.LdpBasicContainerLink
		node.Headers["Allow"] = "GET, HEAD, POST"
	} else {
		node.Headers["Allow"] = "GET, HEAD"
	}

	return node
}

func newNonRdfNode(uri string, graph rdf.RdfGraph, binary string) Node {
	// TODO Figure out a way to pass the binary as a stream
	var node Node
	node.IsRdf = false
	node.Uri = uri
	node.Graph = graph
	node.Binary = binary
	node.Headers = make(map[string]string)
	node.Headers["Link"] = rdf.LdpNonRdfSourceLink
	node.Headers["Allow"] = "GET, HEAD"
	// TODO: guess the content-type from meta
	return node
}
