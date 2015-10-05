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
		} else if token.value != test[1] {
			t.Errorf("Token (%s) parsed incorrectly (%s)", test[1], token.value)
		}
	}
}

func TestBadTokens(t *testing.T) {
	tests := []string{"   ", "  \t ", "...", "<"}
	for _, test := range tests {
		parser := NewTurtleParser(test)
		token, err := parser.GetNextToken()
		if err == nil {
			t.Errorf("Did not detect invalid token: (%s). Result: (%s)", test, token.value)
		}
	}
}

func TestGoodLanguage(t *testing.T) {
	test := "\"hello\"@en-us"
	parser := NewTurtleParser(test)
	token, err := parser.GetNextToken()
	if err != nil {
		t.Errorf("Error parsing token with language: (%s). Error: %s.", test, err)
	} else if token.value != test {
		t.Errorf("Token with language (%s) parsed incorrectly (%s)", test, token.value)
	}
}

func TestBadLanguage(t *testing.T) {
	test := "\"hello\"@/en-us"
	parser := NewTurtleParser(test)
	token, err := parser.GetNextToken()
	if err != nil {
		t.Errorf("Error parsing token with bad language: (%s). Error: %s.", test, err)
	} else if token.value != "\"hello\"@" {
		t.Errorf("Token with bad language (%s) parsed incorrectly (%s)", test, token.value)
	}
}

func TestGoodType(t *testing.T) {
	test := "\"hello\"^^<http:/something>"
	parser := NewTurtleParser(test)
	token, err := parser.GetNextToken()
	if err != nil {
		t.Errorf("Error parsing token with type: (%s). Error: %s.", test, err)
	} else if token.value != test {
		t.Errorf("Token with type (%s) parsed incorrectly (%s)", test, token.value)
	}
}

func TestPeek(t *testing.T) {
	parser := NewTurtleParser("abc")
	if _, nextChar := parser.peek(); nextChar != 'b' {
		t.Errorf("Error on first peek")
	}
	parser.advance()
	if _, nextChar := parser.peek(); nextChar != 'c' {
		t.Errorf("Error on second peek")
	}
	parser.advance()
	if canPeek, _ := parser.peek(); canPeek == true {
		t.Errorf("Failed to detect that it cannot peek anymore")
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
