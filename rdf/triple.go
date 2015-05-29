package rdf

import "strings"
import "log"
import "fmt"

type Triple struct {
	subject   			string 			// always a URI
	predicate 			string 			// always a URI
	object    			string      // can be a URI or a literal
	isObjectLiteral bool
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
	return Triple{subject: subject, predicate: predicate, object: object, isObjectLiteral: isObjectLiteral}
}

func (t Triple) String() string {
	if t.isObjectLiteral {
		return fmt.Sprintf(`<%s> <%s> "%s" .`, t.subject, t.predicate, t.object)
	} 
	return fmt.Sprintf(`<%s> <%s> <%s> .`, t.subject, t.predicate, t.object)
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
		log.Printf("Error parsing %s. Error: %s", line, err)
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