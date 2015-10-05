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
	length  int
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
	parser.length = len(parser.chars)
	return parser
}

func (parser *TurtleParser) Parse() error {
	parser.err = nil
	parser.index = 0
	for parser.canRead() {
		triple, err := parser.GetNextTriple()
		if err != nil {
			parser.err = err
			break
		}
		parser.triples = append(parser.triples, triple)
		parser.advanceWhiteSpace()
	}
	return parser.err
}

func (parser *TurtleParser) ParseOne() (Triple, error) {
	parser.err = nil
	parser.index = 0
	if parser.canRead() {
		return parser.GetNextTriple()
	}
	return Triple{}, errors.New("No triple found.")
}

func (parser TurtleParser) Triples() []Triple {
	return parser.triples
}

func (parser *TurtleParser) GetNextTriple() (Triple, error) {
	var subject, predicate, object Token
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
					triple = NewTripleFromTokens(subject, predicate, object)
				}
			}
		}
	}
	return triple, err
}

func (parser *TurtleParser) GetNextToken() (Token, error) {
	var err error
	var isLiteral, isUri, isNamespaced bool
	var value string

	parser.advanceWhiteSpace()
	parser.advanceComments()
	if !parser.canRead() {
		return Token{}, errors.New("No token found")
	}

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
		return Token{}, errors.New("Invalid first character: [" + parser.charString() + "]")
	}

	if err != nil {
		return Token{}, err
	}

	parser.advance()
	token := Token{value: value, isUri: isUri, isLiteral: isLiteral, isNamespaced: isNamespaced}
	return token, nil
}

// Advances the index to the beginning of the next triple.
func (parser *TurtleParser) AdvanceTriple() error {
	for parser.canRead() {
		if parser.char() == '.' {
			break
		}
		if parser.isWhiteSpaceChar() {
			parser.advance()
			continue
		}
		return errors.New("Triple did not end with a period.")
	}
	parser.advance()
	return nil
}

// Advances the index to the next character.
func (parser *TurtleParser) advance() {
	if parser.canRead() {
		parser.index++
	}
}

func (parser *TurtleParser) advanceWhiteSpace() {
	for parser.canRead() {
		if !parser.isWhiteSpaceChar() {
			break
		}
		parser.advance()
	}
}

func (parser *TurtleParser) advanceComments() {
	if !parser.canRead() || parser.char() != '#' {
		return
	}

	for parser.canRead() {
		if parser.char() == '\n' {
			parser.advanceWhiteSpace()
			if parser.char() != '#' {
				break
			}
		}
		parser.advance()
	}
}

func (parser *TurtleParser) canRead() bool {
	if parser.length == 0 {
		return false
	}
	return parser.index < parser.length
}

func (parser TurtleParser) char() rune {
	return parser.chars[parser.index]
}

func (parser TurtleParser) charString() string {
	return string(parser.chars[parser.index])
}

func (parser TurtleParser) isLanguageChar() bool {
	char := parser.char()
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char == '-')
}

func (parser TurtleParser) isNamespacedChar() bool {
	char := parser.char()
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		(char == ':') ||
		(char == '_')
}

func (parser TurtleParser) isUriChar() bool {
	char := parser.char()
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		(char == ':') || (char == '/') ||
		(char == '%') || (char == '#') ||
		(char == '+') || (char == '-') ||
		(char == '.')
}

func (parser TurtleParser) isWhiteSpaceChar() bool {
	char := parser.char()
	return char == ' ' || char == '\t' || char == '\n' || char == '\r'
}

// Extracts a value in the form xx:yy or xx
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

func (parser *TurtleParser) parseLanguage() (string, error) {
	start := parser.index
	parser.advance()
	for parser.canRead() {
		if parser.isLanguageChar() {
			parser.advance()
		} else {
			break
		}
	}
	return string(parser.chars[start:parser.index]), nil
}

// Extracts a value in quotes, e.g. "hello"
func (parser *TurtleParser) parseString() (string, error) {
	start := parser.index
	parser.advance()
	for parser.canRead() {
		if parser.char() == '"' {
			str := string(parser.chars[start : parser.index+1])
			lang := ""
			// datatype := ""
			var err error
			canPeek, nextChar := parser.peek()
			if canPeek {
				switch nextChar {
				case '@':
					parser.advance()
					lang, err = parser.parseLanguage()
					str += lang
					// case: "^"
					// 	datatype = parser.parseType()
				}
			}
			return str, err
		}
		parser.advance()
	}
	return "", errors.New("String did not end with \"")
}

// Extracts an URI in the form <hello>
func (parser *TurtleParser) parseUri() (string, error) {
	start := parser.index
	parser.advance()
	for parser.canRead() {
		if parser.char() == '>' {
			uri := string(parser.chars[start : parser.index+1])
			return uri, nil
		}
		if !parser.isUriChar() {
			return "", errors.New("Invalid character in URI " + parser.charString())
		}
		parser.advance()
	}
	return "", errors.New("URI did not end with >")
}

func (parser *TurtleParser) peek() (bool, rune) {
	if parser.length > 0 && parser.index < (parser.length-1) {
		return true, parser.chars[parser.index+1]
	}
	return false, 0
}

func stringToRunes(text string) []rune {
	var chars []rune
	for _, c := range text {
		chars = append(chars, c)
	}
	return chars
}
