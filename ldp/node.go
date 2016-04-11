package ldp

import (
	"errors"
	"fmt"
	"github.com/hectorcorrea/rdf"
	"io"
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

type PreferTriples struct {
	Containment      bool
	Membership       bool
	MinimalContainer bool
}

type Node struct {
	isRdf      bool
	uri        string // http://localhost/node1
	subject    string // <http://localhost/node1>
	headers    map[string][]string
	graph      rdf.RdfGraph
	graphExtra rdf.RdfGraph // triples from included resources (see PreferTriples)
	binary     string       // should be []byte or reader

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
	triple := rdf.NewTriple(node.subject, "<"+rdf.LdpContainsUri+">", child.subject)
	err := node.store.AppendToMetaFile(triple.StringLn())
	if err != nil {
		return err
	}

	if node.isDirectContainer {
		return node.addDirectContainerChild(child)
	}
	return nil
}

func (node Node) ContentPref(pref PreferTriples) string {
	if node.isRdf {
		var triples rdf.RdfGraph
		if pref.MinimalContainer {
			// All but ldpContains
			for _, triple := range node.graph {
				if !triple.Is("<" + rdf.LdpContainsUri + ">") {
					triples = append(triples, triple)
				}
			}
		} else {
			triples = node.graph
		}
		triplesStr := triples.String()
		if node.graphExtra != nil {
			triplesStr += "\n" + node.graphExtra.String()
		}
		return triplesStr
	}
	return node.binary
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

func (node Node) contentType() string {
	if node.isRdf {
		return rdf.TurtleContentType
	}
	triple, found := node.graph.FindPredicate(node.subject, contentTypePredicate)
	if !found {
		return "application/binary"
	}
	return util.RemoveQuotes(triple.Object())
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
	etag, etagFound := node.graph.GetObject(node.subject, "<"+rdf.ServerETagUri+">")
	if !etagFound {
		panic(fmt.Sprintf("No etag found for node %s", node.uri))
	}
	return etag
}

func (node Node) HasTriple(predicate, object string) bool {
	return node.graph.HasTriple(node.subject, predicate, object)
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

	userGraph, err := rdf.StringToGraph(triples, node.subject)
	if err != nil {
		return err
	}

	if hasServerManagedProperties(userGraph, node.subject) {
		return ServerManagedPropertyError
	}

	// This is pretty useless as-is since it does not allow to update
	// a triple. It always adds triples.
	node.graph.Append(userGraph)
	return node.save(node.graph, nil)
}

func (node Node) Path() string {
	return util.PathFromUri(node.rootUri, node.uri)
}

func (node Node) String() string {
	return node.uri
}

func (node *Node) appendTriple(predicate, object string) {
	node.graph.AppendTripleStr(node.subject, predicate, object)
}

func (node *Node) setETag() {
	node.graph.SetObject(node.subject, etagPredicate, calculateEtag())
}

func (node *Node) Delete() error {
	return node.store.Delete()
}

func (node *Node) RemoveContainsUri(uri string) error {
	predicate := "<" + rdf.LdpContainsUri + ">"
	object := uri
	deleted := node.graph.DeleteTriple(node.subject, predicate, object)
	if !deleted {
		return errors.New("Failed to deleted the containment triple")
	}
	return node.save(node.graph, nil)
}

func getNode(settings Settings, path string) (Node, error) {
	return GetNode(settings, path, PreferTriples{})
}

func GetNode(settings Settings, path string, pref PreferTriples) (Node, error) {
	node := newNode(settings, path)
	err := node.loadNode(true)

	if pref.Membership && node.IsDirectContainer() {
		// Fetch the triples from the membershipResource
		log.Printf("Fetching membershipResource's graph: %s", node.membershipResourcePath())
		memberNode, err := getNode(settings, node.membershipResourcePath())
		if err != nil {
			return node, err
		}
		node.graphExtra = memberNode.graph
		// TODO: set node.headers["Preference-Applied"] = "return=representation"
	}

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
	graph, err := rdf.StringToGraph(triples, node.subject)
	if err != nil {
		return Node{}, err
	}
	return node, node.save(graph, nil)
}

func NewNonRdfNode(settings Settings, reader io.ReadCloser, path, triples string) (Node, error) {
	node := newNode(settings, path)
	node.isRdf = false
	graph, err := rdf.StringToGraph(triples, node.subject)
	if err != nil {
		return Node{}, err
	}
	return node, node.save(graph, reader)
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
		graph, err = rdf.StringToGraph(triples, node.subject)
		if err != nil {
			return Node{}, err
		}
	}
	return node, node.save(graph, reader)
}

func ReplaceRdfNode(settings Settings, triples string, path string, etag string) (Node, error) {
	node, err := getNode(settings, path)
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

	graph, err := rdf.StringToGraph(triples, node.subject)
	if err != nil {
		return Node{}, err
	}

	if hasServerManagedProperties(graph, node.subject) {
		return Node{}, ServerManagedPropertyError
	}

	return node, node.save(graph, nil)
}

func (node Node) addDirectContainerChild(child Node) error {
	// TODO: account for isMemberOfRelation
	targetUri := util.RemoveAngleBrackets(node.membershipResource)
	targetPath := util.PathFromUri(node.rootUri, targetUri)

	targetNode, err := getNode(node.settings, targetPath)
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

	node.graph, err = rdf.StringToGraph(meta, node.subject)
	if err != nil {
		return err
	}

	if node.graph.IsRdfSource(node.subject) {
		node.isRdf = true
		node.setAsRdf()
	} else {
		node.isRdf = false
		node.setAsNonRdf()
	}
	return nil
}

func (node *Node) save(graph rdf.RdfGraph, reader io.ReadCloser) error {
	node.graph = graph

	if node.graph.IsDirectContainer() {
		// TODO: we might need a different triple for DCs using isMemberOfRelation
		node.appendTriple("<"+rdf.LdpInsertedContentRelationUri+">", "<"+rdf.LdpMemberSubjectUri+">")
	}

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
	node.headers = make(map[string][]string)
	node.headers["Content-Type"] = []string{node.contentType()}

	if node.graph.IsBasicContainer(node.subject) {
		node.headers["Allow"] = []string{"GET, HEAD, POST, PUT, PATCH"}
	} else {
		node.headers["Allow"] = []string{"GET, HEAD, PUT, PATCH"}
	}
	node.headers["Accept-Post"] = []string{rdf.TurtleContentType}
	node.headers["Accept-Patch"] = []string{rdf.TurtleContentType}

	node.headers["Etag"] = []string{node.Etag()}

	links := make([]string, 0)
	links = append(links, rdf.LdpResourceLink)
	if node.graph.IsBasicContainer(node.subject) {
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
	node.headers["Content-Type"] = []string{node.contentType()}
	node.headers["Etag"] = []string{node.Etag()}
}

func (node *Node) membershipResourcePath() string {
	uri := util.RemoveAngleBrackets(node.membershipResource)
	return strings.Replace(uri, node.settings.rootUri, "", 1)
}

func hasServerManagedProperties(graph rdf.RdfGraph, subject string) bool {
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
	node.subject = "<" + node.uri + ">"
	return node
}
