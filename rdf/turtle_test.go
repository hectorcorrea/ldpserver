package rdf

import (
	// "fmt"
	"io/ioutil"
	"testing"
)

func TestOneTriple(t *testing.T) {
	parser := NewTurtleParser("<s> <p> <o> .")
	parser.Parse()
	if len(parser.Triples()) != 1 {
		t.Errorf("Error parsing triples")
	}
}

func TestTwoTriples(t *testing.T) {
	parser := NewTurtleParser("<s> <p> <o> .<s2> <p2> <o2> .")
	parser.Parse()
	if len(parser.Triples()) != 2 {
		t.Errorf("Error parsing triples %d", len(parser.Triples()))
	}
}

func TestTwoTriplesWithComments(t *testing.T) {
	parser := NewTurtleParser("#one comment \n <s> <p> <o> .\n# second comment \n<s2> <p2> <o2> . #last comment")
	err := parser.Parse()
	if err != nil {
		t.Errorf("Error parsing triples %s", err)
	}
	if len(parser.Triples()) != 2 {
		t.Errorf("Unexpected number of triples found: %d", len(parser.Triples()))
	}
}

func TestTriplesWithComma(t *testing.T) {
	test := `<s> <p> <o1> , <o2> .`
	parser := NewTurtleParser(test)
	err := parser.Parse()
	if err != nil {
		t.Errorf("Error parsing comma: %s", err)
	}
	if len(parser.Triples()) != 2 {
		t.Errorf("Incorrect number of triples: %d", len(parser.Triples()))
	}

	t1 := parser.Triples()[0].String()
	if t1 != "<s> <p> <o1> ." {
		t.Errorf("Triple 1 is incorrect: %s", t1)
	}

	t2 := parser.Triples()[1].String()
	if t2 != "<s> <p> <o2> ." {
		t.Errorf("Triple 2 is incorrect: %s", t1)
	}
}

func TestTriplesWithSemicolon(t *testing.T) {
	test := `<s> <p1> <o1> ; <p2> <o2> .`
	parser := NewTurtleParser(test)
	err := parser.Parse()
	if err != nil {
		t.Errorf("Error parsing semicolon: %s", err)
	}
	if len(parser.Triples()) != 2 {
		t.Errorf("Incorrect number of triples: %d", len(parser.Triples()))
	}

	t0 := parser.Triples()[0].String()
	if t0 != "<s> <p1> <o1> ." {
		t.Errorf("Triple 1 is incorrect: %s", t0)
	}

	t1 := parser.Triples()[1].String()
	if t1 != "<s> <p2> <o2> ." {
		t.Errorf("Triple 2 is incorrect: %s", t1)
	}
}

func TestTriplesWithCommaAndSemicolon(t *testing.T) {
	test := `<> a <http://www.w3.org/ns/ldp#RDFSource> , <http://example.com/ns#Bug> ;
	       <http://example.com/ns#severity> "High" ;
	       <http://purl.org/dc/terms/description> "Issues that need to be fixed." ;
	       <http://purl.org/dc/terms/relation> <relatedResource> ;
	       <http://purl.org/dc/terms/title> "Another bug to test." .`
	parser := NewTurtleParser(test)
	err := parser.Parse()
	if err != nil {
		t.Errorf("Error parsing text: %s", err)
	}

	if len(parser.Triples()) != 6 {
		t.Errorf("Incorrect number of triples: %d", len(parser.Triples()))
	}

	t0 := parser.Triples()[0].String()
	if t0 != "<> a <http://www.w3.org/ns/ldp#RDFSource> ." {
		t.Errorf("Triple 1 is incorrect: %s", t0)
	}

	t5 := parser.Triples()[5].String()
	if t5 != `<> <http://purl.org/dc/terms/title> "Another bug to test." .` {
		t.Errorf("Triple 6 is incorrect: %s", t5)
	}
}

func TestW3CFile(t *testing.T) {
	filename := "./w3ctest.nt"
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("Error reading w3ctest.nt file: %s", err)
	}

	text := string(bytes)
	parser := NewTurtleParser(text)
	err = parser.Parse()
	if err != nil {
		t.Errorf("Error parsing W3C ntriples text: %s", err)
	}
}

func TestBaseDirective(t *testing.T) {
	filename := "./w3cbase.nt"
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("Error reading w3cbase.nt file: %s", err)
	}

	text := string(bytes)
	parser := NewTurtleParser(text)
	err = parser.Parse()
	if err != nil {
		t.Errorf("Error parsing W3C base triples: %s", err)
	}

	for _, triple := range parser.Triples() {
		if triple.subject != "<http://ourbaseuri>" {
			t.Errorf("Base not replaced correctly for %s", triple)
		}
	}
}
