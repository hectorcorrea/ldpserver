// A rudimentary implementation of N-Triples 
// Full spec: http://www.w3.org/TR/n-triples/
//
// Supported:
//		basic triples like <s> <p> <o> and <s> <p> "o"
//
// Not supported:
//		blank nodes
//		language in literals e.g. "hola"@es
//		types in literals e.g. "hello"^^<http://www.w3.org/2001/XMLSchema#string>
//		comments
//
package rdf

import "errors"
import "regexp"
import "strings"
// import "log"

type NTriple struct {
	subject   string
	isSubjecttUri bool
	predicate string
	object    string
	isObjectLiteral bool
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

	return NTriple{subject: subject, predicate: predicate, object: object, isObjectLiteral: true}, nil
}


func (triple NTriple) Subject() string {
	return triple.subject
}

func (triple NTriple) Predicate() string {
	return triple.predicate
}

func (triple NTriple) Object() string {
	return triple.object
}

func (triple NTriple) IsObjectLiteral() bool {
	return triple.isObjectLiteral
}

func (triple NTriple) IsObjectUri() bool {
	return !triple.isObjectLiteral
}

func NewNTripleFromString(value string) (NTriple, error) {
	if !strings.HasSuffix(value, " .") {
		return NTriple{}, errors.New("string does not end with ' .'")
	}

	subject := extractIriRef(value)
	if subject == "" {
		return NTriple{}, errors.New("No subject was found in string")
	}

	token2 := value[len(subject) + 1:]
	predicate := extractIriRef(token2)
	if predicate == "" {
		return NTriple{}, errors.New("No predicate was found in string")
	}

	token3 := token2[len(predicate) + 1:]
	isObjectLiteral := false
	object := extractIriRef(token3)
	if object == "" {
		object = extractLiteral(token3)
		if object == "" {
			return NTriple{}, errors.New("No object was found in string")
		}
		isObjectLiteral = true
	}

	return newNTriple(subject, predicate, object, isObjectLiteral), nil
}

func newNTriple(subject, predicate, object string, isObjectLiteral bool) NTriple {
	var triple NTriple
	triple.subject = stripDelimiters(subject)
	triple.isSubjecttUri = true
	triple.predicate = stripDelimiters(predicate)
	triple.isSubjecttUri = true
	triple.object = stripDelimiters(object)
	triple.isObjectLiteral = isObjectLiteral	
	return triple
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

func extractIriRef(value string) string {
	match := iriRefRegEx.FindString(value)
	if strings.HasPrefix(value, match) {
		return match
	}
	return ""
}

func extractLiteral(value string) string {
	match := literalRegEx.FindString(value)
	if strings.HasPrefix(value, match) {
		return match
	}
	return ""
}