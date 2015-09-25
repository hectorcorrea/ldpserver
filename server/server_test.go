package server

import (
	"fmt"
	"ldpserver/rdf"
	"ldpserver/util"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var dataPath string
var theServer Server
var rootUrl = "http://localhost:9001/"
var slug = ""

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
	node, err := theServer.CreateRdfSource("", "/", slug)
	if !node.IsRdf() {
		t.Errorf("Error creating RDF. Error: %s", err)
	}

	node, err = theServer.GetNode(node.Path())
	if err != nil || node.Uri() != util.UriConcat(rootUrl, node.Path()) {
		t.Errorf("err %s, uri %s", err, node.Uri())
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

	dcNode, err = theServer.GetNode(dcNode.Path())
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
		t.Errorf("Error adding child to Direct Container", err)
	}

	// Reload our helper node and make sure the child is referenced on it.
	helperNode, err = theServer.GetNode(helperNode.Path())
	if !helperNode.HasTriple("hasXYZ", childNode.Uri()) {
		t.Error("Helper node did not get new triple when adding to a Direct Container")
	}
}

func TestCreateChildRdf(t *testing.T) {
	parentNode, _ := theServer.CreateRdfSource("", "/", slug)

	rdfNode, err := theServer.CreateRdfSource("", parentNode.Path(), slug)
	if err != nil {
		t.Errorf("Error creating child RDF node under %s", err, parentNode.Uri())
	}

	if !strings.HasPrefix(rdfNode.Uri(), parentNode.Uri()) || rdfNode.Uri() == parentNode.Uri() {
		t.Errorf("Child URI %s does not seem to be under the parent URI %s", rdfNode.Uri(), parentNode.Uri())
	}

	invalidPath := parentNode.Path() + "/invalid"
	invalidNode, err := theServer.CreateRdfSource("", invalidPath, slug)
	if err == nil {
		t.Errorf("A node was added to an invalid path %s %s", err, invalidNode.Uri())
	}

	reader := util.FakeReaderCloser{Text: "HELLO"}
	nonRdfNode, err := theServer.CreateNonRdfSource(reader, parentNode.Path(), slug)
	if err != nil {
		t.Errorf("Error creating child non-RDF node under %s", err, parentNode.Uri())
	}

	if !strings.HasPrefix(nonRdfNode.Uri(), parentNode.Uri()) || nonRdfNode.Uri() == parentNode.Uri() {
		t.Errorf("Child URI %s does not seem to be under the parent URI %s", nonRdfNode.Uri(), parentNode.Uri())
	}

	_, err = theServer.CreateRdfSource("", nonRdfNode.Path(), slug)
	if err == nil {
		t.Errorf("A child was added to a non-RDF node! %s", nonRdfNode.Uri())
	}
}

func TestCreateRdfWithTriples(t *testing.T) {
	triples := "<> <b> <c> .\n<x> <y> <z> .\n"
	node, err := theServer.CreateRdfSource(triples, "/", slug)
	if err != nil || !node.IsRdf() {
		t.Errorf("Error creating RDF")
	}

	node, err = theServer.GetNode(node.Path())
	if err != nil || node.Uri() != util.UriConcat(rootUrl, node.Path()) {
		t.Errorf("err %v, uri %s", err, node.Uri())
	}

	if !node.HasTriple("b", "c") {
		t.Errorf("Blank node not handled correctly %s", node.Uri())
	}

	if node.HasTriple("x", "z") {
		t.Errorf("Unexpected tripled for new subject %s", node.Uri())
	}
}

func TestCreateNonRdf(t *testing.T) {
	reader := util.FakeReaderCloser{Text: "HELLO"}
	node, err := theServer.CreateNonRdfSource(reader, "/", slug)
	if err != nil || node.IsRdf() {
		t.Errorf("Error creating Non RDF")
	}

	node, err = theServer.GetNode(node.Path())
	if err != nil || node.Uri() != util.UriConcat(rootUrl, node.Path()) {
		t.Errorf("err %v, uri %s", err, node.Uri())
	}

	if node.IsRdf() {
		t.Errorf("Node created as RDF instead of Non-RDF")
	}

	if node.Content() != "HELLO" {
		t.Errorf("Non-RDF content is not the expected one, %s", node.Content())
	}
}

func TestCreateDuplicate(t *testing.T) {
	_, err := theServer.CreateRdfSource("", "/", "abc")
	if err != nil {
		t.Error("Could not create new bag")
	}

	_, err = theServer.CreateRdfSource("", "/", "abc")
	if err == nil {
		t.Error("Failed to detect create on existing resource")
	}
}

func TestPatchRdf(t *testing.T) {
	triples := "<> <p1> <o1> .\n<> <p2> <o2> .\n"
	node, _ := theServer.CreateRdfSource(triples, "/", slug)
	node, _ = theServer.GetNode(node.Path())
	if !node.HasTriple("p1", "o1") || !node.HasTriple("p2", "o2") {
		t.Errorf("Expected triple not found %s", node.Content())
	}

	newTriples := "<> <p3> <o3> .\n"
	err := node.Patch(newTriples)
	if err != nil {
		t.Errorf("Error during Patch %s", err)
	} else if !node.HasTriple("p1", "o1") || !node.HasTriple("p2", "o2") || !node.HasTriple("p3", "o3") {
		t.Errorf("Expected triple not after patch found %s", node.Content())
	}
}

func TestPatchNonRdf(t *testing.T) {
	reader1 := util.FakeReaderCloser{Text: "HELLO"}
	node, _ := theServer.CreateNonRdfSource(reader1, "/", slug)
	node, _ = theServer.GetNode(node.Path())
	if node.Content() != "HELLO" {
		t.Errorf("Unexpected non-RDF content found %s", node.Content())
	}

	if err := node.Patch("whatever"); err == nil {
		t.Errorf("Shouldn't be able to patch non-RDF")
	}
}
