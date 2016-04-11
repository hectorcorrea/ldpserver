package rdf

import "fmt"

type Triple struct {
	subject   string
	predicate string
	object    string
}

func NewTriple(subject, predicate, object string) Triple {
	return Triple{subject: subject, predicate: predicate, object: object}
}

func (t Triple) String() string {
	return fmt.Sprintf("%s %s %s .", t.subject, t.predicate, t.object)
}

func (t Triple) StringLn() string {
	return fmt.Sprintf("%s %s %s .\n", t.subject, t.predicate, t.object)
}

func (t Triple) Predicate() string {
	return t.predicate
}

func (t Triple) Object() string {
	return t.object
}

func (t Triple) Is(predicate string) bool {
	return t.predicate == predicate
}

func (triple *Triple) ReplaceBlankUri(blank string) {
	if triple.subject == "<>" {
		triple.subject = blank
	}
	if triple.predicate == "<>" {
		triple.predicate = blank
	}
	if triple.object == "<>" {
		triple.object = blank
	}
}

func StringToTriples(text, blank string) ([]Triple, error) {
	var triples []Triple
	parser := NewTurtleParser(text)
	err := parser.Parse()
	if err != nil {
		return triples, err
	}
	for _, triple := range parser.Triples() {
		triple.ReplaceBlankUri(blank)
	}
	return triples, nil
}
