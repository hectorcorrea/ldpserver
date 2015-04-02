package ldp

import "strings"
import "io"

// import "log"

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

func FileNamesForUri(settings Settings, fullUri string) (string, string) {
	path := fullUri[len(settings.rootUrl):]
	return FileNamesForPath(settings, path)
}

func FileNamesForPath(settings Settings, path string) (string, string) {
	pathOnDisk := PathConcat(settings.dataPath, path)
	metaOnDisk := PathConcat(pathOnDisk, "meta.rdf")
	dataOnDisk := PathConcat(pathOnDisk, "data.txt")
	return metaOnDisk, dataOnDisk
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
