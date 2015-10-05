package ldp

import (
	"errors"
	"fmt"
	"io"
	"ldpserver/rdf"
	"ldpserver/textstore"
	"ldpserver/util"
	"log"
	"strings"
	"time"
)

const NodeNotFound = "Not Found"
const DuplicateNode = "Node already exists"
const metaFile = "meta.rdf"
const dataFile = "data.txt"

type Node struct {
	isRdf   bool
	uri     string
	headers map[string][]string
	graph   rdf.RdfGraph
	binary  string // should be []byte or reader

	settings Settings
	rootUri  string // http://localhost/
	store    textstore.Store

	isBasicContainer   bool
	isDirectContainer  bool
	membershipResource string
	hasMemberRelation  string
	// TODO isMemberOfRelation string
}

func (node Node) Content() string {
	if node.isRdf {
		return node.graph.String()
	}
	return node.binary
}

func (node Node) String() string {
	return node.uri
}

func (node Node) Path() string {
	return util.PathFromUri(node.rootUri, node.uri)
}

func (node Node) Headers() map[string][]string {
	return node.headers
}

func (node Node) IsRdf() bool {
	return node.isRdf
}

func (node Node) IsBasicContainer() bool {
	return node.isBasicContainer
}

func (node Node) IsDirectContainer() bool {
	return node.isDirectContainer
}

func (node Node) HasTriple(predicate, object string) bool {
	return node.graph.HasTriple("<"+node.uri+">", predicate, object)
}

func (node Node) Uri() string {
	return node.uri
}

func (node Node) DebugString() string {
	if !node.isRdf {
		return fmt.Sprintf("Non-RDF: %s", node.uri)
	}

	triples := ""
	for i, triple := range node.graph {
		triples += fmt.Sprintf("%d %s\n", i, triple)
	}
	debugString := fmt.Sprintf("RDF: %s\n %s", node.uri, triples)
	return debugString
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
	if !node.isRdf {
		return errors.New("Cannot PATCH non-RDF Source")
	}

	graph, err := rdf.StringToGraph(triples, "<"+node.uri+">")
	if err != nil {
		return err
	}

	// This is pretty useless as-is since it does not allow to update
	// a triple. It always adds triples.
	// Also, there are some triples that can exist only once (e.g. direct container triples)
	// and this code does not validate them.
	node.graph.Append(graph)

	// write it to disk
	if err := node.writeToDisk(nil); err != nil {
		return err
	}

	return nil
}

func NewRdfNode(settings Settings, triples string, parentPath string, newPath string) (Node, error) {
	path := util.UriConcat(parentPath, newPath)
	node := newNode(settings, path)

	userGraph, err := rdf.StringToGraph(triples, "<"+node.uri+">")
	if err != nil {
		return node, err
	}

	graph := defaultGraph(node.uri)
	graph.Append(userGraph)
	node.setAsRdf(graph)
	err = node.writeToDisk(nil)
	return node, err
}

func NewNonRdfNode(settings Settings, reader io.ReadCloser, parentPath string, newPath string) (Node, error) {
	path := util.UriConcat(parentPath, newPath)
	node := newNode(settings, path)
	graph := defaultNonRdfGraph(node.uri)
	node.setAsNonRdf(graph)
	err := node.writeToDisk(reader)
	return node, err
}

func (node Node) AddChild(child Node) error {
	triple := rdf.NewTriple("<"+node.uri+">", "<"+rdf.LdpContainsUri+">", "<"+child.uri+">")
	err := node.store.AppendToFile(metaFile, triple.StringLn())
	if err != nil {
		return err
	}

	if node.isDirectContainer {
		return node.addDirectContainerChild(child)
	}
	return nil
}

func removeAngleBrackets(text string) string {
	if strings.HasPrefix(text, "<") {
		return text[1 : len(text)-1]
	}
	return text
}
func (node Node) addDirectContainerChild(child Node) error {
	// TODO: account for isMemberOfRelation
	targetUri := removeAngleBrackets(node.membershipResource)
	targetPath := util.PathFromUri(node.rootUri, targetUri)

	targetNode, err := GetNode(node.settings, targetPath)
	if err != nil {
		log.Printf("Could not find target node %s.", targetPath)
		return err
	}

	tripleForTarget := rdf.NewTriple("<"+targetNode.uri+">", node.hasMemberRelation, "<"+child.uri+">")

	err = targetNode.store.AppendToFile(metaFile, tripleForTarget.StringLn())
	if err != nil {
		log.Printf("Error appending child %s to %s. %s", child.uri, targetNode.uri, err)
		return err
	}
	return nil
}

func newNode(settings Settings, path string) Node {
	if strings.HasPrefix(path, "http://") {
		panic("newNode expects a path, received a URI: " + path)
	}
	var node Node
	node.settings = settings
	pathOnDisk := util.PathConcat(settings.dataPath, path)
	node.store = textstore.NewStore(pathOnDisk)
	node.rootUri = settings.RootUri()
	node.uri = util.UriConcat(node.rootUri, path)
	return node
}

