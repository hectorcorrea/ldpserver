package rdf

import "testing"

func TestGoodTokens(t *testing.T) {
	testA := []string{"<hello>", "<hello>"}
	testB := []string{"  \t<hello>", "<hello>"}
	testC := []string{"<hello+world>", "<hello+world>"}
	testD := []string{"\"hello\"", "\"hello\""}
	testE := []string{"title+", "title"}
	testF := []string{"title", "title"}
	testG := []string{"dc:title", "dc:title"}
	tests := [][]string{testA, testB, testC, testD, testE, testF, testG}
	for _, test := range tests {
		if result, _ := GetToken(test[0]); result.value != test[1] {
			t.Errorf("GetToken failed for: (%s) (%s). Result (%s)", test[0], test[1], result.value)
		}
	}
}

func TestBadToken(t *testing.T) {
	tests := []string{"   ", "  \t ", "...", "<"}
	for _, test := range tests {
		if result, err := GetToken(test); err == nil {
			t.Errorf("Did not detect invalid token: (%s). Result: (%s)", test, result)
		}
	}
}

func TestTriple(t *testing.T) {
	parser := NewTurtleParser("<s> <p> <o> .")
	parser.Parse()
	for i, triple := range parser.Triples() {
		t.Errorf("%d %s", i, triple)
	}
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
