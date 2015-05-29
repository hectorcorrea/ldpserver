package util

import "strings"
import "io"
import "regexp"

func PathConcat(path1, path2 string) string {
	if strings.HasSuffix(path1, "/") {
		if strings.HasPrefix(path2, "/") {
			return path1 + path2[1:]
		} else {
			return path1 + path2
		}
	}

	if strings.HasPrefix(path2, "/") {
		return path1 + path2
	}

	return path1 + "/" + path2
}

func UriConcat(path1, path2 string) string {
	return StripSlash(PathConcat(path1, path2))
}

func StripSlash(path string) string {
	if strings.HasSuffix(path, "/") {
		return path[0 : len(path)-1]
	}
	return path
}

func PathFromUri(rootUri, uri string) string {
	if strings.HasPrefix(uri, rootUri) {
		return uri[len(rootUri):]
	}
	return uri
}

func IsAlphaNumeric(str string) bool {
	// Source: https://www.socketloop.com/tutorials/golang-regular-expression-alphanumeric-underscore
	re := regexp.MustCompile("^[a-zA-Z0-9_-]*$")
	return re.MatchString(str)
}

// Used for testing
type FakeReaderCloser struct {
	Text string
}

func (reader FakeReaderCloser) Read(buffer []byte) (int, error) {
	bytes := []byte(reader.Text)
	for i, b := range bytes {
		buffer[i] = b
	}
	return len(bytes), io.EOF
}

func (reader FakeReaderCloser) Close() error {
	return nil
}
