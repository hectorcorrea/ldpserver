package rdf

import (
	"errors"
	"log"
)

type TurtleParser struct {
	tokenizer Tokenizer
	triples   []Triple
}

func NewTurtleParser(text string) TurtleParser {
	tokenizer := NewTokenizer(text)
	parser := TurtleParser{tokenizer: tokenizer}
	return parser
}

func (parser *TurtleParser) Parse() error {
	for parser.tokenizer.CanRead() {
		triple, err := parser.GetNextTriple()
		if err != nil {
			return err
		}
		parser.triples = append(parser.triples, triple)
		parser.tokenizer.AdvanceWhiteSpace()
	}
	return nil
}

func (parser *TurtleParser) Parse2() error {
	for parser.tokenizer.CanRead() {
		triples, err := parser.GetNextTriples()
		if err != nil {
			return err
		}
		for _, triple := range triples {
			parser.triples = append(parser.triples, triple)
		}
		parser.tokenizer.AdvanceWhiteSpace()
	}
	return nil
}

func (parser *TurtleParser) ParseOne() (Triple, error) {
	if parser.tokenizer.CanRead() {
		return parser.GetNextTriple()
	}
	return Triple{}, errors.New("No triple found.")
}

func (parser TurtleParser) Triples() []Triple {
	return parser.triples
}

func (parser *TurtleParser) GetNextTriple() (Triple, error) {
	var subject, predicate, object string
	var err error
	var triple Triple

	subject, err = parser.tokenizer.GetNextToken()
	if err == nil {
		predicate, err = parser.tokenizer.GetNextToken()
		if err == nil {
			object, err = parser.tokenizer.GetNextToken()
			if err == nil {
				err = parser.tokenizer.AdvanceTriple()
				if err == nil {
					triple = NewTriple(subject, predicate, object)
				}
			}
		}
	}
	return triple, err
}

func (parser *TurtleParser) GetNextTriples() ([]Triple, error) {
	var subject, predicate string
	var err error
	var triples []Triple
	var tree Tree

	for err == nil && parser.tokenizer.CanRead() {

		subject, err = parser.tokenizer.GetNextToken()
		if err != nil {
			break
		}

		subjectNode := tree.AddNode(subject)

		predicate, err = parser.tokenizer.GetNextToken()
		if err != nil {
			break
		}

		predicateNode := subjectNode.AddChild(predicate)
		parser.parseObjects(predicateNode)
		// for {
		// 	token, err := parser.tokenizer.GetNextToken()
		// 	if err != nil {
		// 		break
		// 	}

		// 	switch {
		// 	case token == ".":
		// 		// we are done
		// 		break
		// 	case token == ",":
		// 		// the next token will be for the same s/p
		// 		// continue
		// 	default:
		// 		object = token
		// 		predicateNode.AddChild(object)
		// 	}
		// }
		triples = subjectNode.RenderTriples()
		log.Printf("Triple: %s", subjectNode.Render())
	}

	return triples, err
}

func (parser *TurtleParser) parseObjects(predicate *Node) error {
	var err error
	var token string
	for {
		token, err = parser.tokenizer.GetNextToken()
		if err != nil || token == "." {
			// we are done
			break
		} else if token == "," {
			// the next token will be for the same
			// subject + predicate
			continue
		} else {
			// it's a object, add it to the predicate
			predicate.AddChild(token)
		}
	}
	return err
}
