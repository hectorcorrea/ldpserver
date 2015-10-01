package rdf

import "testing"

// func TestSandbox(t *testing.T) {
// 	parser := NewTurtleParser("title+")
// 	token, _ := parser.GetNextToken()
// 	t.Errorf("Token (%s)", token.value)
// }

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

	// for i, triple := range parser.Triples() {
	// 	t.Errorf("%d %s", i, triple)
	// }
}

// func TestTriple(t *testing.T) {
// 	testA := []string{"<s> <p> <o> .", "<s>", "<p>", "<o>"}
// 	testB := []string{"s p o .", "s", "p", "o"}
// 	testC := []string{"xx:s xx:p xx:o .", "xx:s", "xx:p", "xx:o"}
// 	testD := []string{"<http://hello> greeting \"hola\" .", "<http://hello>", "greeting", "\"hola\""}
// 	tests := [][]string{testA, testB, testC, testD}
// 	for _, test := range tests {
// 		s, p, o := GetTriple(test[0])
// 		if s != test[1] || p != test[2] || o != test[3] {
// 			t.Errorf("Did not parse triple correctly: (%s) (%s) (%s)", test[0], test[1], test[2], test[3])
// 		}
// 	}
// }
