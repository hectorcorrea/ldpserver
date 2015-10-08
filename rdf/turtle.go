package rdf

import (
	"errors"
	// "log"
)

type TurtleParser struct {
	scanner Scanner
	triples []Triple
}

func NewTurtleParser(text string) TurtleParser {
	scanner := NewScanner(text)
	parser := TurtleParser{scanner: scanner}
	return parser
}

func (parser *TurtleParser) Parse() error {
	var err error
	for parser.scanner.CanRead() {
		triple, err := parser.GetNextTriple()
		if err != nil {
			break
		}
		parser.triples = append(parser.triples, triple)
		parser.advanceWhiteSpace()
	}
	return err
}

func (parser *TurtleParser) ParseOne() (Triple, error) {
	if parser.scanner.CanRead() {
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

	subject, err = parser.GetNextToken()
	if err == nil {
		predicate, err = parser.GetNextToken()
		if err == nil {
			object, err = parser.GetNextToken()
			if err == nil {
				err = parser.AdvanceTriple()
				if err == nil {
					triple = NewTriple(subject, predicate, object)
				}
			}
		}
	}
	return triple, err
}

func (parser *TurtleParser) GetNextToken() (string, error) {
	var err error
	var value string

	parser.advanceWhiteSpace()
	parser.advanceComments()
	if !parser.scanner.CanRead() {
		return "", errors.New("No token found")
	}

	firstChar := parser.scanner.Char()
	switch {
	case firstChar == '<':
		value, err = parser.parseUri()
	case firstChar == '"':
		value, err = parser.parseString()
	case parser.isNamespacedChar():
		value = parser.parseNamespacedValue()
	default:
		return "", errors.New("Invalid first character: [" + parser.scanner.CharString() + "]")
	}

	if err != nil {
		return "", err
	}

	parser.scanner.Advance()
	return value, nil
}

// Advances the index to the beginning of the next triple.
func (parser *TurtleParser) AdvanceTriple() error {
	for parser.scanner.CanRead() {
		if parser.scanner.Char() == '.' {
			break
		}
		if parser.isWhiteSpaceChar() {
			parser.scanner.Advance()
			continue
		}
		return errors.New("Triple did not end with a period.")
	}
	parser.scanner.Advance()
	return nil
}

func (parser *TurtleParser) advanceWhiteSpace() {
	for parser.scanner.CanRead() {
		if !parser.isWhiteSpaceChar() {
			break
		}
		parser.scanner.Advance()
	}
}

func (parser *TurtleParser) advanceComments() {
	if !parser.scanner.CanRead() || parser.scanner.Char() != '#' {
		return
	}

	for parser.scanner.CanRead() {
		if parser.scanner.Char() == '\n' {
			parser.advanceWhiteSpace()
			if parser.scanner.Char() != '#' {
				break
			}
		}
		parser.scanner.Advance()
	}
}

func (parser TurtleParser) isLanguageChar() bool {
	char := parser.scanner.Char()
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char == '-')
}

func (parser TurtleParser) isNamespacedChar() bool {
	char := parser.scanner.Char()
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		(char == ':') ||
		(char == '_')
}

func (parser TurtleParser) isUriChar() bool {
	char := parser.scanner.Char()
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		(char == ':') || (char == '/') ||
		(char == '%') || (char == '#') ||
		(char == '+') || (char == '-') ||
		(char == '.')
}

func (parser TurtleParser) isWhiteSpaceChar() bool {
	char := parser.scanner.Char()
	return char == ' ' || char == '\t' || char == '\n' || char == '\r'
}

// Extracts a value in the form xx:yy or xx
func (parser *TurtleParser) parseNamespacedValue() string {
	start := parser.scanner.Index()
	parser.scanner.Advance()
	for parser.scanner.CanRead() {
		if parser.isNamespacedChar() {
			parser.scanner.Advance()
			continue
		} else {
			break
		}
	}
	return parser.scanner.SubstringFrom(start)
}

func (parser *TurtleParser) parseLanguage() string {
	start := parser.scanner.Index()
	parser.scanner.Advance()
	for parser.scanner.CanRead() {
		if parser.isLanguageChar() {
			parser.scanner.Advance()
		} else {
			break
		}
	}
	// Should be indicate error if the language is empty?
	return parser.scanner.SubstringFrom(start)
}

// Extracts a value in quotes, for example
//		"hello"
// 		"hello"@en-us
//		"hello"^^<http://somedomain>
func (parser *TurtleParser) parseString() (string, error) {
	start := parser.scanner.Index()
	parser.scanner.Advance()
	for parser.scanner.CanRead() {
		if parser.scanner.Char() == '"' {
			str := parser.scanner.Substring(start, parser.scanner.Index()+1)
			lang := ""
			datatype := ""
			canPeek, nextChar := parser.scanner.Peek()
			var err error
			if canPeek {
				switch nextChar {
				case '@':
					parser.scanner.Advance()
					lang = parser.parseLanguage()
					str += lang
				case '^':
					parser.scanner.Advance()
					datatype, err = parser.parseType()
					str += datatype
				}
			}
			return str, err
		}
		parser.scanner.Advance()
	}
	return "", errors.New("String did not end with \"")
}

func (parser *TurtleParser) parseType() (string, error) {
	canPeek, nextChar := parser.scanner.Peek()
	if !canPeek || nextChar != '^' {
		return "", errors.New("Invalid type delimiter")
	}

	parser.scanner.Advance()
	canPeek, nextChar = parser.scanner.Peek()
	if !canPeek || nextChar != '<' {
		return "", errors.New("Invalid URI in type delimiter")
	}

	parser.scanner.Advance()
	uri, err := parser.parseUri()
	return "^^" + uri, err
}

// Extracts an URI in the form <hello>
func (parser *TurtleParser) parseUri() (string, error) {
	start := parser.scanner.Index()
	parser.scanner.Advance()
	for parser.scanner.CanRead() {
		if parser.scanner.Char() == '>' {
			uri := parser.scanner.Substring(start, parser.scanner.Index()+1)
			return uri, nil
		}
		if !parser.isUriChar() {
			return "", errors.New("Invalid character in URI " + parser.scanner.CharString())
		}
		parser.scanner.Advance()
	}
	return "", errors.New("URI did not end with >")
}
