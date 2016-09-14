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
		tokenizer := NewTokenizer(test[0])
		token, err := tokenizer.GetNextToken()
		if err != nil {
			t.Errorf("Error parsing token: (%s). Error: %s.", test[0], err)
		} else if token != test[1] {
			t.Errorf("Token (%s) parsed incorrectly (%s)", test[1], token)
		}
	}
}

func TestBadTokens(t *testing.T) {
	tests := []string{"<", ">", "{", "}", `"aaa`}
	for _, test := range tests {
		tokenizer := NewTokenizer(test)
		token, err := tokenizer.GetNextToken()
		if err == nil {
			t.Errorf("Did not detect invalid token: (%s). Result: (%s)", test, token)
		}
	}
}

func TestGoodLanguage(t *testing.T) {
	test := "\"hello\"@en-us"
	tokenizer := NewTokenizer(test)
	token, err := tokenizer.GetNextToken()
	if err != nil {
		t.Errorf("Error parsing token with language: (%s). Error: %s.", test, err)
	} else if token != test {
		t.Errorf("Token with language (%s) parsed incorrectly (%s)", test, token)
	}
}

func TestBadLanguage(t *testing.T) {
	test := "\"hello\"@/en-us"
	tokenizer := NewTokenizer(test)
	token, err := tokenizer.GetNextToken()
	if err != nil {
		t.Errorf("Error parsing token with bad language: (%s). Error: %s.", test, err)
	} else if token != "\"hello\"@" {
		t.Errorf("Token with bad language (%s) parsed incorrectly (%s)", test, token)
	}
}

func TestGoodType(t *testing.T) {
	test := "\"hello\"^^<http://something>"
	tokenizer := NewTokenizer(test)
	token, err := tokenizer.GetNextToken()
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
		tokenizer := NewTokenizer(test)
		_, err := tokenizer.GetNextToken()
		if err == nil {
			t.Errorf("Failed to detect bad data type in (%s)", test)
		}
	}
}
