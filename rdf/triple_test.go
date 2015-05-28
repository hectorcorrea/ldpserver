package rdf

import "testing"
import "fmt"

func TestTripleToString(t *testing.T) {
	triple := Triple{subject: "a", predicate: "b", object: "c"}
	str := fmt.Sprintf("%s", triple)
	if str != "<a> <b> <c> ." {
		t.Errorf("Triple to string failed: %s", str)
	}
}

func TestEncode(t *testing.T) {
	triple := Triple{subject: "a<b", predicate: `the "normal" pred`, object: "c"}
	str := fmt.Sprintf("%s", triple)
	if str != `<a\<b> <the \"normal\" pred> <c> .` {
		t.Errorf("Triple to string failed: %s", str)
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
	subject := "http://localhost/root/"
	test := "<> <b> <c> ."
	triple, _ := StringToTriple(test, subject)
	if triple.subject != subject || triple.predicate != "b" || triple.object != "c" {
		t.Error("Triple with blank subject was parsed incorrectly")
	}
}