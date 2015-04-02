package ldp

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

func StringToGraph(theString, rootUri string) (RdfGraph, error) {
	var graph RdfGraph
	if len(theString) == 0 {
		return graph, nil
	}
	lines := splitLines(theString)
	for _, line := range lines {
		if line != "\n" && line != "" {
			// log.Printf("Evaluating %s", line)
			triple, err := StringToTriple(line, rootUri)
			if err != nil {
				return nil, err
			}
			graph = append(graph, triple)
		}
	}
	return graph, nil
}

func (graph *RdfGraph) Append(newGraph RdfGraph) {
	// TODO: Is it OK to duplicate a triple (same subject, pred, and object)
	// or should Append remove duplicates?
	for _, triple := range newGraph {
		*graph = append(*graph, triple)
	}
}

func (graph RdfGraph) IsRdfSource(subject string) bool {
	return graph.Is(subject, RdfTypeUri, LdpRdfSourceUri)
}

func (graph RdfGraph) IsBasicContainer(subject string) bool {
	return graph.Is(subject, RdfTypeUri, LdpBasicContainerUri)
}

func (graph RdfGraph) Is(subject, predicate, object string) bool {
	for _, triple := range graph {
		if triple.subject == subject && triple.predicate == predicate && triple.object == object {
			return true
		}
	}
	return false
}
