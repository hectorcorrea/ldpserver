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
	var err error
	for parser.tokenizer.CanRead() {
		triple, err := parser.GetNextTriple()
		if err != nil {
			break
		}
		parser.triples = append(parser.triples, triple)
		parser.tokenizer.AdvanceWhiteSpace()
	}
	return err
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
