package textstore

import (
	"io"
	"ldpserver/fileio"
	"ldpserver/util"
	"os"
)

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
	if !store.Exists() {
		store.err = store.SaveFile(metaFile, "")
	}
	return store
}

func Exists(folder string) bool {
	return StoreExists(folder)
}

func (store Store) Exists() bool {
	return StoreExists(store.folder)
}

func (store Store) Path() string {
	return store.folder
}

func StoreExists(folder string) bool {
	metaRdf := util.PathConcat(folder, metaFile)
	return fileio.FileExists(metaRdf)
}

func (store Store) Error() error {
	return store.err
}

func (store Store) SaveFile(filename string, content string) error {
	fullFilename := util.PathConcat(store.folder, filename)
	return fileio.WriteFile(fullFilename, content)
}

func (store Store) AppendToFile(filename string, content string) error {
	fullFilename := util.PathConcat(store.folder, filename)
	return fileio.AppendToFile(fullFilename, content)
}

func (store Store) SaveReader(filename string, reader io.ReadCloser) error {
	fullFilename := util.PathConcat(store.folder, filename)
	out, err := os.Create(fullFilename)
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(out, reader)
	return out.Close()
}

func (store Store) ReadFile(filename string) (string, error) {
	fullFilename := util.PathConcat(store.folder, filename)
	return fileio.ReadFile(fullFilename)
}
