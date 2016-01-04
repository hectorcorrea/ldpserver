package rdf

// import "log"

type RdfGraph []Triple

func (triples RdfGraph) String() string {
	theString := ""
	for _, triple := range triples {
		theString += triple.StringLn()
	}
	return theString
}

func StringToGraph(theString, rootUri string) (RdfGraph, error) {
	var err error
	var graph RdfGraph
	if len(theString) > 0 {
		parser := NewTurtleParser(theString)
		err = parser.Parse()
		if err == nil {
			for _, triple := range parser.Triples() {
				triple.ReplaceBlankUri(rootUri)
				graph = append(graph, triple)
			}
		}
	}
	return graph, err
}

func (graph *RdfGraph) Append(newGraph RdfGraph) {
	// TODO: Is it OK to duplicate a triple (same subject, pred, and object)
	// or should Append remove duplicates?
	// TODO: There are some triples that cannot be duplicated (e.g. direct container definition triples)
	//       but perhaps that should be validated in the LDP module, not on the RDF module.
	for _, triple := range newGraph {
		*graph = append(*graph, triple)
	}
}

func (graph *RdfGraph) AppendTriple(triple Triple) {
	newGraph := RdfGraph{triple}
	graph.Append(newGraph)
}

func (graph RdfGraph) IsRdfSource(subject string) bool {
	predicate := "<" + RdfTypeUri + ">"
	object := "<" + LdpRdfSourceUri + ">"
	return graph.HasTriple(subject, predicate, object)
}

func (graph RdfGraph) IsBasicContainer(subject string) bool {
	predicate := "<" + RdfTypeUri + ">"
	object := "<" + LdpBasicContainerUri + ">"
	return graph.HasTriple(subject, predicate, object)
}

func (graph RdfGraph) IsDirectContainer() bool {
	_, _, isDirectContainer := graph.GetDirectContainerInfo()
	return isDirectContainer
}

func (graph RdfGraph) GetDirectContainerInfo() (string, string, bool) {
	// TODO: validate only one instance of each these predicates is found on the graph
	// (perhas the validation should only be when adding/updating triples)
	membershipResource := ""
	hasMemberRelation := ""
	for _, triple := range graph {
		switch triple.predicate {
		case "<" + LdpMembershipResource + ">":
			membershipResource = triple.object
		case "<" + LdpHasMemberRelation + ">":
			hasMemberRelation = triple.object
		}
		if membershipResource != "" && hasMemberRelation != "" {
			return membershipResource, hasMemberRelation, true
		}
	}
	return "", "", false
}

func (graph RdfGraph) HasPredicate(predicate string) bool {
	for _, triple := range graph {
		if triple.predicate == predicate {
			return true
		}
	}
	return false
}

func (graph *RdfGraph) FindTriple(subject, predicate string) (*Triple, bool) {
	// I don't quite like this dereferrencing of the graph into triples
	// (*graph) and then getting a pointer to each individual item
	// &triples[i] but I am not sure if there is a better way.
	triples := *graph
	for i, _ := range triples {
		triple := &triples[i]
		if triple.subject == subject && triple.predicate == predicate {
			return triple, true
		}
	}
	// "a" is an alias for RdfType
	// look to see if we can find it by alias
	switch {
	case predicate == "a":
		return graph.FindTriple(subject, "<"+RdfTypeUri+">")
	case predicate == "<"+RdfTypeUri+">":
		return graph.FindTriple(subject, "a")
	}
	return nil, false
}

func (graph RdfGraph) HasTriple(subject, predicate, object string) bool {
	for _, triple := range graph {
		// TODO: this does not handle the a/rdftype alias case
		found := (triple.subject == subject) && (triple.predicate == predicate) && (triple.object == object)
		if found {
			return true
		}
	}
	return false
}

func (graph RdfGraph) GetObject(subject, predicate string) (string, bool) {
	triple, found := graph.FindTriple(subject, predicate)
	if found {
		return triple.object, true
	}
	return "", false
}

func (graph *RdfGraph) SetObject(subject, predicate, object string) {
	var t *Triple
	t, found := graph.FindTriple(subject, predicate)
	if found {
		t.object = object
		return
	}

	// Add a new triple to the graph with the subject/predicate/object
	newTriple := NewTriple(subject, predicate, object)
	graph.AppendTriple(newTriple)
}
