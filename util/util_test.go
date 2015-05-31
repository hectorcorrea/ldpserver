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

func TestIsValidSlug(t *testing.T) {
	validTests := []string{"123", "abc", "123abc", "abc123", "abc_123", "abc-123", "-", "_", "hello.jpg"}
	for _, test := range validTests {
		if !IsValidSlug(test) {
			t.Errorf("IsValidSlug failed: %s", test)
		}
	}

	invalidTests := []string{"/123", "a/bc", "abc..xyz", "a?", "a:", ".", "..", "a\\bc", `a"bc`}
	for _, test := range invalidTests {
		if IsValidSlug(test) {
			t.Errorf("IsValidSlug failed: %s", test)
		}
	}
}

func TestPathFromUri(t *testing.T) {
	rootUri := "http://somewhere.com/"

	testA := "http://somewhere.com/hello/world"
	resultA := PathFromUri(rootUri, testA)
	if resultA != "hello/world" {
		t.Errorf("PathFromUri failed for: %s, %s", testA, resultA)
	}

	testB := "http://different.com/hello/world"
	resultB := PathFromUri(rootUri, testB)
	if resultB != testB {
		t.Errorf("PathFromUri failed for: %s, %s", testB, resultB)
	}
}
