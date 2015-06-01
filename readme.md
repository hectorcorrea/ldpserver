This is a mini LDP Server in Go.

LDP stands for Linked Data Platform and the W3 spec for it can be found [here]( http://www.w3.org/TR/ldp/)

*Warning*: This is my sandbox project as I learn both Go and LDP. The code in this repo very likely does not follow Go's best practices and it certainly does not conform to the LDP spec.


## Compile and run the server
If Go is installed on your machine:

    cd ~/src
    git clone git@github.com:hectorcorrea/ldpserver.git
    cd ldpserver
    go build
    ./ldpserver

If you are new to Go follow these steps instead:

    # Download and install Go from: http://golang.org/doc/install
    # 
    # Go is very picky about the location of the code (e.g. the code must be 
    # inside an src folder.) Here is a setup that will work with minimal effort 
    # and configuration on your part. You can skip the first step if you 
    # already have an ~/src folder.
    #
    mkdir ~/src
    export GOPATH=~/
    cd ~/src
    git clone git@github.com:hectorcorrea/ldpserver.git
    cd ldpserver
    go build
    ./ldpserver


## Operations supported
With the server running, you can use `cURL` to submit requests to it. For example, to fetch the root node

    curl locahost:9001

POST to the root (the Slug defaults to "node" + a sequential number)

    curl -X POST localhost:9001

Fetch the node created

    curl localhost:9001/node1

POST a non-RDF to the root

    curl -X POST --header "Link: http://www.w3.org/ns/ldp#NonRDFSource; rel=\"type\"" --data "hello world" localhost:9001

    curl -X POST --header "Link: http://www.w3.org/ns/ldp#NonRDFSource; rel=\"type\"" --data-binary "@filename" localhost:9001

Fetch the non-RDF created

    curl localhost:9001/node2

HTTP HEAD operations are supported

    curl -I localhost:9001/
    curl -I localhost:9001/node1
    curl -I localhost:9001/node2

Add an RDF source to add a child node (you can only add to RDF sources)

    curl -X POST localhost:9001/node1

See that the child was added

    curl localhost:9001/node1

Fetch the child

    curl localhost:9001/node1/node3

Create a node with a custom Slug

    curl -X POST --header "Slug: demo" localhost:9001

Fetch node created

    curl localhost:9001/demo

Create an *LDP Direct Container* that uses `/node1` as its `membershipResource` (notice the `$'text'` syntax to preserve carriage returns in the triples) 

    curl -X POST -d $'<> <http://www.w3.org/ns/ldp#hasMemberRelation> <someRel> .\n<> <http://www.w3.org/ns/ldp#membershipResource> <http://localhost:9001/node1> .\n' localhost:9001


## Demo
Take a look at `demo.sh` file for an example of a shell script that executes some of the operations supported. To run this demo make sure the LDP Server is running in a separate terminal window, for example:

    # Run the LDP Server in one terminal window
    ./ldpserver

    # Run the demo script in a separate terminal window
    ./demo.sh


## Storage
Every resource (RDF or non-RDF) is saved in a folder inside the data folder.

Every RDF source is saved on its own folder with single file inside of it. This file is always `meta.rdf` and it has the triples of the node.

Non-RDF are also saved on their own folder and with a `meta.rdf` file for their metadata but also a file `data.bin` with the non-RDF content.

For example, if we have two nodes (blog1 and blog2) and blog1 is an RDF node and blog2 is a non-RDF then the data structure would look as follow:

    /data/meta.rdf          (root node)
    /data/blog1/meta.rdf    (RDF for blog1)
    /data/blog2/meta.rdf    (RDF for blog2)
    /data/blog2/data.bin    (binary for blog2)


## Overview of the Code

* `main.go` is the launcher program. It's only job is to kick off the web server.
* `web/web.go` is the web server. It's job is to handle HTTP requests and responses. This is the only part of the code that is aware of the web.
* `server/server.go` handles most of the operations like creating new nodes and fetching existing ones.
* `ldp/node.go` handles operations at the individual node level (fetching and saving.)
* `rdf/` contains utilities to parse and update RDF triples and graphs.


## Misc Notes
Empty subjects and objects (<>) are only accepted when creating new nodes (via HTTP POST) and they are immediately converted to the actual URI that they represent.


## TODO
A lot. 

* Support isMemberOfRelation in Direct Containers.

* Support Indirect Containers. 

* Support HTTP PUT, PATCH, and DELETE. 

* I am currently using n-triples rather than turtle because n-triples require less parsing (e.g. no prefixes to be aware of). This should eventually be changed to support and default to turtle.

* Make sure the ntriples pass a minimum validation. For starters take a look at this set: http://www.w3.org/2000/10/rdf-tests/rdfcore/ntriples/test.nt

* Provide a mechanism to fetch the meta data for a non-RDF (e.g. via a query string or an HTTP header parameter)

* Use BagIt file format to store data (http://en.wikipedia.org/wiki/BagIt)

* Make sure the proper links are included in the HTTP response for all kind of resources. 