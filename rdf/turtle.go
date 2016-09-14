// A basic RDF Turtle parser
// http://www.w3.org/TR/turtle/
//
// TurtleParser is the parser which uses Tokenizer to
// break down the text into meaningful tokens (URIs, strings,
// separators, et cetera.)

// Tokenizer in turn uses Scanner to handle the character by
// character operations.
//
// TurtleParser uses a tree-like structure (via SubjectNode
// and PredicateNode) to keep track of the subject, predicate,
// and object values as they are parsed. This structure allows
// us to parse multi-predicate (;) and multi-object (,) triples.
//
// Sample usage:
//     parser := NewTurtleParser("<s> <p1> <o1> , <o2> ; <p2> <o3> .")
//     err := parser.Parse()
//     for i, triple := range parser.Triples() {
//         log.Printf("Triple %d: %s", i, triple)
//     }
// Gives:
//     Triple 0: <s> <p1> <o1> .
//     Triple 1: <s> <p1> <o2> .
//     Triple 2: <s> <p2> <o3> .
//
package rdf

import (
	"errors"
	// "log"
	"strings"
)

type Directive struct {
	name  string
	value string
}

type TurtleParser struct {
	tokenizer  Tokenizer
	triples    []Triple
	directives []Directive
}

func NewTurtleParser(text string) TurtleParser {
	tokenizer := NewTokenizer(text)
	parser := TurtleParser{tokenizer: tokenizer}
	return parser
}

func (parser *TurtleParser) Parse() error {
	for parser.tokenizer.CanRead() {
		err := parser.parseNextTriples()
		if err != nil {
			return err
		}
		parser.tokenizer.AdvanceWhiteSpace()
	}

	parser.applyBaseDirective()
	return nil
}

func (parser TurtleParser) Triples() []Triple {
	return parser.triples
}

func (parser *TurtleParser) applyBaseDirective() {
	if len(parser.directives) == 0 {
		return
	}

	if parser.directives[0].name != "@base" {
		// unknown directive
		return
	}

	base := parser.directives[0]
	for i, triple := range parser.triples {
		if triple.subject == "<>" {
			parser.triples[i].subject = base.value
		}
		if triple.object == "<>" {
			parser.triples[i].object = base.value
		}
	}
}

func (parser *TurtleParser) parseNextTriples() error {
	var err error
	var token string

	for err == nil && parser.tokenizer.CanRead() {
		token, err = parser.tokenizer.GetNextToken()
		if err != nil || token == "" {
			break
		}

		isDirective := strings.HasPrefix(token, "@")
		if isDirective {
			err = parser.parseNextDirective(token)
			continue
		}

		// triples
		subject := NewSubjectNode(token)
		err = parser.parsePredicates(&subject)
		if err == nil {
			for _, triple := range subject.RenderTriples() {
				parser.triples = append(parser.triples, triple)
			}
		}

	}
	return err
}

func (parser *TurtleParser) parseNextDirective(name string) error {
	if !parser.tokenizer.CanRead() {
		return errors.New("No value found for directive (" + name + ")")
	}

	value, err := parser.tokenizer.GetNextToken()
	if err != nil {
		return err
	}

	token, err := parser.tokenizer.GetNextToken()
	if token != "." {
		return errors.New("Could not find end of directive (" + name + ")")
	}

	if err == nil {
		directive := Directive{name: name, value: value}
		parser.directives = append(parser.directives, directive)
	}
	return err
}

func (parser *TurtleParser) parsePredicates(subject *SubjectNode) error {
	var err error
	var token string

	for err == nil && parser.tokenizer.CanRead() {
		token, err = parser.tokenizer.GetNextToken()
		if err != nil || token == "." {
			// we are done
			break
		}
		predicate := subject.AddPredicate(token)
		token, err = parser.parseObjects(predicate)
		if err != nil {
			break
		} else if token == "." {
			// we are done, next triple will be for a different subject
			break
		} else if token == ";" {
			// next triple will be for the same subject
			continue
		} else {
			err = errors.New("Unexpected token parsing predicates (" + token + ")")
		}
	}
	return err
}

func (parser *TurtleParser) parseObjects(predicate *PredicateNode) (string, error) {
	var err error
	var token string
	for parser.tokenizer.CanRead() {
		token, err = parser.tokenizer.GetNextToken()
		if err != nil || token == "." || token == ";" {
			// we are done
			break
		} else if token == "," {
			// the next token will be for the same
			// subject + predicate
			continue
		} else {
			// it's a object, add it to the predicate
			predicate.AddObject(token)
		}
	}
	return token, err
}
