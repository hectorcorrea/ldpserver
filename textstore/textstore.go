package textstore

import (
	"errors"
	"io"
	"ldpserver/fileio"
	"ldpserver/util"
	"os"
)

var AlreadyExistsError = errors.New("Already exists")
var CreateDeletedError = errors.New("Attempting to create a store that has been previously deleted")

const metaFile string = "meta.rdf"
const dataFile string = "data.bin"
const deletedMarkFile string = "deleted"

type Store struct {
	folder string
	err    error
}

func NewStore(folder string) Store {
	return Store{folder: folder}
}

func CreateStore(folder string) Store {
	store := NewStore(folder)
	switch {
	case store.Exists():
		store.err = AlreadyExistsError
	case store.isDeleted():
		store.err = CreateDeletedError
	default:
		store.err = store.SaveMetaFile("")
	}
	return store
}

func (store Store) Exists() bool {
	return storeExists(store.folder)
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

	return store.markAsDeleted()
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

func (store Store) isDeleted() bool {
	deletedFile := util.PathConcat(store.folder, deletedMarkFile)
	return fileio.FileExists(deletedFile)
}

func (store Store) markAsDeleted() error {
	fullFilename := util.PathConcat(store.folder, deletedMarkFile)
	return fileio.WriteFile(fullFilename, "deleted")
}

func storeExists(folder string) bool {
	// we consider that it exists as long as there is a
	// metadata file on it.
	// Notice that we don't look for the Deleted Mark File
	// when determining if it exists or not.
	metaRdf := util.PathConcat(folder, metaFile)
	return fileio.FileExists(metaRdf)
}
