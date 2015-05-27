package rdf

import "testing"
import "fmt"

func TestGraphToString(t *testing.T) {
	triple1 := Triple{subject: "a", predicate: "b", object: "c"}
	triple2 := Triple{subject: "x", predicate: "y", object: "z"}
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
	triple1 := "<a> <b> <c>.\n"
	triple2 := "<x> <y> <z>.\n"
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
		t.Errorf("Is() test failed for graph [%s]", graph)
	}

	if graph.HasTriple("s", "x", "o") {
		t.Errorf("Is() test failed for graph [%s]", graph)
	}
}
