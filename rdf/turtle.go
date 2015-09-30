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
		// parser.advanceIndex()
		// if parser.atEnd() {
		// 	break
		// }
	}
	return parser.err
}

func (parser TurtleParser) Triples() []Triple {
	return parser.triples
}

func (parser *TurtleParser) advanceIndex() {
	if !parser.atEnd() {
		parser.index++
	}
}

func (parser TurtleParser) atLastChar() bool {
	return parser.index == len(parser.chars)-1
}

func (parser TurtleParser) atEnd() bool {
	return parser.index > len(parser.chars)-1
}

func (parser TurtleParser) GetNextTriple() (Triple, error) {
	var subject, predicate, object Token
	var index int
	subject, index, _ = GetTokenFromRune(parser.chars, parser.index)
	parser.index = index
	parser.advanceIndex()

	predicate, index, _ = GetTokenFromRune(parser.chars, parser.index)
	parser.index = index
	parser.advanceIndex()

	object, parser.index, _ = GetTokenFromRune(parser.chars, parser.index)
	parser.index = index
	parser.advanceIndex()

	index, err := tripleEndsOK(parser.chars, parser.index)
	parser.index = index
	parser.advanceIndex()
	if err != nil {
		return Triple{}, err
	}

	if object.isLiteral {
		return NewTripleLit(subject.value, predicate.value, object.value), nil
	}

	return NewTripleUri(subject.value, predicate.value, object.value), nil
}

func GetToken(text string) (Token, error) {
	chars := stringToRunes(text)
	token, _, err := GetTokenFromRune(chars, 0)
	return token, err
}

// Gets the token from a string/run
// Returns
//    string the token
//    int the position where the string ends in relation to the original string
//    error (if any)
func GetTokenFromRune(chars []rune, index int) (Token, int, error) {
	start := parseWhiteSpace(chars, index)
	if start >= len(chars) {
		return Token{}, -1, errors.New("End of line reached. No token found after parsing white space.")
	}
	firstChar := chars[start]
	var end int
	var err error
	var isLiteral, isUri, isNamespaced bool
	switch {
	case firstChar == '<':
		isUri = true
		end, err = parseUri(chars, start+1)
	case firstChar == '"':
		isLiteral = true
		end, err = parseString(chars, start+1)
	case isNamespacedChar(firstChar):
		isNamespaced = true
		end, err = parseNamespacedValue(chars, start+1)
	default:
		return Token{}, -1, errors.New("Invalid first character")
	}
	if err != nil {
		return Token{}, -1, err
	}
	value := string(chars[start : end+1])
	token := Token{value: value, isUri: isUri, isLiteral: isLiteral, isNamespaced: isNamespaced}
	return token, end, nil
}

func tripleEndsOK(chars []rune, index int) (int, error) {
	var i int
	for i = index; i < len(chars) && isWhiteSpaceChar(chars[i]); i++ {
	}

	if i == len(chars) || chars[i] != '.' {
		return -1, errors.New("Triple did not end with a period.")
	}
	return i, nil
}

func parseNamespacedValue(chars []rune, index int) (int, error) {
	var i int
	for i = index; i < len(chars) && isNamespacedChar(chars[i]); i++ {
	}
	return i - 1, nil
}

func parseString(chars []rune, index int) (int, error) {
	foundDelimiter := false
	var i int
	for i = index; i < len(chars); i++ {
		if chars[i] == '"' {
			foundDelimiter = true
			break
		}
	}
	if !foundDelimiter {
		return -1, errors.New("String did not end with \"")
	}
	return i, nil
}

func parseUri(chars []rune, index int) (int, error) {
	foundDelimiter := false
	var i int
	for i = index; i < len(chars); i++ {
		if chars[i] == '>' {
			foundDelimiter = true
			break
		}
		if !isUriChar(chars[i]) {
			return -1, errors.New("Invalid character in URI")
		}
	}
	if !foundDelimiter {
		return -1, errors.New("URI did not end with >")
	}
	return i, nil
}

func parseWhiteSpace(chars []rune, index int) int {
	var i int
	for i = index; i < len(chars) && isWhiteSpaceChar(chars[i]); i++ {
	}
	return i
}

func isWhiteSpaceChar(char rune) bool {
	if char == ' ' || char == '\t' ||
		char == '\n' || char == '\r' {
		return true
	}
	return false
}

func isUriChar(char rune) bool {
	if (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		(char == ':') || (char == '/') ||
		(char == '%') || (char == '#') ||
		(char == '+') {
		return true
	}
	return false
}

func isNamespacedChar(char rune) bool {
	if (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		(char == ':') {
		return true
	}
	return false
}

func stringToRunes(text string) []rune {
	var chars []rune
	for _, c := range text {
		chars = append(chars, c)
	}
	return chars
}
