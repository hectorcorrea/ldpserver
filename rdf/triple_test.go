package rdf

import "testing"
import "fmt"

// import "log"

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

func TestStringToTriple1(t *testing.T) {
	test := "<a> <b> <c>.\n"
	triple, err := StringToTriple(test, "")
	if err != nil {
		t.Errorf("Triple for %s failed: %s", test, err)
	} else if triple.subject != "a" || triple.predicate != "b" || triple.object != "c" {
		t.Errorf("Triple %s incorrectly parsed", test)
	}
}

func TestStringToTriple2(t *testing.T) {
	test := `<a> <3 \< 2> <no>.\n`
	triple, err := StringToTriple(test, "")
	if err != nil {
		t.Errorf("Test for %s failed: %s", test, err)
	} else if triple.subject != "a" || triple.predicate != `3 \< 2` || triple.object != "no" {
		t.Errorf("Triple %s incorrectly parsed", test)
	}
}

func TestStringToTriple3(t *testing.T) {
	test := `<a> <3 < 2> <no>.\n`
	triple, err := StringToTriple(test, "")
	if err != nil {
		t.Errorf("Test for %s failed: %s", test, err)
	} else if triple.subject != "a" || triple.predicate != `3 < 2` || triple.object != "no" {
		t.Errorf("Triple %s incorrectly parsed", test)
	}
}

func TestStringToTripleBlankSubject(t *testing.T) {
	subject := "http://localhost/root/"
	test := "<> <b> <c>.\n"
	triple, err := StringToTriple(test, subject)
	if err != nil {
		t.Errorf("Triple for %s failed: %s", test, err)
	} else if triple.subject != subject || triple.predicate != "b" || triple.object != "c" {
		t.Errorf("Triple %s incorrectly parsed", test)
	}
}

func TestStringToTriple9(t *testing.T) {
	// TODO: This test fails
	// it parses the predicate as "3 \"
	// test := `<a> <3 \> 2> <yes>.\n`
	// triple, err := StringToTriple(test)
	// if err != nil {
	//   t.Errorf("Test for %s failed: %s", test, err)
	// } else if triple.subject != "a" || triple.predicate != `3 \> 2`  || triple.object != "yes" {
	//   t.Errorf("Triple %s incorrectly parsed\n %s\n %s\n %s", test, triple.subject, triple.predicate, triple. object)
	// }
}
