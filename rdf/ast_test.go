package rdf

import "testing"

func TestTree(t *testing.T) {
	tree := NewTree()
	subject := tree.AddNode("<s>")
	predicate := subject.AddChild("<p>")
	predicate.AddChild("<o1a>")
	predicate.AddChild("<o1b>")
	predicate2 := subject.AddChild("<p2>")
	predicate2.AddChild("<o2>")
	text := tree.Render()
	if text != "<s> <p> <o1a> .\n<s> <p> <o1b> .\n<s> <p2> <o2> .\n" {
		t.Errorf("Render gave unexpected value\n%s", text)
	}

}
