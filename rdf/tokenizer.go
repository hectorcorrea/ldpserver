package rdf

import (
	"errors"
	"fmt"
	// "log"
)

type Tokenizer struct {
	scanner Scanner
}

func NewTokenizer(text string) Tokenizer {
	return Tokenizer{scanner: NewScanner(text)}
}

func (tokenizer *Tokenizer) GetNextToken() (string, error) {
	var err error
	var value string

	tokenizer.AdvanceWhiteSpace()
	tokenizer.AdvanceComments()
	if !tokenizer.scanner.CanRead() {
		return "", nil
	}

	firstChar := tokenizer.scanner.Char()
	switch {
	case firstChar == '.':
		value = "."
	case firstChar == ',':
		value = ","
	case firstChar == ';':
		value = ";"
	case firstChar == '@':
		value, err = tokenizer.parseDirective()
	case firstChar == '<':
		value, err = tokenizer.parseUri()
	case firstChar == '"':
		value, err = tokenizer.parseString()
	case tokenizer.isNamespacedChar():
		value = tokenizer.parseNamespacedValue()
	default:
		return "", tokenizer.Error("Invalid first character")
	}

	if err != nil {
		return "", err
	}

	tokenizer.scanner.Advance()
	return value, nil
}

// Advances the index to the beginning of the next triple.
func (tokenizer *Tokenizer) AdvanceTriple() error {
	for tokenizer.CanRead() {
		if tokenizer.scanner.Char() == '.' {
			break
		}
		if tokenizer.isWhiteSpaceChar() {
			tokenizer.scanner.Advance()
			continue
		}
		return tokenizer.Error("Triple did not end with a period.")
	}
	tokenizer.scanner.Advance()
	return nil
}

func (tokenizer *Tokenizer) CanRead() bool {
	return tokenizer.scanner.CanRead()
}

func (tokenizer *Tokenizer) AdvanceWhiteSpace() {
	for tokenizer.CanRead() {
		if !tokenizer.isWhiteSpaceChar() {
			break
		}
		tokenizer.scanner.Advance()
	}
}

func (tokenizer *Tokenizer) AdvanceComments() {
	if !tokenizer.CanRead() || tokenizer.scanner.Char() != '#' {
		return
	}

	for tokenizer.CanRead() {
		if tokenizer.scanner.Char() == '\n' {
			tokenizer.AdvanceWhiteSpace()
			if !tokenizer.CanRead() || tokenizer.scanner.Char() != '#' {
				break
			}
		}
		tokenizer.scanner.Advance()
	}
}

func (tokenizer Tokenizer) isLanguageChar() bool {
	char := tokenizer.scanner.Char()
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char == '-')
}

func (tokenizer Tokenizer) isDirectiveChar() bool {
	char := tokenizer.scanner.Char()
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char == '@')
}

func (tokenizer Tokenizer) isNamespacedChar() bool {
	char := tokenizer.scanner.Char()
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		(char == ':') ||
		(char == '_')
}

func (tokenizer Tokenizer) isUriChar() bool {
	char := tokenizer.scanner.Char()
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		(char == ':') || (char == '/') ||
		(char == '%') || (char == '#') ||
		(char == '+') || (char == '-') ||
		(char == '.') || (char == '_')
}

func (tokenizer Tokenizer) isWhiteSpaceChar() bool {
	char := tokenizer.scanner.Char()
	return char == ' ' || char == '\t' || char == '\n' || char == '\r'
}

// Extracts a value in the form xx:yy or xx
func (tokenizer *Tokenizer) parseNamespacedValue() string {
	start := tokenizer.scanner.Index()
	tokenizer.scanner.Advance()
	for tokenizer.CanRead() {
		if tokenizer.isNamespacedChar() {
			tokenizer.scanner.Advance()
			continue
		} else {
			break
		}
	}
	return tokenizer.scanner.SubstringFrom(start)
}

func (tokenizer *Tokenizer) parseLanguage() string {
	start := tokenizer.scanner.Index()
	tokenizer.scanner.Advance()
	for tokenizer.CanRead() {
		if tokenizer.isLanguageChar() {
			tokenizer.scanner.Advance()
		} else {
			break
		}
	}
	// Should be indicate error if the language is empty?
	return tokenizer.scanner.SubstringFrom(start)
}

// Extracts a value in the form @hello
func (tokenizer *Tokenizer) parseDirective() (string, error) {
	start := tokenizer.scanner.Index()
	tokenizer.scanner.Advance()
	for tokenizer.CanRead() {
		if tokenizer.isDirectiveChar() {
			tokenizer.scanner.Advance()
		} else {
			break
		}
	}

	directive := tokenizer.scanner.SubstringFrom(start)
	if directive == "" {
		return "", tokenizer.Error("Empty directive detected")
	}

	return directive, nil
}

// Extracts a value in quotes, for example
//		"hello"
//      "hello \"world\""
// 		"hello"@en-us
//		"hello"^^<http://somedomain>
func (tokenizer *Tokenizer) parseString() (string, error) {
	start := tokenizer.scanner.Index()
	lastChar := tokenizer.scanner.Char()
	tokenizer.scanner.Advance()
	for tokenizer.CanRead() {
		if tokenizer.scanner.Char() == '"' {
			if lastChar == '\\' {
				lastChar = tokenizer.scanner.Char()
				tokenizer.scanner.Advance()
				continue
			}
			str := tokenizer.scanner.Substring(start, tokenizer.scanner.Index()+1)
			lang := ""
			datatype := ""
			canPeek, nextChar := tokenizer.scanner.Peek()
			var err error
			if canPeek {
				switch nextChar {
				case '@':
					tokenizer.scanner.Advance()
					lang = tokenizer.parseLanguage()
					str += lang
				case '^':
					tokenizer.scanner.Advance()
					datatype, err = tokenizer.parseType()
					str += datatype
				}
			}
			return str, err
		}
		lastChar = tokenizer.scanner.Char()
		tokenizer.scanner.Advance()
	}
	return "", tokenizer.Error("String did not end with \"")
}

func (tokenizer *Tokenizer) parseType() (string, error) {
	canPeek, nextChar := tokenizer.scanner.Peek()
	if !canPeek || nextChar != '^' {
		return "", tokenizer.Error("Invalid type delimiter")
	}

	tokenizer.scanner.Advance()
	canPeek, nextChar = tokenizer.scanner.Peek()
	if !canPeek || nextChar != '<' {
		return "", tokenizer.Error("Invalid URI in type delimiter")
	}

	tokenizer.scanner.Advance()
	uri, err := tokenizer.parseUri()
	return "^^" + uri, err
}

// Extracts an URI in the form <hello>
func (tokenizer *Tokenizer) parseUri() (string, error) {
	start := tokenizer.scanner.Index()
	tokenizer.scanner.Advance()
	for tokenizer.CanRead() {
		if tokenizer.scanner.Char() == '>' {
			uri := tokenizer.scanner.Substring(start, tokenizer.scanner.Index()+1)
			return uri, nil
		}
		if !tokenizer.isUriChar() {
			return "", tokenizer.Error("Invalid character in URI")
		}
		tokenizer.scanner.Advance()
	}
	return "", tokenizer.Error("URI did not end with >")
}

func (tokenizer *Tokenizer) Error(message string) error {
	lastChar := ""
	if tokenizer.CanRead() {
		lastChar = tokenizer.scanner.CharString()
	}
	errorMsg := fmt.Sprintf("%s. Character (%s) at %s.", message, lastChar, tokenizer.scanner.Position())
	return errors.New(errorMsg)
}
