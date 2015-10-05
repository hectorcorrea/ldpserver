package rdf

import "fmt"

type Triple struct {
	subject   string
	predicate string
	object    string
}

func NewTripleFromTokens(subject, predicate, object Token) Triple {
	return Triple{subject: subject.value, predicate: predicate.value, object: object.value}
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

func StringToTriple(line, blank string) (Triple, error) {
	parser := NewTurtleParser(line)
	triple, err := parser.ParseOne()
	if err == nil {
		triple.ReplaceBlankUri(blank)
	}
	return triple, err
}
