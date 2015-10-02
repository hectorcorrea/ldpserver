package rdf

import "strings"
import "log"
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
	return t.String() + "\n"
}

// Creates a triple from a string. We assume the string is in N-Triple format
// and thefore looks like this:
//
//    <subject> <predicate> <object> .
//
// or like this:
//
//    <subject> <predicate> "object" .
//
func StringToTriple(line, blank string) (Triple, error) {
	if len(blank) > 0 {
		line = replaceBlanksInTriple(line, blank)
	}

	// Parse the line in N-Triple format into an NTriple object.
	ntriple, err := NewNTripleFromString(line)
	if err != nil {
		log.Printf("Error parsing triple %s. Error: %s \n", line, err)
		return Triple{}, err
	}

	// Convert the N-Triple to a triple.
	return newTripleFromNTriple(ntriple), nil
}

func replaceBlanksInTriple(line, blank string) string {
	isEmptySubject := strings.HasPrefix(line, "<>")
	if isEmptySubject {
		line = "<" + blank + ">" + line[2:]
	}

	// notice that we purposefully don't replace the <>
	// if it is found in the predicate.

	isEmptyObject := strings.HasSuffix(line, "<> .")
	if isEmptyObject {
		line = line[:len(line)-4] + "<" + blank + "> ."
	}
	return line
}
