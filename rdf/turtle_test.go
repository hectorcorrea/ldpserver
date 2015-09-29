package rdf

import "testing"

// import "fmt"

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
		if result, _ := GetToken(test[0]); result != test[1] {
			t.Errorf("GetToken failed for: (%s) (%s). Result (%s)", test[0], test[1], result)
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

// This is not yet working
// func TestTriple(t *testing.T) {
// 	s, p, o := GetTriple("<s> <p> <o> .")
// 	if s != "<s>" || p != "<p>" || o != "<o>" {
// 		t.Errorf("Did not triple correctly: (%s) (%s) (%s)", s, p, o)
// 	}
// }
