package bagit

import (
  "ldpserver/fileio"
  "ldpserver/util"
)

type Bag struct {
  folder string
  dataFolder string 
}

func CreateBag(folder string) (Bag, error) {
  dataFolder := util.PathConcat(folder, "data")
  bag := Bag{folder: folder, dataFolder: dataFolder}
  if err := bag.writeBagItTxt(); err != nil {
    return Bag{}, err
  }
  if err := bag.writeManifest(); err != nil {
    return Bag{}, err
  }
  return bag, nil
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

func (bag Bag) writeBagItTxt() error {
  bagItText := "BagIt-Version: 0.97\nTag-File-Character-Encoding: UTF-8\n"
  bagIt := util.PathConcat(bag.folder, "bagit.txt")
  return fileio.WriteFile(bagIt, bagItText)
}

func (bag Bag) writeManifest() error {
  text := "data/meta.rdf TBD\ndata/meta.rdf.id TBD\n"
  manifestFile := util.PathConcat(bag.folder, "manifest-md5.txt")
  return fileio.WriteFile(manifestFile, text)
}
