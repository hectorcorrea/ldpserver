package fileio

import "fmt"
import "os"
import "io"
import "io/ioutil"
import "strings"
import "errors"

func PathFromFilename(filename string) (string, error) {
	lastSlash := strings.LastIndex(filename, "/")
	if lastSlash < 0 {
		errorMsg := fmt.Sprintf("Cannot determine path from filename (%s)", filename)
		return "", errors.New(errorMsg)
	}
	return filename[0:lastSlash], nil
}

func WriteFile(filename, content string) error {
	path, err := PathFromFilename(filename)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(path, 0777); err != nil {
		return err
	}

	var access os.FileMode = 0644
	return ioutil.WriteFile(filename, []byte(content), access)
}

func AppendToFile(filename, text string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err2 := file.WriteString(text)
	return err2
}

func FileExists(filename string) bool {
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
