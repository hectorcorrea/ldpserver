package rdf

import "testing"

func TestIsIriRef(t *testing.T) {
	validIriRefs := []string{"<hello>", "<http://hello/world/>", "<hello:world>"}
	for _, iri := range validIriRefs {
		if !isIriRef(iri) {
			t.Errorf("NTriple failed to detect valid IRIREF: %s", iri)
		}
	}

	invalidChars := "<" + string([]byte{1, 2}) + ">"
	if isIriRef(invalidChars) {
		t.Errorf("NTriple failed to detect invalid characters in IRIREF")
	}

	invalidIriRefs := []string{"hello", "<he<llo>", "<hello>>", "<he{ll}o>", "<h|ello>", "hel\"lo",
		"<hello:world", "<hello world>"}
	for _, iri := range invalidIriRefs {
		if isIriRef(iri) {
			t.Errorf("NTriple failed to detect invalid IRIREF: %s", iri)
		}
	}
}

func TestIsLiteral(t *testing.T) {
	// Surrounded by quotes and properly encoded
	validLiterals := []string{`"hello"`, `"<h|ello>"`, `"<he><llo>"`, `"hello\world"`, `"hello 'world' "`}
	for _, literal := range validLiterals {
		if !isLiteral(literal) {
			t.Errorf("NTriple failed to detect valid literal: %s", literal)
		}
	}

	invalidChars := `"hello ` + string([]byte{0xA, 0xD}) + ` world"`
	if isLiteral(invalidChars) {
		t.Errorf("NTriple failed to detect invalid characters in literal")
	}

	// Not surrounded by quotes
	invalidLiterals := []string{`hello`, `"hello`, `hel"lo`, `hello"`}
	for _, literal := range invalidLiterals {
		if isLiteral(literal) {
			t.Errorf("NTriple failed to detect invalid literal: %s", literal)
		}
	}

	// Funcky encodings of \ and "
	invalidLiterals = []string{`"hello \\" world"`, `hello \" world"`}
	for _, literal := range invalidLiterals {
		if isLiteral(literal) {
			t.Errorf("NTriple failed to detect invalid literal: %s", literal)
		}
	}
}

func TestNewNTripleFromString(t *testing.T) {

	invalidStrings := []string{" .", "<s> <p> <o>", `<s> "p" "o" .`, `<s> <p> <o .`}
	for _, str := range invalidStrings {
		if _, err := NewNTripleFromString(str); err == nil {
			t.Errorf("NTriple failed to detect error in invalid string: %s", str)
		}		
	}

	validStrings := []string{"<s> <p> <o> .", `<s> <p> "o" .`}
	for _, str := range validStrings {
		if _, err := NewNTripleFromString(str); err != nil {
			t.Errorf("NTriple failed to detect valid string: %s. Error: %s", str, err)
		}		
	}

}