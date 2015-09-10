package textstore

import (
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
    t.Errorf("Found an unexpected text store at %s", store.Path())
  }

  store = CreateStore(dataPath)
  if !store.Exists() {
    t.Errorf("Error creating text store at %s", store.Path())
  }

  if err := store.SaveFile("demo.txt", "hello"); err != nil {
    t.Errorf("Error %s saving text file demo.txt at %s", err, store.Path())
  }

  text, err := store.ReadFile("demo.txt")
  if err != nil {
    t.Errorf("Error %s reading text file demo.txt at %s", err, store.Path())
  }

  if text != "hello" {
    t.Errorf("Unexpected text %s found when reading store at %s", err, store.Path())
  }

}


