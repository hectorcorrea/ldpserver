package rdf

import (
	"errors"
	// "log"
)

type Token struct {
	value        string
	isUri        bool
	isLiteral    bool
	isNamespaced bool
}

type TurtleParser struct {
	index   int
	text    string
	chars   []rune
	triples []Triple
	err     error
}

func NewTurtleParser(text string) TurtleParser {
	// Convert the original string to an array of unicode runes.
	// This allows us to iterate on it as if it was an array
	// of ASCII chars even if there are Unicode characters on it
	// that use 2-4 bytes.
	chars := stringToRunes(text)
	parser := TurtleParser{text: text, chars: chars}
	return parser
}

func (parser TurtleParser) Parse() error {
	parser.err = nil
	parser.index = 0
	for {
		triple, err := parser.GetNextTriple()
		if err != nil {
			parser.err = err
			break
		}

		parser.triples = append(parser.triples, triple)
		break
	}
	return parser.err
}

func (parser TurtleParser) Triples() []Triple {
	return parser.triples
}

func (parser *TurtleParser) GetNextTriple() (Triple, error) {
	subject, _ := parser.GetNextToken()
	predicate, _ := parser.GetNextToken()
	object, _ := parser.GetNextToken()

	err := parser.AdvanceTriple()
	if err != nil {
		return Triple{}, err
	}

	if object.isLiteral {
		return NewTripleLit(subject.value, predicate.value, object.value), nil
	}

	return NewTripleUri(subject.value, predicate.value, object.value), nil
}

func (parser *TurtleParser) GetNextToken() (Token, error) {
	var err error
	var isLiteral, isUri, isNamespaced bool
	var value string

	parser.advanceWhiteSpace()
	firstChar := parser.char()

	switch {
	case firstChar == '<':
		isUri = true
		value, err = parser.parseUri()
	case firstChar == '"':
		isLiteral = true
		value, err = parser.parseString()
	case parser.isNamespacedChar():
		isNamespaced = true
		value = parser.parseNamespacedValue()
	default:
		return Token{}, errors.New("Invalid first character")
	}
	if err != nil {
		return Token{}, err
	}

	parser.advance()
	token := Token{value: value, isUri: isUri, isLiteral: isLiteral, isNamespaced: isNamespaced}
	return token, nil
}

func (parser *TurtleParser) AdvanceTriple() error {
	for parser.canRead() {
		if parser.char() == '.' {
			break
		}
		if parser.isWhiteSpaceChar() {
			parser.advance()
			continue
		}
		return errors.New("Triple did not end with a period2.")
	}
	parser.advance()
	return nil
}

func (parser *TurtleParser) advance() {
	if !parser.atEnd() {
		parser.index++
	}
}

func (parser *TurtleParser) advanceWhiteSpace() {
	for parser.canRead() {
		if parser.atLastChar() || !parser.isWhiteSpaceChar() {
			break
		}
		parser.advance()
	}
}

func (parser TurtleParser) atEnd() bool {
	if len(parser.chars) == 0 {
		return true
	}
	return parser.index > len(parser.chars)-1
}

func (parser TurtleParser) atLastChar() bool {
	return parser.index == len(parser.chars)-1
}

func (parser *TurtleParser) canRead() bool {
	return !parser.atEnd()
}

func (parser TurtleParser) char() rune {
	return parser.chars[parser.index]
}

func (parser TurtleParser) isNamespacedChar() bool {
	char := parser.char()
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		(char == ':')
}

func (parser TurtleParser) isUriChar() bool {
	char := parser.char()
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		(char == ':') || (char == '/') ||
		(char == '%') || (char == '#') ||
		(char == '+')
}

func (parser TurtleParser) isWhiteSpaceChar() bool {
	char := parser.char()
	return char == ' ' || char == '\t' || char == '\n' || char == '\r'
}

func (parser *TurtleParser) parseNamespacedValue() string {
	start := parser.index
	parser.advance()
	for parser.canRead() {
		if parser.isNamespacedChar() {
			parser.advance()
			continue
		} else {
			break
		}
	}
	return string(parser.chars[start:parser.index])
}

func (parser *TurtleParser) parseString() (string, error) {
	// TODO: Move the advance outside of here.
	// We should already be inside the URI.
	start := parser.index
	parser.advance()
	for parser.canRead() {
		if parser.char() == '"' {
			uri := string(parser.chars[start : parser.index+1])
			return uri, nil
		}
		parser.advance()
	}
	return "", errors.New("String did not end with \"")
}

func (parser *TurtleParser) parseUri() (string, error) {
	// TODO: Move the advance outside of here.
	// We should already be inside the URI.
	start := parser.index
	parser.advance()
	for parser.canRead() {
		if parser.char() == '>' {
			uri := string(parser.chars[start : parser.index+1])
			return uri, nil
		}
		if !parser.isUriChar() {
			return "", errors.New("Invalid character in URI")
		}
		parser.advance()
	}
	return "", errors.New("URI did not end with >")
}

func stringToRunes(text string) []rune {
	var chars []rune
	for _, c := range text {
		chars = append(chars, c)
	}
	return chars
}
