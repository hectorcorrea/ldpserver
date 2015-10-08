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

// func TestLdpTestSuiteSample(t *testing.T) {
// 	test := `<> a <http://www.w3.org/ns/ldp#RDFSource> , <http://example.com/ns#Bug> ;
//         <http://example.com/ns#severity> "High" ;
//         <http://purl.org/dc/terms/description> "Issues that need to be fixed." ;
//         <http://purl.org/dc/terms/relation> <relatedResource> ;
//         <http://purl.org/dc/terms/title> "Another bug to test." .`
// 	parser := NewTurtleParser(test)
// 	err := parser.Parse()
// 	if err != nil {
// 		t.Errorf("Error parsing LDP Test Suite demo\n%s\n", err)
// 	}
// }
