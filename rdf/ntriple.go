// A rudimentary implementation to support N-Triples 
// Full spec: http://www.w3.org/TR/n-triples/
package rdf

import "errors"
import "regexp"
import "strings"

type NTriple struct {
	subject   string
	predicate string
	object    string
}

// Spec: '<' ([^#x00-#x20<>"{}|^`\] | UCHAR)* '>'
var iriRefRegEx = regexp.MustCompile("<([^\x00-\x20<>\"{}|^\\`]*)>") 

// Spec: '"' ([^#x22#x5C#xA#xD] | ECHAR | UCHAR)* '"'
// Source for this reg ex: http://inamidst.com/proj/rdf/ntriples.py 
//                         (via http://lists.w3.org/Archives/Public/www-archive/2004Oct/0034.html)
var literalRegEx = regexp.MustCompile(`"([^"\\\x0A\x0D]*(?:\\.[^"\\]*)*)"`)


func NewNTriple(subject, predicate, object string) (NTriple, error) {
	if !isSubject(subject) {
		return NTriple{}, errors.New("Invalid subject: " + subject)
	}

	if !isPredicate(predicate) {
		return NTriple{}, errors.New("Invalid predicate: " + predicate)
	}

	if !isObject(object) {
		return NTriple{}, errors.New("Invalid object: " + object)
	}

	return NTriple{subject: subject, predicate: predicate, object: object}, nil
}


func NewNTripleFromString(value string) (NTriple, error) {

	if !strings.HasSuffix(value, " .") {
		return NTriple{}, errors.New("string does not end with ' .'")
	}

	subject := iriRefRegEx.FindString(value)
	if subject == "" {
		return NTriple{}, errors.New("No subject was found in string")
	}

	token2 := value[len(subject) + 1:]
	predicate := iriRefRegEx.FindString(token2)
	if predicate == "" {
		return NTriple{}, errors.New("No predicate was found in string")
	}

	token3 := token2[len(predicate) + 1:]
	object := iriRefRegEx.FindString(token3)
	if object == "" {
		object = literalRegEx.FindString(token3)
		if object == "" {
			return NTriple{}, errors.New("No object was found in string")
		}
	} 

	return NewNTriple(subject, predicate, object)
}

func isSubject(value string) bool {
	return isIriRef(value) 
}

func isPredicate(value string) bool {
	return isIriRef(value)
}

func isObject(value string) bool {
	return isIriRef(value) || isLiteral(value)
}

func isIriRef(value string) bool {
	match := iriRefRegEx.FindString(value)
	return match == value
}

func isLiteral(value string) bool {
	match := literalRegEx.FindString(value)
	return match == value
}