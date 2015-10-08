package rdf

import "testing"

func TestGoodTokens(t *testing.T) {
	testA := []string{"<hello>", "<hello>"}
	testB := []string{"  \t<hello>", "<hello>"}
	testC := []string{"<hello+world>", "<hello+world>"}
	testD := []string{"\"hello\"", "\"hello\""}
	testE := []string{"title+", "title"} // this is the correct test
	testF := []string{"title", "title"}
	testG := []string{"dc:title", "dc:title"}
	tests := [][]string{testA, testB, testC, testD, testE, testF, testG}
	for _, test := range tests {
		parser := NewTurtleParser(test[0])
		token, err := parser.GetNextToken()
		if err != nil {
			t.Errorf("Error parsing token: (%s). Error: %s.", test[0], err)
		} else if token != test[1] {
			t.Errorf("Token (%s) parsed incorrectly (%s)", test[1], token)
		}
	}
}

func TestBadTokens(t *testing.T) {
	tests := []string{"   ", "  \t ", "...", "<"}
	for _, test := range tests {
		parser := NewTurtleParser(test)
		token, err := parser.GetNextToken()
		if err == nil {
			t.Errorf("Did not detect invalid token: (%s). Result: (%s)", test, token)
		}
	}
}

func TestGoodLanguage(t *testing.T) {
	test := "\"hello\"@en-us"
	parser := NewTurtleParser(test)
	token, err := parser.GetNextToken()
	if err != nil {
		t.Errorf("Error parsing token with language: (%s). Error: %s.", test, err)
	} else if token != test {
		t.Errorf("Token with language (%s) parsed incorrectly (%s)", test, token)
	}
}

func TestBadLanguage(t *testing.T) {
	test := "\"hello\"@/en-us"
	parser := NewTurtleParser(test)
	token, err := parser.GetNextToken()
	if err != nil {
		t.Errorf("Error parsing token with bad language: (%s). Error: %s.", test, err)
	} else if token != "\"hello\"@" {
		t.Errorf("Token with bad language (%s) parsed incorrectly (%s)", test, token)
	}
}

func TestGoodType(t *testing.T) {
	test := "\"hello\"^^<http://something>"
	parser := NewTurtleParser(test)
	token, err := parser.GetNextToken()
	if err != nil {
		t.Errorf("Error parsing token with type: (%s). Error: %s.", test, err)
	} else if token != test {
		t.Errorf("Token with type (%s) parsed incorrectly (%s)", test, token)
	}
}

func TestBadType(t *testing.T) {
	testA := "\"hello\"^<http://something>"
	testB := "\"hello\"^^http://something>"
	testC := "\"hello\"^^<http://something"
	tests := []string{testA, testB, testC}
	for _, test := range tests {
		parser := NewTurtleParser(test)
		_, err := parser.GetNextToken()
		if err == nil {
			t.Errorf("Failed to detect bad data type in (%s)", test)
		}
	}
}

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
