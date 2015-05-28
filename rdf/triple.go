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

// Creates a triple from a string in the following format
//    <subject> <predicate> <object> .
func StringToTriple(line, blank string) (Triple, error) {
	if len(blank) > 0 {
		line = strings.Replace(line, "<>", "<"+blank+">", -1)
	}
	// Make sure the string is a valid N-Triple.
	ntriple, err := NewNTripleFromString(line)
	if err != nil {
		log.Printf("Error parsing %s. Error: %s", line, err)
		return Triple{}, err
	}
	// Convert the N-Triple to a triple.
	return newTripleFromNTriple(ntriple), nil
}