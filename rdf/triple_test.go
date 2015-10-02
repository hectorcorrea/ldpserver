package rdf

import "testing"
import "fmt"

func TestTripleToString(t *testing.T) {
	triple1 := NewTripleUri("<s>", "<p>", "<o>")
	str := fmt.Sprintf("%s", triple1)
	if str != "<s> <p> <o> ." {
		t.Errorf("Triple to string failed: %s", str)
	}

	triple2 := NewTripleLit("<s>", "<p>", `"o"`)
	str2 := fmt.Sprintf("%s", triple2)
	if str2 != `<s> <p> "o" .` {
		t.Errorf("Triple to string failed: %s", str2)
	}
}

func TestStringToTriple(t *testing.T) {
	validTests := []string{`<a> <b> <c> .`, `<a> <b> "c" .`}
	for _, test := range validTests {
		_, err := StringToTriple(test, "")
		if err != nil {
			t.Errorf("Failed to parse valid triple %s. Err: %s", test, err)
		}
	}

	invalidTests := []string{`<a> <3 \< 2> <no> .`, `<a> <3 < 2> <no>.\n`, `<a> <3 \> 2> <yes> .`}
	for _, test := range invalidTests {
		_, err := StringToTriple(test, "")
		if err == nil {
			t.Errorf("Failed to detect bad triple in %s.", test)
		}
	}
}

func TestStringToTripleBlankSubject(t *testing.T) {
	testUri := "http://localhost/root/"
	triple, _ := StringToTriple("<> <p> <o> .", testUri)
	if triple.subject != testUri || triple.predicate != "p" || triple.object != "o" {
		t.Error("Triple with blank subject was parsed incorrectly")
	}

	triple, _ = StringToTriple("<s> <p> <> .", testUri)
	if triple.subject != "s" || triple.predicate != "p" || triple.object != testUri {
		t.Error("Triple with blank object was parsed incorrectly")
	}

	triple, err := StringToTriple("<s> <> <o> .", testUri)
	if err == nil {
		t.Error("Triple with blank predicate parsed incorrectly")
	}
}
