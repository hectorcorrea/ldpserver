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
	err := node.store.AppendToMetaFile(triple.StringLn())
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

func (node Node) Metadata() string {
	return node.graph.String()
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

	userGraph, err := rdf.StringToGraph(triples, "<"+node.uri+">")
	if err != nil {
		return err
	}

	if hasServerManagedProperties(userGraph, node.uri) {
		return ServerManagedPropertyError
	}

	// This is pretty useless as-is since it does not allow to update
	// a triple. It always adds triples.
	node.graph.Append(userGraph)
	return node.save(nil)
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

func (node *Node) Delete() error {
	return node.store.Delete()
}

func (node *Node) RemoveContainsUri(uri string) error {
	subject := "<" + node.uri + ">"
	predicate := "<" + rdf.LdpContainsUri + ">"
	object := uri
	deleted := node.graph.DeleteTriple(subject, predicate, object)
	if !deleted {
		return errors.New("Failed to deleted the containment triple")
	}
	return node.save(nil)
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
	node.isRdf = true
	graph, err := rdf.StringToGraph(triples, "<"+node.uri+">")
	if err != nil {
		return Node{}, err
	}
	node.graph = graph
	return node, node.save(nil)
}

func NewNonRdfNode(settings Settings, reader io.ReadCloser, path, triples string) (Node, error) {
	node := newNode(settings, path)
	node.isRdf = false
	graph, err := rdf.StringToGraph(triples, "<"+node.uri+">")
	if err != nil {
		return Node{}, err
	}
	node.graph = graph
	return node, node.save(reader)
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
	node.graph = graph
	return node, node.save(reader)
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

	if hasServerManagedProperties(graph, node.uri) {
		return Node{}, ServerManagedPropertyError
	}

	node.graph = graph
	return node, node.save(nil)
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

	err = targetNode.store.AppendToMetaFile(tripleForTarget.StringLn())
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
	node.binary, err = node.store.ReadDataFile()
	return err
}

func (node *Node) loadMeta() error {
	if !node.store.Exists() {
		return NodeNotFoundError
	}

	meta, err := node.store.ReadMetaFile()
	if err != nil {
		return err
	}

	node.graph, err = rdf.StringToGraph(meta, node.uri)
	if err != nil {
		return err
	}

	if node.graph.IsRdfSource("<" + node.uri + ">") {
		node.isRdf = true
		node.setAsRdf()
	} else {
		node.isRdf = false
		node.setAsNonRdf()
	}
	return nil
}

func (node *Node) save(reader io.ReadCloser) error {
	node.setETag()
	node.appendTriple(rdfTypePredicate, "<"+rdf.LdpResourceUri+">")
	if node.isRdf {
		node.appendTriple(rdfTypePredicate, "<"+rdf.LdpRdfSourceUri+">")
		node.appendTriple(rdfTypePredicate, "<"+rdf.LdpContainerUri+">")
		node.appendTriple(rdfTypePredicate, "<"+rdf.LdpBasicContainerUri+">")
		node.setAsRdf()
	} else {
		node.appendTriple(rdfTypePredicate, "<"+rdf.LdpNonRdfSourceUri+">")
		node.setAsNonRdf()
	}
	return node.writeToDisk(reader)
}

func (node *Node) writeToDisk(reader io.ReadCloser) error {
	// Write the RDF metadata
	err := node.store.SaveMetaFile(node.graph.String())
	if err != nil {
		return err
	}

	if node.isRdf {
		return nil
	}

	// Write the binary...
	err = node.store.SaveDataFile(reader)
	if err != nil {
		return err
	}

	// ...update the copy in memory (this would get
	// tricky when we switch "node.binary" to a
	// reader.
	node.binary, err = node.store.ReadDataFile()
	return err
}

func (node *Node) setAsRdf() {
	subject := "<" + node.uri + ">"
	node.headers = make(map[string][]string)
	node.headers["Content-Type"] = []string{rdf.TurtleContentType}

	if node.graph.IsBasicContainer(subject) {
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
	node.binary = ""
	node.headers = make(map[string][]string)

	describedByLink := fmt.Sprintf("<%s?metadata=yes>; rel=\"describedby\"; anchor=\"%s\"", node.uri, node.uri)
	node.headers["Link"] = []string{describedByLink, rdf.LdpResourceLink, rdf.LdpNonRdfSourceLink}

	node.headers["Allow"] = []string{"GET, HEAD, PUT"}
	node.headers["Content-Type"] = []string{removeQuotes(node.nonRdfContentType())}
	node.headers["Etag"] = []string{node.Etag()}
}

func hasServerManagedProperties(graph rdf.RdfGraph, uri string) bool {
	subject := "<" + uri + ">"

	// TODO: What other server-managed properties should we handle?
	properties := []string{rdf.LdpResourceUri, rdf.LdpRdfSourceUri, rdf.LdpNonRdfSourceUri,
		rdf.LdpContainerUri, rdf.LdpBasicContainerUri, rdf.LdpDirectContainerUri, rdf.LdpContainsUri,
		rdf.LdpConstrainedBy}

	for _, property := range properties {
		if graph.HasPredicate(subject, "<"+property+">") {
			return true
		}
	}
	return false
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

func removeQuotes(text string) string {
	if strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"") {
		return text[1 : len(text)-1]
	}
	return text
}
