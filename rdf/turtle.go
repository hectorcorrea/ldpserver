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
		triples, err := parser.GetNextTriples()
		if err != nil {
			return err
		}
		for _, triple := range triples {
			log.Printf("triple %s", triple)
			parser.triples = append(parser.triples, triple)
		}
		parser.tokenizer.AdvanceWhiteSpace()
	}
	return nil
}

// func (parser *TurtleParser) Parse2() error {
// 	for parser.tokenizer.CanRead() {
// 		triples, err := parser.GetNextTriples()
// 		if err != nil {
// 			return err
// 		}
// 		for _, triple := range triples {
// 			parser.triples = append(parser.triples, triple)
// 		}
// 		parser.tokenizer.AdvanceWhiteSpace()
// 	}
// 	return nil
// }

func (parser *TurtleParser) ParseOne() (Triple, error) {
	var triple Triple
	var err error
	if !parser.tokenizer.CanRead() {
		return Triple{}, errors.New("No triples were found.")
	}

	triples, err := parser.GetNextTriples()
	if err == nil {
		if len(triples) > 0 {
			triple = triples[0]
		} else {
			err = errors.New("No triples were found")
		}
	}
	return triple, err
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
	var err error
	var triples []Triple
	var token string

	for err == nil && parser.tokenizer.CanRead() {
		token, err = parser.tokenizer.GetNextToken()
		if err != nil {
			if err.Error() == "No token found" {
				err = nil
			}
			break
		}
		log.Printf("subject: %s", token)
		subject := NewNode(token)

		token, err = parser.tokenizer.GetNextToken()
		if err != nil {
			break
		}

		log.Printf("predicate: %s", token)
		predicate := subject.AddChild(token)
		parser.parseObjects(predicate)
		for _, triple := range subject.RenderTriples() {
			triples = append(triples, triple)
		}
	}
	return triples, err
}

func (parser *TurtleParser) parseObjects(predicate *Node) error {
	var err error
	var token string
	for parser.tokenizer.CanRead() {
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
			log.Printf("object: %s", token)
			predicate.AddChild(token)
		}
	}
	return err
}
