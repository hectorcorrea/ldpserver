package rdf

import "strings"

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
				graph = append(graph, triple)
			}
		}
	}
	return graph, nil
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

func (graph RdfGraph) IsRdfSource(subject string) bool {
	return graph.HasTriple("<"+subject+">", "<"+RdfTypeUri+">", "<"+LdpRdfSourceUri+">")
}

func (graph RdfGraph) IsBasicContainer(subject string) bool {
	return graph.HasTriple(subject, RdfTypeUri, LdpBasicContainerUri)
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
		if triple.predicate == LdpMembershipResource {
			membershipResource = triple.object
		} else if triple.predicate == LdpHasMemberRelation {
			hasMemberRelation = triple.object
		}
		if membershipResource != "" && hasMemberRelation != "" {
			return membershipResource, hasMemberRelation, true
		}
	}
	return "", "", false
}

func (graph RdfGraph) HasTriple(subject, predicate, object string) bool {
	for _, triple := range graph {
		if triple.subject == subject && triple.predicate == predicate && triple.object == object {
			return true
		}
	}
	return false
}

func splitLines(text string) []string {
	allLines := strings.Split(text, "\n")
	lines := make([]string, len(allLines))
	for _, line := range allLines {
		if len(line) > 0 && line != "\n" {
			lines = append(lines, line)
		}
	}
	return lines
}
