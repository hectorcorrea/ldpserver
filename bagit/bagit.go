package bagit

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"ldpserver/fileio"
	"ldpserver/util"
)

type Bag struct {
	folder     string
	dataFolder string
	err        error
}

func NewBag(folder string) Bag {
	if reserved, err := isReservedFolder(folder); reserved == true {
		return Bag{err: err}
	}
	dataFolder := util.PathConcat(folder, "data")
	return Bag{folder: folder, dataFolder: dataFolder}
}

func CreateBag(folder string) Bag {
	if reserved, err := isReservedFolder(folder); reserved == true {
		return Bag{err: err}
	}
	dataFolder := util.PathConcat(folder, "data")
	bag := Bag{folder: folder, dataFolder: dataFolder}
	bag.err = bag.createBagItTxt()
	if bag.err != nil {
		return bag
	}
	bag.err = bag.createManifest()
	return bag
}

func (bag Bag) Exists() bool {
	bagIt := util.PathConcat(bag.folder, "bagit.txt")
	return fileio.FileExists(bagIt)
}

func BagExists(folder string) bool {
	bagIt := util.PathConcat(folder, "bagit.txt")
	return fileio.FileExists(bagIt)
}

func (bag Bag) Error() error {
	return bag.err
}

func (bag Bag) SaveFile(filename string, content string) error {
	fullFilename := util.PathConcat(bag.dataFolder, filename)
	err := fileio.WriteFile(fullFilename, content)
	if err != nil {
		return err
	}
	// TODO: calculate md5 of file
	// TODO: append/update file to manifest
	return nil
}

// this function really shouldn't be on the bagit module
// but if we leave it out, we'll need to expose the file
// path structure (e.g. data/xyz)
func (bag Bag) AppendToFile(filename string, content string) error {
	fullFilename := util.PathConcat(bag.dataFolder, filename)
	err := fileio.AppendToFile(fullFilename, content)
	if err != nil {
		return err
	}
	// TODO: calculate md5 of file
	return nil
}

func (bag Bag) SaveReader(filename string, reader io.ReadCloser) error {
	fullFilename := util.PathConcat(bag.dataFolder, filename)
	out, err := os.Create(fullFilename)
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(out, reader)
	// TODO: calculate md5 of file
	// TODO: append/update file to manifest
	return out.Close()
}

func (bag Bag) ReadFile(filename string) (string, error) {
	fullFilename := util.PathConcat(bag.dataFolder, filename)
	// log.Printf("BagIt.ReadFile %s", fullFilename)
	return fileio.ReadFile(fullFilename)
}

func (bag Bag) createBagItTxt() error {
	// TODO: replace create + write with single write
	bagIt := util.PathConcat(bag.folder, "bagit.txt")
	err := fileio.CreateFile(bagIt)
	if err != nil {
		return err
	}
	bagItText := "BagIt-Version: 0.97\nTag-File-Character-Encoding: UTF-8\n"
	return fileio.WriteFile(bagIt, bagItText)
}

func (bag Bag) createManifest() error {
	manifestFile := util.PathConcat(bag.folder, "manifest-md5.txt")
	return fileio.CreateFile(manifestFile)
	// text := "data/meta.rdf TBD\ndata/meta.rdf TBD\n"
	// return fileio.WriteFile(manifestFile, text)
}

func isReservedFolder(folder string) (bool, error) {
	base := path.Base(folder)
	if base == "data" || base == "bagit.txt" || base == "manifest-md5.txt" {
		errorMsg := fmt.Sprintf("Reserved bag name. Bag cannot be named [%s].", base)
		return true, errors.New(errorMsg)
	}
	return false, nil
}
