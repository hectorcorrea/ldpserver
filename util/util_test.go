package util

import "testing"

func TestPathConcat(t *testing.T) {
	if test := PathConcat("a", "b"); test != "a/b" {
		t.Errorf("PathConcat failed: %s", test)
	}

	if test := PathConcat("a/", "b"); test != "a/b" {
		t.Errorf("PathConcat failed: %s", test)
	}

	if test := PathConcat("a", "/b"); test != "a/b" {
		t.Errorf("PathConcat failed: %s", test)
	}

	if test := PathConcat("a/", "/b"); test != "a/b" {
		t.Errorf("PathConcat failed: %s", test)
	}

	if test := PathConcat("a/", "/"); test != "a/" {
		t.Errorf("PathConcat failed: %s", test)
	}

	if test := PathConcat("/", "/b"); test != "/b" {
		t.Errorf("PathConcat failed: %s", test)
	}

	if test := PathConcat("/", "/"); test != "/" {
		t.Errorf("PathConcat failed: %s", test)
	}
}

func TestUriConcat(t *testing.T) {
	if test := UriConcat("localhost", "/"); test != "localhost" {
		t.Errorf("UriConcat failed: %s", test)
	}

	if test := UriConcat("localhost", ""); test != "localhost" {
		t.Errorf("UriConcat failed: %s", test)
	}

	if test := UriConcat("localhost/", ""); test != "localhost" {
		t.Errorf("UriConcat failed: %s", test)
	}

	if test := UriConcat("localhost/", "/"); test != "localhost" {
		t.Errorf("UriConcat failed: %s", test)
	}
}

func TestStripSlash(t *testing.T) {
	if test := StripSlash("abc/"); test != "abc" {
		t.Errorf("PathConcat failed: %s", test)
	}

	if test := StripSlash("abc"); test != "abc" {
		t.Errorf("PathConcat failed: %s", test)
	}

	if test := StripSlash("abc//"); test != "abc/" {
		t.Errorf("PathConcat failed: %s", test)
	}

	if test := StripSlash("/abc/"); test != "/abc" {
		t.Errorf("PathConcat failed: %s", test)
	}

	if test := StripSlash("/"); test != "" {
		t.Errorf("PathConcat failed: %s", test)
	}
}

func TestIsAlphaNumeric(t *testing.T) {
	validTests := []string{"123", "abc", "123abc", "abc123", "abc_123", "abc-123", "-", "-"}
	for _, test := range validTests {
		if !IsAlphaNumeric(test) {
			t.Errorf("IsAlphaNumeric failed: %s", test)
		}
	}

	invalidTests := []string{"/123", "a/bc", "abc.xyz", "a?", "a:"}
	for _, test := range invalidTests {
		if IsAlphaNumeric(test) {
			t.Errorf("IsAlphaNumeric failed: %s", test)
		}
	}
}
