package rdf

import "strings"
import "regexp"
import "fmt"
import "errors"

type Triple struct {
	subject   string
	predicate string
	object    string
}

var specialChars string = `"\<>`

func NewTriple(subject, predicate, object string) Triple {
	return Triple{subject: subject, predicate: predicate, object: object}
}

func (t Triple) String() string {
	return "<" + encode(t.subject) + "> " +
		"<" + encode(t.predicate) + "> " +
		"<" + encode(t.object) + "> ."
}

func (t Triple) StringLn() string {
	return t.String() + "\n"
}

func stripDelimiters(text string) string {
	return text[1 : len(text)-1]
}

// Creates a triple from a string in the following format
//    <subject> <predicate> <object> .
func StringToTriple(line, blank string) (Triple, error) {
	var triple Triple
	if len(blank) > 0 {
		line = strings.Replace(line, "<>", "<"+blank+">", -1)
	}
	re := regexp.MustCompile("<([^>]+)>")
	matches := re.FindAllString(line, -1)
	if len(matches) == 3 {
		triple.subject = stripDelimiters(matches[0])
		triple.predicate = stripDelimiters(matches[1])
		triple.object = stripDelimiters(matches[2])
		return triple, nil
	}
	errorMsg := fmt.Sprintf("%d elements found in triple %s", len(matches), line)
	return triple, errors.New(errorMsg)
}

func encode(value string) string {
	if strings.IndexAny(value, specialChars) == -1 {
		return value
	}
	return doEncode(value)
}

func doEncode(value string) string {
	// This is a horrible way of doing the encoding,
	// but it would do for now.
	encodings := make(map[string]string)
	encodings[`"`] = `\"`
	encodings[`\`] = `\\`
	encodings["<"] = "\\<"
	encodings[">"] = "\\>"
	encoded := ""
	for _, r := range value {
		char := string(r)
		if replacement, ok := encodings[char]; ok {
			encoded += replacement
		} else {
			encoded += char
		}
	}
	return encoded
}
