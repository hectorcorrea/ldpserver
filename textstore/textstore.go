package textstore

import (
	"errors"
	"io"
	"ldpserver/fileio"
	"ldpserver/util"
	"os"
)

var AlreadyExistsError = errors.New("Already exists")

const metaFile string = "meta.rdf"
const dataFile string = "data.bin"

type Store struct {
	folder string
	err    error
}

func NewStore(folder string) Store {
	return Store{folder: folder}
}

func CreateStore(folder string) Store {
	store := NewStore(folder)
	if store.Exists() {
		store.err = AlreadyExistsError
	} else {
		store.err = store.SaveMetaFile("")
	}
	return store
}

func (store Store) Exists() bool {
	return storeExists(store.folder)
}

func (store Store) Path() string {
	return store.folder
}

func storeExists(folder string) bool {
	metaRdf := util.PathConcat(folder, metaFile)
	return fileio.FileExists(metaRdf)
}

func (store Store) Error() error {
	return store.err
}

func (store Store) Delete() error {
	// delete the metafile
	metaFileFullPath := util.PathConcat(store.folder, metaFile)
	err := os.Remove(metaFileFullPath)
	if err != nil {
		return err
	}

	// delete the data file
	dataFileFullPath := util.PathConcat(store.folder, dataFile)
	if fileio.FileExists(dataFileFullPath) {
		err = os.Remove(dataFileFullPath)
		if err != nil {
			return err
		}
	}

	// delete the store folder
	return os.Remove(store.folder)
}

func (store Store) SaveMetaFile(content string) error {
	fullFilename := util.PathConcat(store.folder, metaFile)
	return fileio.WriteFile(fullFilename, content)
}

func (store Store) AppendToMetaFile(content string) error {
	fullFilename := util.PathConcat(store.folder, metaFile)
	return fileio.AppendToFile(fullFilename, content)
}

func (store Store) SaveDataFile(reader io.ReadCloser) error {
	fullFilename := util.PathConcat(store.folder, dataFile)
	out, err := os.Create(fullFilename)
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(out, reader)
	return out.Close()
}

// Should this return a reader?
func (store Store) ReadMetaFile() (string, error) {
	fullFilename := util.PathConcat(store.folder, metaFile)
	return fileio.ReadFile(fullFilename)
}

// Should this return a reader?
func (store Store) ReadDataFile() (string, error) {
	fullFilename := util.PathConcat(store.folder, dataFile)
	return fileio.ReadFile(fullFilename)
}
