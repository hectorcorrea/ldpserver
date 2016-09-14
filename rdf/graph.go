package rdf

import "log"

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
	return graph.HasTriple(subject, "a", "<"+LdpRdfSourceUri+">")
}

func (graph RdfGraph) IsBasicContainer(subject string) bool {
	return graph.HasTriple(subject, "a", "<"+LdpBasicContainerUri+">")
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
	_, found := graph.FindPredicate(subject, predicate)
	return found
}

func (graph *RdfGraph) FindPredicate(subject, predicate string) (*Triple, bool) {
	return graph.findPredicate(subject, predicate, true)
}

func (graph *RdfGraph) findPredicate(subject, predicate string, recurr bool) (*Triple, bool) {
	for i, triple := range *graph {
		if triple.subject == subject && triple.predicate == predicate {
			// return a reference to the original triple
			return &(*graph)[i], true
		}
	}
	if recurr {
		// "a" is an alias for RdfType
		// look to see if we can find it by alias
		switch {
		case predicate == "a":
			return graph.findPredicate(subject, "<"+RdfTypeUri+">", false)
		case predicate == "<"+RdfTypeUri+">":
			return graph.findPredicate(subject, "a", false)
		}
	}
	return nil, false
}

func (graph *RdfGraph) findTriple(subject, predicate, object string, recurr bool) (*Triple, bool) {
	for i, t := range *graph {
		if t.subject == subject && t.predicate == predicate && t.object == object {
			// return a reference to the original triple
			return &(*graph)[i], true
		}
	}
	// "a" is an alias for RdfType
	// look to see if we can find it by alias
	if recurr {
		switch {
		case predicate == "a":
			return graph.findTriple(subject, "<"+RdfTypeUri+">", object, false)
		case predicate == "<"+RdfTypeUri+">":
			return graph.findTriple(subject, "a", object, false)
		}
	}
	return nil, false
}

func (graph *RdfGraph) DeleteTriple(subject, predicate, object string) bool {
	var newGraph RdfGraph
	deleted := false
	for _, triple := range *graph {
		// This does not handle the predicate a vs RdfType like find does. Should it?
		if triple.subject == subject && triple.predicate == predicate && triple.object == object {
			// don't add it to the new graph
			deleted = true
		} else {
			newGraph = append(newGraph, triple)
		}
	}

	if deleted {
		*graph = newGraph
	}
	return deleted
}

func (graph *RdfGraph) appendTriple(subject, predicate, object string) bool {
	t, found := graph.findTriple(subject, predicate, object, true)
	if found {
		if t.predicate == "a" && predicate != "a" {
			t.predicate = predicate
			log.Printf("**> replaced a with %s", predicate)
		}
		// nothing to do
		return false
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
	return graph.appendTriple(t.subject, t.predicate, t.object)
}

func (graph *RdfGraph) AppendTripleStr(subject, predicate, object string) bool {
	return graph.appendTriple(subject, predicate, object)
}

func (graph RdfGraph) HasTriple(subject, predicate, object string) bool {
	_, found := graph.findTriple(subject, predicate, object, true)
	return found
}

func (graph RdfGraph) GetObject(subject, predicate string) (string, bool) {
	triple, found := graph.FindPredicate(subject, predicate)
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
	triple, found := graph.FindPredicate(subject, predicate)
	if found {
		triple.object = object
		return
	}

	// Add a new triple to the graph with the subject/predicate/object
	newTriple := NewTriple(subject, predicate, object)
	newGraph := RdfGraph{newTriple}
	graph.Append(newGraph)
}
