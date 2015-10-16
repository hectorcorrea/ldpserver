package rdf

import "testing"

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
	parser.Parse()
	if len(parser.Triples()) != 2 {
		t.Errorf("Error parsing triples %d", len(parser.Triples()))
	}
}

func TestLdpComma(t *testing.T) {
	test := `<s> <p> <o1> , <o2> .`
	parser := NewTurtleParser(test)
	err := parser.Parse2()
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

// func TestLdpTestSuiteSample(t *testing.T) {
// 	test := `<> a <http://www.w3.org/ns/ldp#RDFSource> , <http://example.com/ns#Bug> ;
//         <http://example.com/ns#severity> "High" ;
//         <http://purl.org/dc/terms/description> "Issues that need to be fixed." ;
//         <http://purl.org/dc/terms/relation> <relatedResource> ;
//         <http://purl.org/dc/terms/title> "Another bug to test." .`
// 	parser := NewTurtleParser(test)
// 	err := parser.Parse()
// 	t.Errorf("Triples\n%s\n", parser.Triples())
// 	if err != nil {
// 		t.Errorf("Error parsing LDP Test Suite demo\n%s\n", err)
// 	}
// }
