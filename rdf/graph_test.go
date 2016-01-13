package rdf

import "testing"
import "fmt"

func TestGraphToString(t *testing.T) {
	triple1 := Triple{subject: "<a>", predicate: "<b>", object: "<c>"}
	triple2 := Triple{subject: "<x>", predicate: "<y>", object: "<z>"}
	var graph RdfGraph
	graph = append(graph, triple1, triple2)
	str := fmt.Sprintf("%s", graph)
	if str != "<a> <b> <c> .\n<x> <y> <z> .\n" {
		t.Errorf("Graph to string failed: %s", str)
	}
}

func TestStringToGraph(t *testing.T) {
	var graph RdfGraph
	var err error
	triple1 := "<a> <b> <c> .\n"
	triple2 := "<x> <y> <z> .\n"
	graph, err = StringToGraph(triple1+triple2, "")
	if err != nil || len(graph) != 2 {
		t.Errorf("Unexpected number of triples found: %d %s", len(graph), err)
	}

	graph, err = StringToGraph("\n"+triple1+"\n"+triple2+"\n", "")
	if err != nil || len(graph) != 2 {
		t.Errorf("Failed to remove empty lines %d %s", len(graph), err)
	}
}

func TestAppend(t *testing.T) {
	var graph2 RdfGraph
	t1 := Triple{subject: "s", predicate: "p", object: "o"}
	graph1 := RdfGraph{t1}
	for _, x := range graph1 {
		graph2 = append(graph2, x)
	}

	if len(graph2) != 1 {
		t.Errorf("Graph not appended: [%s]", graph2)
	}
}

func TestHasTriple(t *testing.T) {
	triple := Triple{subject: "s", predicate: "p", object: "o"}
	graph := RdfGraph{triple}

	if !graph.HasTriple("s", "p", "o") {
		t.Errorf("HasTriple test failed for graph [%s]", graph)
	}

	if graph.HasTriple("s", "x", "o") {
		t.Errorf("HasTriple test failed for graph [%s]", graph)
	}
}

func TestFindPredicate(t *testing.T) {
	triple := Triple{subject: "s", predicate: "a", object: "something"}
	graph := RdfGraph{triple, triple}

	if _, found := graph.FindPredicate("s", "a"); !found {
		t.Errorf("FindPredicate test failed for valid triple")
	}

	if _, found := graph.FindPredicate("s", "b"); found {
		t.Errorf("FindPredicate test failed for invalid triple")
	}
}

func TestFindPredicateAliasA(t *testing.T) {
	triple := Triple{subject: "s", predicate: "a", object: "something"}
	graph := RdfGraph{triple, triple}

	if _, found := graph.FindPredicate("s", "<"+RdfTypeUri+">"); !found {
		t.Errorf("FindPredicate test failed when using rdf type in fullname")
	}
}

func TestFindPredicateAliasRdfType(t *testing.T) {
	triple := Triple{subject: "s", predicate: "<" + RdfTypeUri + ">", object: "something"}
	graph := RdfGraph{triple, triple}

	if _, found := graph.FindPredicate("s", "a"); !found {
		t.Errorf("FindPredicate test failed when using 'a' rather than rdf type fullname")
	}
}

func TestSetObject(t *testing.T) {
	triple := Triple{subject: "s", predicate: "p", object: "o"}
	graph := RdfGraph{triple}

	graph.SetObject("s", "p", "o2")
	if graph.HasTriple("s", "p", "o") {
		t.Errorf("SetObject found the original triple (after it was replaced)")
	}

	if !graph.HasTriple("s", "p", "o2") {
		t.Errorf("SetObject did not find triple with new value")
	}

	graph.SetObject("s", "p2", "o3")
	if !graph.HasTriple("s", "p2", "o3") {
		t.Errorf("SetObject did not find new triple")
	}

	if !graph.HasTriple("s", "p", "o2") {
		t.Errorf("SetObject did not triple with new value (after adding new triple)")
	}
}

func TestDeleteTriple(t *testing.T) {
	t1 := Triple{subject: "s1", predicate: "p1", object: "o1"}
	t2 := Triple{subject: "s2", predicate: "p2", object: "o2"}
	t3 := Triple{subject: "s3", predicate: "p3", object: "o3"}
	graph := RdfGraph{t1, t2, t3}

	deleted := graph.DeleteTriple("s2", "p2", "o2")
	if !deleted {
		t.Errorf("Did not delete triple from graph")
	}

	if graph.HasTriple("s2", "p2", "o2") {
		t.Errorf("Deleted triple found in graph")
	}

	deleted = graph.DeleteTriple("s2", "p2", "o2")
	if deleted {
		t.Errorf("Delete triple deleted a non-existing triple")
	}
}
