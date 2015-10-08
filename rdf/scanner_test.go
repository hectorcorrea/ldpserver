package rdf

import "testing"

func TestPeek(t *testing.T) {
	scanner := NewScanner("abc")
	if _, nextChar := scanner.Peek(); nextChar != 'b' {
		t.Errorf("Error on first peek")
	}
	scanner.Advance()
	if _, nextChar := scanner.Peek(); nextChar != 'c' {
		t.Errorf("Error on second peek")
	}
	scanner.Advance()
	if canPeek, _ := scanner.Peek(); canPeek == true {
		t.Errorf("Failed to detect that it cannot peek anymore")
	}
}
