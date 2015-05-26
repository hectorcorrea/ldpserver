package ldp

import "time"
import "ldpserver/util"
import "ldpserver/fileio"
import "ldpserver/rdf"
import "io"
import "os"
import "errors"
import "log"
import "strings"

const NodeNotFound = "Not Found"

type Node struct {
	IsRdf   bool
	Uri     string
	Headers map[string][]string
	Graph   rdf.RdfGraph // TODO: should this be an embedded type? (even better, maybe it should be private)
	Binary  string       // should be []byte or reader

	dataPath   string // /xyz/data/
	nodeOnDisk string // /xyz/data/blog1/
	metaOnDisk string // /xyz/data/blog1/meta.rdf
	dataOnDisk string // /xyz/data/blog1/data.txt
	rootUri    string // http://localhost/
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

func (node Node) Path() string {
	return node.Uri[len(node.rootUri):]
}

func (node Node) IsBasicContainer() bool {
	return node.Graph.IsBasicContainer(node.Uri)
}

func (node Node) IsDirectContainer() bool {
	return node.Graph.IsDirectContainer()
}

func (node Node) Is(predicate, object string) bool {
	return node.Graph.Is(node.Uri, predicate, object)
}

func GetNode(settings Settings, path string) (Node, error) {
	node := newNode(settings, path)
	err := node.loadNode(true)
	return node, err
}

func GetHead(settings Settings, path string) (Node, error) {
	node := newNode(settings, path)
	err := node.loadNode(false)
	return node, err
}

func (node *Node) Patch(triples string) error {
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
	if err := node.writeToDisk(nil); err != nil {
		return err
	}

	return nil
}

func NewRdfNode(settings Settings, triples string, parentPath string, newPath string) (Node, error) {
	path := util.UriConcat(parentPath, newPath)
	node := newNode(settings, path)

	userGraph, err := rdf.StringToGraph(triples, node.Uri)
	if err != nil {
		return node, err
	}

	graph := defaultGraph(node.Uri)
	graph.Append(userGraph)
	node.makeRdf(graph)
	err = node.writeToDisk(nil)
	return node, err
}

func NewNonRdfNode(settings Settings, reader io.ReadCloser, parentPath string, newPath string) (Node, error) {
	path := util.UriConcat(parentPath, newPath)
	node := newNode(settings, path)
	graph := defaultNonRdfGraph(node.Uri)
	// TODO: pass the reader to make so that save can use it
	node.makeNonRdf(graph)
	err := node.writeToDisk(reader)
	return node, err
}

func (node Node) AddChild(child Node) error {
	triple := rdf.NewTriple(node.Uri, rdf.LdpContainsUri, child.Uri)
	err := fileio.AppendToFile(node.metaOnDisk, triple.StringLn())
	if err != nil {
		log.Printf("%s", err)
		return err
	}
	return nil
}

func newNode(settings Settings, path string) Node {
	if strings.HasPrefix(path, "http://") {
		panic("newNode expects a path, received a URI: " + path)
	}
	var node Node
	node.dataPath = settings.dataPath
	node.nodeOnDisk = util.PathConcat(node.dataPath, path)
	node.metaOnDisk = util.PathConcat(node.nodeOnDisk, "meta.rdf")
	node.dataOnDisk = util.PathConcat(node.nodeOnDisk, "data.txt")
	node.rootUri = settings.RootUri()
	node.Uri = util.UriConcat(node.rootUri, path)
	return node
}

func (node *Node) loadNode(isIncludeBody bool) error {
	err := node.loadMeta()
	if err != nil {
		return err
	}

	if node.IsRdf || isIncludeBody == false {
		return nil
	}

	err2 := node.loadBinary()
	return err2
}

func (node *Node) loadBinary() error {
	var err error
	node.Binary, err = fileio.ReadFile(node.dataOnDisk)
	return err
}

func (node *Node) loadMeta() error {
	log.Printf("Reading %s", node.metaOnDisk)
	if !fileio.FileExists(node.metaOnDisk) {
		return errors.New(NodeNotFound)
	}

	meta, err := fileio.ReadFile(node.metaOnDisk)
	if err != nil {
		return err
	}

	graph, err := rdf.StringToGraph(meta, node.Uri)
	if err != nil {
		return err
	}

	if graph.IsRdfSource(node.Uri) {
		node.makeRdf(graph)
	} else {
		node.makeNonRdf(graph)
	}
	return nil
}

func (node Node) writeToDisk(reader io.ReadCloser) error {
	// Write the RDF metadata
	err := fileio.WriteFile(node.metaOnDisk, node.Graph.String())
	if err != nil {
		return err
	}

	if node.IsRdf {
		return nil
	}

	// Write the binary
	out, err := os.Create(node.dataOnDisk)
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

func (node *Node) makeRdf(graph rdf.RdfGraph) {
	node.IsRdf = true
	node.Graph = graph
	node.Headers = make(map[string][]string)
	node.Headers["Content-Type"] = []string{"text/plain"}

	if graph.IsBasicContainer(node.Uri) {
		node.Headers["Allow"] = []string{"GET, HEAD, POST"}
	} else {
		node.Headers["Allow"] = []string{"GET, HEAD"}
	}

	links := make([]string, 0)
	links = append(links, rdf.LdpResourceLink)
	if graph.IsBasicContainer(node.Uri) {
		links = append(links, rdf.LdpContainerLink)
		links = append(links, rdf.LdpBasicContainerLink)
		if graph.IsDirectContainer() {
			links = append(links, rdf.LdpDirectContainerLink)
		}
	}
	node.Headers["Link"] = links
}

func (node *Node) makeNonRdf(graph rdf.RdfGraph) {
	// TODO Figure out a way to pass the binary as a stream
	node.IsRdf = false
	node.Graph = graph
	node.Binary = ""
	node.Headers = make(map[string][]string)
	node.Headers["Link"] = []string{rdf.LdpNonRdfSourceLink}
	node.Headers["Allow"] = []string{"GET, HEAD"}
	// TODO: guess the content-type from meta
}
