package rdf

import (
	"errors"
	// "log"
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
		triples, err := parser.getNextTriples()
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

func (parser TurtleParser) Triples() []Triple {
	return parser.triples
}

func (parser *TurtleParser) getNextTriples() ([]Triple, error) {
	var err error
	var triples []Triple
	var token string

	for err == nil && parser.tokenizer.CanRead() {
		token, err = parser.tokenizer.GetNextToken()
		if err != nil || token == "" {
			break
		}
		subject := NewSubjectNode(token)
		err = parser.parsePredicates(&subject)
		if err == nil {
			for _, triple := range subject.RenderTriples() {
				triples = append(triples, triple)
			}
		}
	}
	return triples, err
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
