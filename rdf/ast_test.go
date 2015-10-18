package rdf

import "testing"

func TestTree(t *testing.T) {
	subject := NewSubjectNode("<s>")
	predicate := subject.AddPredicate("<p>")
	predicate.AddObject("<o1a>")
	predicate.AddObject("<o1b>")
	predicate2 := subject.AddPredicate("<p2>")
	predicate2.AddObject("<o2>")
	text := subject.Render()
	if text != "<s> <p> <o1a> .\n<s> <p> <o1b> .\n<s> <p2> <o2> .\n" {
		t.Errorf("Render gave unexpected value\n%s", text)
	}
}
