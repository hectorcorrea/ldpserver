package fileio

import "os"
import "io"
import "io/ioutil"
import "path/filepath"

// http://stackoverflow.com/a/18415935/446681
var normalAccess os.FileMode = 0644

func WriteFile(filename, content string) error {
	err := createPathForFilename(filename)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, []byte(content), normalAccess)
}

func CreateFile(filename string) error {
	err := createPathForFilename(filename)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, normalAccess)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

func AppendToFile(filename, text string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(text)
	return err
}

func FileExists(filename string) bool {
	// http://stackoverflow.com/a/12518877/446681
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func ReadFile(filename string) (content string, err error) {
	bytes, err := ioutil.ReadFile(filename)
	if err == nil {
		content = string(bytes)
	}
	return content, err
}

func ReaderToString(reader io.ReadCloser) (string, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(bytes[:]), nil
}

func createPathForFilename(filename string) error {
	path := filepath.Dir(filename)
	if err := os.MkdirAll(path, 0777); err != nil {
		return err
	}
	return nil
}
