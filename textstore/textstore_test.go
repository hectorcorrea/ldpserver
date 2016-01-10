package textstore

import (
	"ldpserver/util"
	"os"
	"path/filepath"
	"testing"
)

var dataPath string

func init() {
	dataPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
}

func TestTextStore(t *testing.T) {
	store := NewStore(dataPath)
	if store.Exists() {
		t.Errorf("Found an unexpected text store at %s", dataPath)
	}

	store = CreateStore(dataPath)
	if !store.Exists() {
		t.Errorf("Error creating text store at %s", dataPath)
	}

	reader := util.FakeReaderCloser{Text: "hello"}
	if err := store.SaveDataFile(reader); err != nil {
		t.Errorf("Error %s saving text to data file at %s", err, dataPath)
	}

	text, err := store.ReadDataFile()
	if err != nil {
		t.Errorf("Error %s reading text from data file at %s", err, dataPath)
	}

	if text != "hello" {
		t.Errorf("Unexpected text %s found when reading store at %s", err, dataPath)
	}

	store = CreateStore(dataPath)
	if store.Error() == nil {
		t.Errorf("Failed to detect override on create")
	}
}
