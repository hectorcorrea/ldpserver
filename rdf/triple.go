package rdf

import "strings"
import "fmt"

type Triple struct {
	subject         string // always a URI, e.g. <hello>
	predicate       string // always a URI, e.g. <hello>
	object          string // can be a URI or a literal, e.g. <hello> or "hello"
	isObjectLiteral bool
}

func NewTripleFromTokens(subject, predicate, object Token) Triple {
	if !subject.isUri {
		// TODO: Should we allow this?
		panic("Subject is not a URI")
	}
	if !predicate.isUri {
		// TODO: Should we allow this?
		panic("Predicate is not a URI")
	}
	return newTriple(subject.value, predicate.value, object.value, object.isLiteral)
}

func NewTripleUri(subject, predicate, object string) Triple {
	return newTriple(subject, predicate, object, false)
}

func NewTripleLit(subject, predicate, object string) Triple {
	return newTriple(subject, predicate, object, true)
}

func newTripleFromNTriple(ntriple NTriple) Triple {
	return newTriple(ntriple.Subject(), ntriple.Predicate(), ntriple.Object(), ntriple.IsObjectLiteral())
}

func newTriple(subject, predicate, object string, isObjectLiteral bool) Triple {
	// Temporary hack while I fix the code that is passing triples
	// without <> or ""
	if !strings.HasPrefix(subject, "<") {
		subject = "<" + subject + ">"
	}
	if !strings.HasPrefix(predicate, "<") {
		predicate = "<" + predicate + ">"
	}
	if isObjectLiteral {
		if !strings.HasPrefix(object, "\"") {
			object = "\"" + object + "\""
		}
	} else {
		if !strings.HasPrefix(object, "<") {
			object = "<" + object + ">"
		}
	}
	return Triple{subject: subject, predicate: predicate, object: object, isObjectLiteral: isObjectLiteral}
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
