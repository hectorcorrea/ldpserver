package rdf

import (
	"errors"
)

func GetTriple(text string) (string, string, string) {
	chars := stringToRunes(text)
	subject, a, _ := GetTokenFromRune(chars, 0)
	predicate, b, _ := GetTokenFromRune(chars, a)
	object, _, _ := GetTokenFromRune(chars, b)
	return subject, predicate, object
}

func GetToken(text string) (string, error) {
	chars := stringToRunes(text)
	token, _, err := GetTokenFromRune(chars, 0)
	return token, err
}

// Gets the token from a string/run
// Returns
//    string the token
//    int the position where the string ends in relation to the original string
//    error (if any)
func GetTokenFromRune(chars []rune, index int) (string, int, error) {
	start := parseWhiteSpace(chars, index)
	if start >= len(chars) {
		return "", -1, errors.New("End of line reached. No token found after parsing white space.")
	}
	firstChar := chars[start]
	var end int
	var err error
	switch {
	case firstChar == '<':
		end, err = parseUri(chars, start+1)
	case firstChar == '"':
		end, err = parseString(chars, start+1)
	case isNamespacedChar(firstChar):
		end, err = parseNamespacedValue(chars, start+1)
	default:
		return "", -1, errors.New("Invalid first character")
	}
	if err != nil {
		return "", -1, err
	}
	token := chars[start : end+1]
	return string(token), end, nil
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
