package rdf

import "strings"
import "log"

type Triple struct {
	subject   			string 			// always a URI
	predicate 			string 			// always a URI
	object    			string      // can be a URI or a literal
	isObjectLiteral bool
}

var specialChars string = `"\<>`

func NewTripleUri(subject, predicate, object string) Triple {
	return newTriple(subject, predicate, object, false)
}

func NewTripleLit(subject, predicate, object string) Triple {
	return newTriple(subject, predicate, object, true)
}

func newTripleFromNTriple(ntriple NTriple) Triple {
	return newTriple(ntriple.Subject(), ntriple.Predicate(), ntriple.Object(), ntriple.IsObjectLiteral()) 
}

func newTriple(subject, predicate, object string, isObjectLiteral bool) Triple {
	return Triple{subject: subject, predicate: predicate, object: object, isObjectLiteral: isObjectLiteral}
}

func (t Triple) String() string {
	str := "<" + encode(t.subject) + "> <" + encode(t.predicate) + "> "
	if t.isObjectLiteral {
		str += `"` + encode(t.object) + `" .`
	} else {
		str += "<" + encode(t.object) + "> ."
	}
	return str
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
	if len(blank) > 0 {
		line = strings.Replace(line, "<>", "<"+blank+">", -1)
	}
	ntriple, err := NewNTripleFromString(line)
	if err != nil {
		log.Printf("Error parsing %s. Error: %s", line, err)
		return Triple{}, err
	}
	return newTripleFromNTriple(ntriple), nil
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
