myfirstbag/
|-- data
|   \-- 27613-h
|       \-- images
|           \-- q172.png
|           \-- q172.txt
|-- manifest-md5.txt
|     49afbd86a1ca9f34b677a3f09655eae9 data/27613-h/images/q172.png
|     408ad21d50cef31da4df6d9ed81b01a7 data/27613-h/images/q172.txt
\-- bagit.txt
      BagIt-Version: 0.97
      Tag-File-Character-Encoding: UTF-8


/
/node1
/node2
/node3
/node3/node4
/node3/node5
/node3/bag


By prefixing all bag folders with "bag" I can prevent collisions 
(e.g. somebody creating a data bag overriding the data folder)

If I append the slug to the "bag" folder then I can read the 
actual file without opening the manifest.

root/
|-- bagit.txt
|-- manifest-md5.txt            (data/meta.rdf and data/meta.rdf.id)
|-- data/
|   |-- meta.rdf
|   |-- meta.rdf.id
|
|-- bag-node1/                  RDF source
|   |-- bagit.txt
|   |-- manifest-md5.txt        (data/meta.rdf)
|   |-- data/
|       |-- meta.rdf      
|
|-- bag-photo.jpg/              Non-RDF source
|   |-- bagit.txt
|   |-- manifest-md5.txt        (data/meta.rdf and data.photo.jpg)
|   |-- data/
|       |-- meta.rdf
|       |-- photo.jpg
|
|-- bag-node3/                  RDF Source
|   |-- bagit.txt
|   |-- manifest-md5.txt        (data/meta.rdf)
|   |-- data
|       |-- meta.rdf
|   |-- bag-node4               RDF Source
|       |-- bagit.txt
|       |-- manifest-md5.txt    (data/meta.rdf)
|       |-- data
|           |-- meta.rdf
|   |-- bag-something.txt       RDF Source
|       |-- bagit.txt
|       |-- manifest-md5.txt
|       |-- data
|           |-- meta.rdf
|           |-- something.txt
|
|   |-- bag-data                RDF Source, slug was "data"
|       |-- bagit.txt
|       |-- manifest-md5.txt    (data/meta.rdf)
|       |-- data
|           |-- meta.rdf
|   
