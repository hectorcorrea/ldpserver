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

var NodeNotFoundError = errors.New("Node not found")
var DuplicateNodeError = errors.New("Node already exists")
var EtagMissingError = errors.New("Missing Etag")
var EtagMismatchError = errors.New("Etag mismatch")
var ServerManagedPropertyError = errors.New("Attempted to update server managed property")

const metaFile = "meta.rdf"
const dataFile = "data.txt"
const etagPredicate = "<" + rdf.ServerETagUri + ">"
const rdfTypePredicate = "<" + rdf.RdfTypeUri + ">"
const contentTypePredicate = "<" + rdf.ServerContentTypeUri + ">"

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

func (node Node) Content() string {
	if node.isRdf {
		return node.graph.String()
	}
	return node.binary
}

func (node Node) nonRdfContentType() string {
	if node.isRdf {
		panic("Cannot call NonRdfContentType() for an RDF node")
	}
	triple, found := node.graph.FindTriple("<"+node.uri+">", contentTypePredicate)
	if !found {
		return "application/binary"
	}
	return triple.Object()
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

func (node *Node) Etag() string {
	subject := "<" + node.uri + ">"
	etag, etagFound := node.graph.GetObject(subject, "<"+rdf.ServerETagUri+">")
	if !etagFound {
		panic(fmt.Sprintf("No etag found for node %s", node.uri))
	}
	return etag
}

func (node Node) HasTriple(predicate, object string) bool {
	return node.graph.HasTriple("<"+node.uri+">", predicate, object)
}

func (node Node) Headers() map[string][]string {
	return node.headers
}

func (node Node) IsBasicContainer() bool {
	return node.isBasicContainer
}

func (node Node) IsDirectContainer() bool {
	return node.isDirectContainer
}

func (node Node) IsRdf() bool {
	return node.isRdf
}

func (node Node) Uri() string {
	return node.uri
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

func (node Node) Path() string {
	return util.PathFromUri(node.rootUri, node.uri)
}

func (node Node) String() string {
	return node.uri
}

func (node *Node) appendTriple(predicate, object string) {
	node.graph.AppendTriple2("<"+node.uri+">", predicate, object)
}

func (node *Node) setETag() {
	node.graph.SetObject("<"+node.uri+">", etagPredicate, calculateEtag())
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

func NewRdfNode(settings Settings, triples string, path string) (Node, error) {
	node := newNode(settings, path)
	graph, err := rdf.StringToGraph(triples, "<"+node.uri+">")
	if err != nil {
		return Node{}, err
	}
	return node, node.writeRdfToDisk(graph)
}

func NewNonRdfNode(settings Settings, reader io.ReadCloser, path, triples string) (Node, error) {
	node := newNode(settings, path)
	graph, err := rdf.StringToGraph(triples, "<"+node.uri+">")
	if err != nil {
		return Node{}, err
	}
	return node, node.writeNonRdfToDisk(graph, reader)
}

func ReplaceNonRdfNode(settings Settings, reader io.ReadCloser, path, etag, triples string) (Node, error) {
	node, err := GetHead(settings, path)
	if err != nil {
		return Node{}, err
	}

	if node.isRdf {
		return Node{}, errors.New("Cannot replace RDF source with a Non-RDF source")
	}

	if etag == "" {
		return Node{}, EtagMissingError
	}

	if node.Etag() != etag {
		// log.Printf("Cannot replace RDF source. Etag mismatch. Expected: %s. Found: %s", node.Etag(), etag)
		return Node{}, EtagMismatchError
	}

	var graph rdf.RdfGraph
	if triples != "" {
		graph, err = rdf.StringToGraph(triples, "<"+node.uri+">")
		if err != nil {
			return Node{}, err
		}
	}

	return node, node.writeNonRdfToDisk(graph, reader)
}

func ReplaceRdfNode(settings Settings, triples string, path string, etag string) (Node, error) {
	node, err := GetNode(settings, path)
	if err != nil {
		return Node{}, err
	}

	if !node.isRdf {
		return Node{}, errors.New("Cannot replace non-RDF source with an RDF source")
	}

	if etag == "" {
		return Node{}, EtagMissingError
	}

	if node.Etag() != etag {
		// log.Printf("Cannot replace RDF source. Etag mismatch. Expected: %s. Found: %s", node.Etag(), etag)
		return Node{}, EtagMismatchError
	}

	graph, err := rdf.StringToGraph(triples, "<"+node.uri+">")
	if err != nil {
		return Node{}, err
	}

	// TODO: What other server-managed properties should we handle?
	if graph.HasPredicate("<"+node.uri+">", "<"+rdf.LdpContainsUri+">") {
		return Node{}, ServerManagedPropertyError
	}

	return node, node.writeRdfToDisk(graph)
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
		return NodeNotFoundError
	}

	meta, err := node.store.ReadFile(metaFile)
	if err != nil {
		return err
	}

	node.graph, err = rdf.StringToGraph(meta, node.uri)
	if err != nil {
		return err
	}

	if node.graph.IsRdfSource("<" + node.uri + ">") {
		node.setAsRdf()
	} else {
		node.setAsNonRdf()
	}
	return nil
}

func (node *Node) writeNonRdfToDisk(graph rdf.RdfGraph, reader io.ReadCloser) error {
	node.graph = graph
	node.setETag()
	node.appendTriple(rdfTypePredicate, "<"+rdf.LdpResourceUri+">")
	node.appendTriple(rdfTypePredicate, "<"+rdf.LdpNonRdfSourceUri+">")
	node.setAsNonRdf()
	return node.writeToDisk(reader)
}

func (node *Node) writeRdfToDisk(graph rdf.RdfGraph) error {
	node.graph = graph
	node.setETag()
	node.appendTriple(rdfTypePredicate, "<"+rdf.LdpResourceUri+">")
	node.appendTriple(rdfTypePredicate, "<"+rdf.LdpRdfSourceUri+">")
	node.appendTriple(rdfTypePredicate, "<"+rdf.LdpContainerUri+">")
	node.appendTriple(rdfTypePredicate, "<"+rdf.LdpBasicContainerUri+">")
	node.setAsRdf()
	return node.writeToDisk(nil)
}

func (node *Node) writeToDisk(reader io.ReadCloser) error {
	// Write the RDF metadata
	err := node.store.SaveFile(metaFile, node.graph.String())
	if err != nil {
		return err
	}

	if node.isRdf {
		return nil
	}

	// Write the binary...
	err = node.store.SaveReader(dataFile, reader)
	if err != nil {
		return err
	}

	// ...update the copy in memory (this would get
	// tricky when we switch "node.binary" to a
	// reader.
	node.binary, err = node.store.ReadFile(dataFile)
	return err
}

func (node *Node) setAsRdf() {
	subject := "<" + node.uri + ">"
	node.isRdf = true
	node.headers = make(map[string][]string)
	node.headers["Content-Type"] = []string{rdf.TurtleContentType}

	if node.graph.IsBasicContainer(subject) {
		// Is there a way to indicate that PUT is allowed
		// for creation only (and not to overwrite?)
		node.headers["Allow"] = []string{"GET, HEAD, POST, PUT, PATCH"}
	} else {
		node.headers["Allow"] = []string{"GET, HEAD, PUT, PATCH"}
	}
	node.headers["Accept-Post"] = []string{"text/turtle"}
	node.headers["Accept-Patch"] = []string{"text/turtle"}

	node.headers["Etag"] = []string{node.Etag()}

	links := make([]string, 0)
	links = append(links, rdf.LdpResourceLink)
	if node.graph.IsBasicContainer(subject) {
		node.isBasicContainer = true
		links = append(links, rdf.LdpContainerLink)
		links = append(links, rdf.LdpBasicContainerLink)
		// TODO: validate membershipResource is a sub-URI of rootURI
		node.membershipResource, node.hasMemberRelation, node.isDirectContainer = node.graph.GetDirectContainerInfo()
		if node.isDirectContainer {
			links = append(links, rdf.LdpDirectContainerLink)
		}
	}
	node.headers["Link"] = links
}

func (node *Node) setAsNonRdf() {
	// TODO Figure out a way to pass the binary as a stream
	node.isRdf = false
	node.binary = ""
	node.headers = make(map[string][]string)
	node.headers["Link"] = []string{rdf.LdpResourceLink, rdf.LdpNonRdfSourceLink}
	node.headers["Allow"] = []string{"GET, HEAD, PUT"}
	node.headers["Content-Type"] = []string{node.nonRdfContentType()}
	node.headers["Etag"] = []string{node.Etag()}
}

func calculateEtag() string {
	// TODO: Come up with a more precise value.
	now := time.Now().Format(time.RFC3339)
	etag := strings.Replace(now, ":", "_", -1)
	return "\"" + etag + "\""
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

func removeAngleBrackets(text string) string {
	if strings.HasPrefix(text, "<") && strings.HasSuffix(text, ">") {
		return text[1 : len(text)-1]
	}
	return text
}
