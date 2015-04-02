package fileio

import "testing"

type testpair struct {
	test     string
	expected string
}

func TestPathFromFilename(t *testing.T) {
	var tests = []testpair{
		{"/hello/world", "/hello"},
		{"/hello/", "/hello"},
		{"/", ""},
	}

	for _, pair := range tests {
		result, _ := PathFromFilename(pair.test)
		if result != pair.expected {
			t.Errorf("failed for %s, expected %s, got %s", pair.test, pair.expected, result)
		}
	}
}

func TestPathFromFilenameErrors(t *testing.T) {
	var tests = []string{
		"",
		"hello",
	}

	for _, test := range tests {
		_, err := PathFromFilename(test)
		if err == nil {
			t.Errorf("expected error for %s", test)
		}
	}
}
