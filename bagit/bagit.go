package bagit

import (
  "log"
  "ldpserver/fileio"
  "ldpserver/util"
)

type Bag struct {
  folder string
  dataFolder string 
  err error
}

func CreateBag(folder string) Bag {
  log.Printf("BagIt.CreateBag() %s", folder)
  dataFolder := util.PathConcat(folder, "data")
  bag := Bag{folder: folder, dataFolder: dataFolder}
  bag.err = bag.createBagItTxt()
  if bag.err != nil {
    return bag
  }
  bag.err = bag.createManifest()
  return bag
}

func Exists(folder string) bool {
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
