package rdf

import "testing"
import "fmt"

func TestTripleToString(t *testing.T) {
	triple1 := NewTriple("<s>", "<p>", "<o>")
	str := fmt.Sprintf("%s", triple1)
	if str != "<s> <p> <o> ." {
		t.Errorf("Triple to string failed: %s", str)
	}

	triple2 := NewTriple("<s>", "<p>", `"o"`)
	str2 := fmt.Sprintf("%s", triple2)
	if str2 != `<s> <p> "o" .` {
		t.Errorf("Triple to string failed: %s", str2)
	}
}

func TestStringToTriple(t *testing.T) {
	validTests := []string{`<a> <b> <c> .`, `<a> <b> "c" .`}
	for _, test := range validTests {
		_, err := StringToTriples(test, "")
		if err != nil {
			t.Errorf("Failed to parse valid triple %s. Err: %s", test, err)
		}
	}

	invalidTests := []string{`<a> <3 \< 2> <no> .`, `<a> <3 < 2> <no>.\n`, `<a> <3 \> 2> <yes> .`}
	for _, test := range invalidTests {
		_, err := StringToTriples(test, "")
		if err == nil {
			t.Errorf("Failed to detect bad triple in %s.", test)
		}
	}
}

func TestReplaceBlank(t *testing.T) {
	testUri := "<http://localhost/root/>"
	triple := NewTriple("<>", "<p>", "<o>")
	triple.ReplaceBlankUri(testUri)
	if triple.subject != testUri || triple.predicate != "<p>" || triple.object != "<o>" {
		t.Error("Blank subject handled incorretly")
	}

	triple = NewTriple("<s>", "<>", "<o>")
	triple.ReplaceBlankUri(testUri)
	if triple.subject != "<s>" || triple.predicate != testUri || triple.object != "<o>" {
		t.Error("Blank predicate handled incorretly")
	}

	triple = NewTriple("<s>", "<p>", "<>")
	triple.ReplaceBlankUri(testUri)
	if triple.subject != "<s>" || triple.predicate != "<p>" || triple.object != testUri {
		t.Error("Blank object handled incorretly")
	}
}
