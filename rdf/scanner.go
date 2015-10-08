package rdf

type Scanner struct {
	index  int
	text   string
	chars  []rune
	length int
}

func NewScanner(text string) Scanner {
	// Convert the original string to an array of unicode runes.
	// This allows us to iterate on it as if it was an array
	// of ASCII chars even if there are Unicode characters on it
	// that use 2-4 bytes.
	chars := stringToRunes(text)
	scanner := Scanner{text: text, chars: chars, length: len(chars)}
	return scanner
}

func (scanner Scanner) Index() int {
	return scanner.index
}

func (scanner Scanner) Substring(start, end int) string {
	return string(scanner.chars[start:end])
}

func (scanner Scanner) SubstringFrom(start int) string {
	return string(scanner.chars[start:scanner.index])
}

// Advances the index to the next character.
func (scanner *Scanner) Advance() {
	if scanner.CanRead() {
		scanner.index++
	}
}

func (scanner *Scanner) CanRead() bool {
	if scanner.length == 0 {
		return false
	}
	return scanner.index < scanner.length
}

func (scanner Scanner) Char() rune {
	return scanner.chars[scanner.index]
}

func (scanner Scanner) CharString() string {
	return string(scanner.chars[scanner.index])
}

func (scanner *Scanner) Peek() (bool, rune) {
	if scanner.length > 0 && scanner.index < (scanner.length-1) {
		return true, scanner.chars[scanner.index+1]
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
