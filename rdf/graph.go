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

func (graph RdfGraph) IsRdfSource(subject string) bool {
	// predicate := "<" + RdfTypeUri + ">"
	object := "<" + LdpRdfSourceUri + ">"
	return graph.SubjectIs(subject, object)
}

func (graph RdfGraph) IsBasicContainer(subject string) bool {
	// predicate := "<" + RdfTypeUri + ">"
	object := "<" + LdpBasicContainerUri + ">"
	return graph.SubjectIs(subject, object)
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

func (graph RdfGraph) HasPredicate(subject, predicate string) bool {
	_, found := graph.FindTriple(subject, predicate)
	return found
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

func (graph *RdfGraph) DeleteTriple(subject, predicate, object string) bool {
	var newGraph RdfGraph
	deleted := false
	// I don't quite like this dereferrencing of the graph into triples
	// (*graph) and then getting a pointer to each individual item
	// &triples[i] but I am not sure if there is a better way.
	triples := *graph
	for i, _ := range triples {
		triple := &triples[i]
		if triple.subject == subject && triple.predicate == predicate && triple.object == object {
			deleted = true
		} else {
			newGraph = append(newGraph, triples[i])
		}
	}

	if deleted {
		*graph = newGraph
	}
	return deleted
}

func (graph *RdfGraph) appendTriple(subject, predicate, object string, recurr bool) bool {
	// I don't quite like this dereferrencing of the graph into triples
	// (*graph) and then getting a pointer to each individual item
	// &triples[i] but I am not sure if there is a better way.
	triples := *graph
	for i, _ := range triples {
		triple := &triples[i]
		if triple.subject == subject && triple.predicate == predicate && triple.object == object {
			// nothing to do
			return false
		}
	}

	if recurr {
		// "a" is an alias for RdfType
		// look to see if we can find it by alias
		switch {
		case predicate == "a":
			return graph.appendTriple(subject, "<"+RdfTypeUri+">", object, false)
		case predicate == "<"+RdfTypeUri+">":
			return graph.appendTriple(subject, "a", object, false)
		}
	}

	// Append the new triple
	newTriple := NewTriple(subject, predicate, object)
	*graph = append(*graph, newTriple)
	return true
}

func (graph *RdfGraph) Append(newGraph RdfGraph) {
	for _, triple := range newGraph {
		graph.AppendTriple(triple)
	}
}

func (graph *RdfGraph) AppendTriple(t Triple) bool {
	return graph.appendTriple(t.subject, t.predicate, t.object, true)
}

func (graph *RdfGraph) AppendTriple2(subject, predicate, object string) bool {
	return graph.appendTriple(subject, predicate, object, true)
}

func (graph RdfGraph) SubjectIs(subject, object string) bool {
	if graph.HasTriple(subject, "a", object) {
		return true
	}
	return graph.HasTriple(subject, "<"+RdfTypeUri+">", object)
}

func (graph RdfGraph) HasTriple(subject, predicate, object string) bool {
	// TODO: this does not handle the a/rdftype alias case
	// For now, user SujectIs() for those cases
	for _, triple := range graph {
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

// Set the object for a subject/predicate
// This is only useful for subject/predicates that can appear only once
// on the graph. If a subject/predicate can appear multiple times, this
// method will find and overwrite the first instance only.
func (graph *RdfGraph) SetObject(subject, predicate, object string) {
	triple, found := graph.FindTriple(subject, predicate)
	if found {
		triple.object = object
		return
	}

	// Add a new triple to the graph with the subject/predicate/object
	newTriple := NewTriple(subject, predicate, object)
	newGraph := RdfGraph{newTriple}
	graph.Append(newGraph)
}