func (node *Node) loadNode(isIncludeBody bool) error {
	err := node.loadMeta()
	if err != nil {
		return err
	}

	if node.isRdf || isIncludeBody == false {
		return nil
	}

	return node.loadBinary()
}

func (node *Node) loadBinary() error {
	var err error
	node.binary, err = node.store.ReadFile(dataFile)
	return err
}

func (node *Node) loadMeta() error {
	if !node.store.Exists() {
		return errors.New(NodeNotFound)
	}

	meta, err := node.store.ReadFile(metaFile)
	if err != nil {
		return err
	}

	graph, err := rdf.StringToGraph(meta, node.uri)
	if err != nil {
		return err
	}

	if graph.IsRdfSource("<" + node.uri + ">") {
		node.setAsRdf(graph)
	} else {
		node.setAsNonRdf(graph)
	}
	return nil
}

func (node Node) writeToDisk(reader io.ReadCloser) error {
	// Write the RDF metadata
	err := node.store.SaveFile(metaFile, node.graph.String())
	if err != nil {
		return err
	}

	if node.isRdf {
		return nil
	}

	// Write the binary
	return node.store.SaveReader(dataFile, reader)
}

func defaultGraph(uri string) rdf.RdfGraph {
	subject := "<" + uri + ">"
	// define the triples
	resource := rdf.NewTripleUri(subject, "<"+rdf.RdfTypeUri+">", "<"+rdf.LdpResourceUri+">")
	rdfSource := rdf.NewTripleUri(subject, "<"+rdf.RdfTypeUri+">", "<"+rdf.LdpRdfSourceUri+">")
	// TODO: Not all RDFs resources should be containers
	basicContainer := rdf.NewTripleUri(subject, "<"+rdf.RdfTypeUri+">", "<"+rdf.LdpBasicContainerUri+">")
	title := rdf.NewTripleLit(subject, "<"+rdf.DcTitleUri+">", "\"This is a new entry\"")
	nowString := "\"" + time.Now().Format(time.RFC3339) + "\""
	created := rdf.NewTripleLit(subject, "<"+rdf.DcCreatedUri+">", nowString)
	// create the graph
	graph := rdf.RdfGraph{resource, rdfSource, basicContainer, title, created}
	return graph
}

func defaultNonRdfGraph(uri string) rdf.RdfGraph {
	subject := "<" + uri + ">"
	// define the triples
	resource := rdf.NewTripleUri(subject, "<"+rdf.RdfTypeUri+">", "<"+rdf.LdpResourceUri+">")
	nonRdfSource := rdf.NewTripleUri(subject, "<"+rdf.RdfTypeUri+">", "<"+rdf.LdpNonRdfSourceUri+">")
	title := rdf.NewTripleLit(subject, "<"+rdf.DcTitleUri+">", "\"This is a new entry\"")
	nowString := "\"" + time.Now().Format(time.RFC3339) + "\""
	created := rdf.NewTripleLit(subject, "<"+rdf.DcCreatedUri+">", nowString)
	// create the graph
	graph := rdf.RdfGraph{resource, nonRdfSource, title, created}
	return graph
}

func (node *Node) setAsRdf(graph rdf.RdfGraph) {
	node.isRdf = true
	node.graph = graph
	node.headers = make(map[string][]string)
	node.headers["Content-Type"] = []string{rdf.TurtleContentType}

	if graph.IsBasicContainer("<" + node.uri + ">") {
		// Is there a way to indicate that PUT is allowed
		// for creation only (and not to overwrite?)
		node.headers["Allow"] = []string{"GET, HEAD, POST, PUT"}
	} else {
		node.headers["Allow"] = []string{"GET, HEAD"}
	}

	links := make([]string, 0)
	links = append(links, rdf.LdpResourceLink)
	if graph.IsBasicContainer("<" + node.uri + ">") {
		node.isBasicContainer = true
		links = append(links, rdf.LdpContainerLink)
		links = append(links, rdf.LdpBasicContainerLink)
		// TODO: validate membershipResource is a sub-URI of rootURI
		node.membershipResource, node.hasMemberRelation, node.isDirectContainer = graph.GetDirectContainerInfo()
		if node.isDirectContainer {
			links = append(links, rdf.LdpDirectContainerLink)
		}
	}
	node.headers["Link"] = links
}

func (node *Node) setAsNonRdf(graph rdf.RdfGraph) {
	// TODO Figure out a way to pass the binary as a stream
	node.isRdf = false
	node.graph = graph
	node.binary = ""
	node.headers = make(map[string][]string)
	node.headers["Link"] = []string{rdf.LdpNonRdfSourceLink}
	node.headers["Allow"] = []string{"GET, HEAD"}
	node.headers["Content-Type"] = []string{"application/binary"}
	// TODO: guess the content-type from meta
}
