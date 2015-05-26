package server

import (
	"ldpserver/util"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var dataPath string
var theServer Server
var rootUrl = "http://localhost:9001/"
var slug = "blog"

func init() {
	dataPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	theServer = NewServer(rootUrl, dataPath)
}

func TestCreateRdf(t *testing.T) {
	node, _ := theServer.CreateRdfSource("", "/", slug)
	if !node.IsRdf {
		t.Errorf("Error creating RDF")
	}

	path := node.Uri[len(rootUrl):] // /blog8
	node, err := theServer.GetNode(path)
	if err != nil || node.Uri != util.UriConcat(rootUrl, path) {
		t.Errorf("err %s, uri %s", err, node.Uri)
	}
}

func TestCreateChildRdf(t *testing.T) {
	parentNode, _ := theServer.CreateRdfSource("", "/", slug)
	parentPath := parentNode.Uri[len(rootUrl):] // /blog8

	rdfNode, err := theServer.CreateRdfSource("", parentPath, slug)
	if err != nil {
		t.Errorf("Error creating child RDF node under %s", err, parentNode.Uri)
	}

	if !strings.HasPrefix(rdfNode.Uri, parentNode.Uri) || rdfNode.Uri == parentNode.Uri {
		t.Errorf("Child URI %s does not seem to be under the parent URI %s", rdfNode.Uri, parentNode.Uri)
	}

	invalidPath := parentPath + "/invalid"
	invalidNode, err := theServer.CreateRdfSource("", invalidPath, slug)
	if err == nil {
		t.Errorf("A node was added to an invalid path %s %s", err, invalidNode.Uri)
	}

	reader := util.FakeReaderCloser{Text: "HELLO"}
	nonRdfNode, err := theServer.CreateNonRdfSource(reader, parentPath, slug)
	if err != nil {
		t.Errorf("Error creating child non-RDF node under %s", err, parentNode.Uri)
	}

	if !strings.HasPrefix(nonRdfNode.Uri, parentNode.Uri) || nonRdfNode.Uri == parentNode.Uri {
		t.Errorf("Child URI %s does not seem to be under the parent URI %s", nonRdfNode.Uri, parentNode.Uri)
	}

	nonRdfPath := nonRdfNode.Uri[len(rootUrl):] // /blog8/blog9
	_, err = theServer.CreateRdfSource("", nonRdfPath, slug)
	if err == nil {
		t.Errorf("A child was added to a non-RDF node! %s", nonRdfNode.Uri)
	}
}

func TestCreateRdfWithTriples(t *testing.T) {
	triples := "<> <b> <c> .\r\n<x> <y> <z> ."
	node, err := theServer.CreateRdfSource(triples, "/", slug)
	if err != nil || !node.IsRdf {
		t.Errorf("Error creating RDF")
	}

	path := node.Uri[len(rootUrl):] // /blog8
	node, err = theServer.GetNode(path)
	if err != nil || node.Uri != util.UriConcat(rootUrl, path) {
		t.Errorf("err %v, uri %s", err, node.Uri)
	}

	if !node.Is("b", "c") {
		t.Errorf("Blank node not handled correctly %s", node.Uri)
	}

	if node.Is("x", "z") {
		t.Errorf("Unexpected tripled for new subject %s", node.Uri)
	}
}

func TestCreateNonRdf(t *testing.T) {
	reader := util.FakeReaderCloser{Text: "HELLO"}
	node, err := theServer.CreateNonRdfSource(reader, "/", slug)
	if err != nil || node.IsRdf {
		t.Errorf("Error creating Non RDF")
	}

	path := node.Uri[len(rootUrl):] // /blog8
	node, err = theServer.GetNode(path)
	if err != nil || node.Uri != util.UriConcat(rootUrl, path) {
		t.Errorf("err %v, uri %s", err, node.Uri)
	}

	if node.IsRdf {
		t.Errorf("Node created as RDF instead of Non-RDF")
	}

	if node.Content() != "HELLO" {
		t.Errorf("Non-RDF content is not the expected one, %s", node.Content())
	}
}

func TestPatchRdf(t *testing.T) {
	triples := "<> <p1> <o1> .\n<> <p2> <o2> .\n"
	node, _ := theServer.CreateRdfSource(triples, "/", slug)
	path := node.Uri[len(rootUrl):]
	node, _ = theServer.GetNode(path)
	if !node.Is("p1", "o1") || !node.Is("p2", "o2") {
		t.Errorf("Expected triple not found %s", node.Content())
	}

	newTriples := "<> <p3> <o3> .\n"
	err := node.Patch(newTriples)
	if err != nil {
		t.Errorf("Error during Patch %s", err)
	} else if !node.Is("p1", "o1") || !node.Is("p2", "o2") || !node.Is("p3", "o3") {
		t.Errorf("Expected triple not after patch found %s", node.Content())
	}
}

func TestPatchNonRdf(t *testing.T) {
	reader1 := util.FakeReaderCloser{Text: "HELLO"}
	node, _ := theServer.CreateNonRdfSource(reader1, "/", slug)
	path := node.Uri[len(rootUrl):]
	node, _ = theServer.GetNode(path)
	if node.Content() != "HELLO" {
		t.Errorf("Unexpected non-RDF content found %s", node.Content())
	}

	if err := node.Patch("whatever"); err == nil {
		t.Errorf("Shouldn't be able to patch non-RDF")
	}
}
