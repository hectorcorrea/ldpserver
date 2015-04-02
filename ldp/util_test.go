package ldp

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
