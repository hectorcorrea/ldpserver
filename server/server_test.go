package server

import (
	"fmt"
	"github.com/hectorcorrea/rdf"
	"ldpserver/ldp"
	"ldpserver/util"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var dataPath string
var theServer Server
var rootUrl = "http://localhost:9001/"
var emptySlug = ""

func init() {
	dataPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	theServer = NewServer(rootUrl, dataPath)
}

func TestBadSlug(t *testing.T) {
	_, err := theServer.CreateRdfSource("", "/", "/invalid/")
	if err == nil {
		t.Error("Failed to detect an invalid slug")
	}
}

func TestCreateRdf(t *testing.T) {
	_, err := theServer.CreateRdfSource("", "/", "slugA")
	if err != nil {
		t.Errorf("Error creating RDF. Error: %s", err)
	}

	node, err := theServer.GetNode("/slugA", ldp.PreferTriples{})
	if err != nil {
		t.Errorf("Error fetching new node /slugA. Error: %s", err)
	}

	if !node.IsRdf() {
		t.Errorf("RDF source was created but not as RDF")
	}

	node2, err := theServer.CreateRdfSource("", "/", "slugA")
	if err != nil {
		t.Errorf("Error %s while attemping to create duplicate node", err)
	}

	if node2.Path() == "/slugA" {
		t.Errorf("A duplicate node was created")
	}
}

func TestReplaceRdf(t *testing.T) {
	triples := "<> xx:version \"version1\" ."
	node, err := theServer.ReplaceRdfSource(triples, "/", "rdf-test", "ignore-etag")
	log.Printf("1. %s", node.Content())
	if err != nil {
		t.Errorf("Error creating a new RDF node with replace: %s", err)
	}

	path := node.Path()[1:]
	etag := node.Etag()
	triples = "<> xx:version \"version2\" ."
	node, err = theServer.ReplaceRdfSource(triples, "/", path, etag)
	log.Printf("2. %s", node.Content())
	if err != nil {
		t.Errorf("Error replacing RDF node: %s", err)
	}

	if !node.HasTriple("xx:version", "\"version2\"") {
		t.Errorf("Error replacing RDF node. Updated triple not found")
	}

	_, err = theServer.ReplaceRdfSource(triples, "/", path, "bad-etag")
	if err != ldp.EtagMismatchError {
		t.Errorf("Failed to detect etag mismatch: %s", err)
	}

	_, err = theServer.ReplaceRdfSource(triples, "/", path, "")
	if err != ldp.EtagMissingError {
		t.Errorf("Failed to detect missing etag: %s", err)
	}
}

func TestCreateDirectContainer(t *testing.T) {
	// Create a helper node
	helperNode, err := theServer.CreateRdfSource("", "/", "other")

	// Create the direct container (pointing to the helper node)
	dcTriple1 := fmt.Sprintf("<> <%s> <%s> .\n", rdf.LdpMembershipResource, helperNode.Uri())
	dcTriple2 := fmt.Sprintf("<> <%s> <hasXYZ> .\n", rdf.LdpHasMemberRelation)
	dcTriples := dcTriple1 + dcTriple2
	dcNode, err := theServer.CreateRdfSource(dcTriples, "/", "dc")
	if err != nil {
		t.Errorf("Error creating direct container", err)
	}

	dcNode, err = theServer.GetNode(dcNode.Path(), ldp.PreferTriples{})
	if err != nil {
		t.Errorf("Error fetching direct container", err)
	}

	if !dcNode.IsBasicContainer() {
		t.Errorf("Direct container fetched but not marked as a BASIC container %s", dcNode.Content())
	}
	if !dcNode.IsDirectContainer() {
		t.Errorf("Direct container fetched but not marked as a DIRECT container %s", dcNode.Content())
	}

	// Add a child to the direct container
	childNode, err := theServer.CreateRdfSource("", dcNode.Path(), "child")
	if err != nil {
		t.Errorf("Error adding child to Direct Container %s", err)
	}

	// Reload our helper node and make sure the child is referenced on it.
	helperNode, err = theServer.GetNode(helperNode.Path(), ldp.PreferTriples{})
	if !helperNode.HasTriple("<hasXYZ>", "<"+childNode.Uri()+">") {
		t.Error("Helper node did not get new triple when adding to a Direct Container")
	}
}

func TestCreateChildRdf(t *testing.T) {
	parentNode, _ := theServer.CreateRdfSource("", "/", emptySlug)

	rdfNode, err := theServer.CreateRdfSource("", parentNode.Path(), emptySlug)
	if err != nil {
		t.Errorf("Error creating child RDF node under %s", err, parentNode.Uri())
	}

	if !strings.HasPrefix(rdfNode.Uri(), parentNode.Uri()) || rdfNode.Uri() == parentNode.Uri() {
		t.Errorf("Child URI %s does not seem to be under the parent URI %s", rdfNode.Uri(), parentNode.Uri())
	}

	invalidPath := parentNode.Path() + "/invalid"
	invalidNode, err := theServer.CreateRdfSource("", invalidPath, emptySlug)
	if err == nil {
		t.Errorf("A node was added to an invalid path %s %s", err, invalidNode.Uri())
	}

	reader := util.FakeReaderCloser{Text: "HELLO"}
	nonRdfNode, err := theServer.CreateNonRdfSource(reader, parentNode.Path(), emptySlug, "")
	if err != nil {
		t.Errorf("Error creating child non-RDF node under %s. Error: %s", parentNode.Uri(), err)
	}

	if !strings.HasPrefix(nonRdfNode.Uri(), parentNode.Uri()) || nonRdfNode.Uri() == parentNode.Uri() {
		t.Errorf("Child URI %s does not seem to be under the parent URI %s", nonRdfNode.Uri(), parentNode.Uri())
	}

	_, err = theServer.CreateRdfSource("", nonRdfNode.Path(), emptySlug)
	if err == nil {
		t.Errorf("A child was added to a non-RDF node! %s", nonRdfNode.Uri())
	}
}

func TestCreateRdfWithTriples(t *testing.T) {
	triples := "<> <b> <c> .\n<x> <y> <z> .\n"
	node, err := theServer.CreateRdfSource(triples, "/", emptySlug)
	if err != nil || !node.IsRdf() {
		t.Errorf("Error creating RDF")
	}

	node, err = theServer.GetNode(node.Path(), ldp.PreferTriples{})
	if err != nil || node.Uri() != util.UriConcat(rootUrl, node.Path()) {
		t.Errorf("err %v, uri %s", err, node.Uri())
	}

	if !node.HasTriple("<b>", "<c>") {
		t.Errorf("Blank node not handled correctly %s", node.Uri())
		t.Errorf(node.DebugString())
	}

	if node.HasTriple("x", "z") {
		t.Errorf("Unexpected tripled for new subject %s", node.Uri())
	}
}

func TestCreateNonRdf(t *testing.T) {
	reader := util.FakeReaderCloser{Text: "HELLO"}
	_, err := theServer.CreateNonRdfSource(reader, "/", "hello", "")
	if err != nil {
		t.Errorf("Error creating Non RDF")
	}

	node, err := theServer.GetNode("hello", ldp.PreferTriples{})
	if err != nil {
		t.Errorf("Could not read new Non-RDF node: %s", err)
	}

	if node.IsRdf() {
		t.Errorf("Node created as RDF instead of Non-RDF")
	}

	if node.Content() != "HELLO" {
		t.Errorf("Non-RDF content is not the expected one, %s", node.Content())
	}

	node, err = theServer.CreateNonRdfSource(reader, "/", "hello", "")
	if err != nil {
		t.Errorf("Error when attempting to create a duplicate node. Error: %s", err)
	}

	if node.Path() == "/hello" {
		t.Errorf("Failed to generate a new slug for the duplicate node")
	}
}

func TestReplaceNonRdf(t *testing.T) {
	path := "/non-rdf-test"
	reader := util.FakeReaderCloser{Text: "HELLO"}
	node, err := theServer.ReplaceNonRdfSource(reader, path, "ignored-etag", "")
	if err != nil {
		t.Errorf("Error creating a new non-RDF node with replace: %s", err)
	}

	reader2 := util.FakeReaderCloser{Text: "BYE"}
	node, err = theServer.ReplaceNonRdfSource(reader2, path, node.Etag(), "")
	if err != nil {
		t.Errorf("Error replacing Non-RDF node: %s", err)
	}

	if node.Content() != "BYE" {
		t.Errorf("Non-RDF content was not replaced. %s", node.Content())
	}

	_, err = theServer.ReplaceNonRdfSource(reader, path, "bad-etag", "")
	if err != ldp.EtagMismatchError {
		t.Errorf("Failed to detect etag mismatch: %s", err)
	}

	_, err = theServer.ReplaceNonRdfSource(reader, path, "", "")
	if err != ldp.EtagMissingError {
		t.Errorf("Failed to detect missing etag: %s", err)
	}
}

func TestPatchRdf(t *testing.T) {
	triples := "<> <p1> <o1> .\n<> <p2> <o2> .\n"
	node, _ := theServer.CreateRdfSource(triples, "/", emptySlug)
	node, _ = theServer.GetNode(node.Path(), ldp.PreferTriples{})
	if !node.HasTriple("<p1>", "<o1>") || !node.HasTriple("<p2>", "<o2>") {
		t.Errorf("Expected triple not found %s", node.Content())
	}

	newTriples := "<> <p3> <o3> .\n"
	err := node.Patch(newTriples)
	if err != nil {
		t.Errorf("Error during Patch %s", err)
	} else if !node.HasTriple("<p1>", "<o1>") ||
		!node.HasTriple("<p2>", "<o2>") ||
		!node.HasTriple("<p3>", "<o3>") {
		t.Errorf("Expected triple not after patch found %s", node.Content())
	}
}

func TestPatchNonRdf(t *testing.T) {
	reader1 := util.FakeReaderCloser{Text: "HELLO"}
	node, _ := theServer.CreateNonRdfSource(reader1, "/", emptySlug, "")
	node, _ = theServer.GetNode(node.Path(), ldp.PreferTriples{})
	if node.Content() != "HELLO" {
		t.Errorf("Unexpected non-RDF content found %s", node.Content())
	}

	if err := node.Patch("whatever"); err == nil {
		t.Errorf("Shouldn't be able to patch non-RDF")
	}
}
