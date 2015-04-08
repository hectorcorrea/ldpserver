package server

import "testing"
import "strings"
import "ldpserver/ldp"

var dataPath = "/Users/hector/dev/gotest/src/ldpserver/data_test"
var rootUrl = "http://localhost:9001/"
var settings ldp.Settings

func init() {
	settings = ldp.SettingsNew(dataPath, rootUrl)
	ldp.CreateRoot(settings)
}

func TestCreateRdf(t *testing.T) {
	node, _ := CreateRdfSource(settings, "", "/")
	if !node.IsRdf {
		t.Errorf("Error creating RDF")
	}

	path := node.Uri[len(rootUrl):] // /blog8
	node, err := GetNode(settings, path)
	if err != nil || node.Uri != ldp.UriConcat(rootUrl, path) {
		t.Errorf("err %s, uri %s", err, node.Uri)
	}
}

func TestCreateChildRdf(t *testing.T) {
	parentNode, _ := CreateRdfSource(settings, "", "/")
	parentPath := parentNode.Uri[len(rootUrl):] // /blog8

	rdfNode, err := CreateRdfSource(settings, "", parentPath)
	if err != nil {
		t.Errorf("Error creating child RDF node under %s", err, parentNode.Uri)
	}

	if !strings.HasPrefix(rdfNode.Uri, parentNode.Uri) || rdfNode.Uri == parentNode.Uri {
		t.Errorf("Child URI %s does not seem to be under the parent URI %s", rdfNode.Uri, parentNode.Uri)
	}

	invalidPath := parentPath + "/invalid"
	invalidNode, err := CreateRdfSource(settings, "", invalidPath)
	if err == nil {
		t.Errorf("A node was added to an invalid path %s %s", err, invalidNode.Uri)
	}

	reader := ldp.FakeReaderCloser{Text: "HELLO"}
	nonRdfNode, err := CreateNonRdfSource(settings, reader, parentPath)
	if err != nil {
		t.Errorf("Error creating child non-RDF node under %s", err, parentNode.Uri)
	}

	if !strings.HasPrefix(nonRdfNode.Uri, parentNode.Uri) || nonRdfNode.Uri == parentNode.Uri {
		t.Errorf("Child URI %s does not seem to be under the parent URI %s", nonRdfNode.Uri, parentNode.Uri)
	}

	nonRdfPath := nonRdfNode.Uri[len(rootUrl):] // /blog8/blog9
	_, err = CreateRdfSource(settings, "", nonRdfPath)
	if err == nil {
		t.Errorf("A child was added to a non-RDF node! %s", nonRdfNode.Uri)
	}
}

func TestCreateRdfWithTriples(t *testing.T) {
	triples := "<> <b> <c> .\r\n<x> <y> <z> ."
	node, err := CreateRdfSource(settings, triples, "/")
	if err != nil || !node.IsRdf {
		t.Errorf("Error creating RDF")
	}

	path := node.Uri[len(rootUrl):] // /blog8
	node, err = GetNode(settings, path)
	if err != nil || node.Uri != ldp.UriConcat(rootUrl, path) {
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
	reader := ldp.FakeReaderCloser{Text: "HELLO"}
	node, err := CreateNonRdfSource(settings, reader, "/")
	if err != nil || node.IsRdf {
		t.Errorf("Error creating Non RDF")
	}

	path := node.Uri[len(rootUrl):] // /blog8
	node, err = GetNode(settings, path)
	if err != nil || node.Uri != ldp.UriConcat(rootUrl, path) {
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
	node, _ := CreateRdfSource(settings, triples, "/")
	path := node.Uri[len(rootUrl):]
	node, _ = GetNode(settings, path)
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
	reader1 := ldp.FakeReaderCloser{Text: "HELLO"}
	node, _ := CreateNonRdfSource(settings, reader1, "/")
	path := node.Uri[len(rootUrl):]
	node, _ = GetNode(settings, path)
	if node.Content() != "HELLO" {
		t.Errorf("Unexpected non-RDF content found %s", node.Content())
	}

	if err := node.Patch("whatever"); err == nil {
		t.Errorf("Shouldn't be able to patch non-RDF")
	}
}
