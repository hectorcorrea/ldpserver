This is a mini LDP Server in Go.

Linked Data Platform (LDP) is a W3C recommendation that defines rules for how to
implement an HTTP API for read-write Linked Data. The official recommendation can
be found [here](http://www.w3.org/TR/ldp/).
You can also find a more gentle introduction to LDP in
[my blog](http://hectorcorrea.com/blog/introduction-to-ldp/67).

*Warning*: This is my sandbox project as I learn both Go and LDP. The code in this repo very likely does not follow Go's best practices and it certainly does not conform to the LDP spec (yet).


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

If you don't care about the source code, the fastest way to get started is to [download the executable for your platform](https://github.com/hectorcorrea/ldpserver/releases) from the release tab, make it an executable on your box, and run it.


## Operations supported
With the server running, you can use `cURL` to submit requests to it. For example, to fetch the root node

    curl localhost:9001

POST to the root (the Slug defaults to "node" + a sequential number)

    curl -X POST localhost:9001

Fetch the node created

    curl localhost:9001/node1

POST a non-RDF to the root

    curl -X POST --header "Content-Type: text/plain" --data "hello world" localhost:9001

    curl -X POST --header "Content-Type: image/jpeg" --data-binary "@filename.jpg" localhost:9001

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

Create an *LDP Direct Container* `/dc1` that uses `/node1` as its `membershipResource` and `someRel` as the member relation...

    curl -X POST --header "Content-Type: text/turtle" --header "Slug: dc1" -d "<> <http://www.w3.org/ns/ldp#hasMemberRelation> someRel ; <http://www.w3.org/ns/ldp#membershipResource> <http://localhost:9001/node1> ." localhost:9001

...add a node to the direct container

    curl -X POST --header "Slug: child1" localhost:9001/dc1

...fetch `/node1` and notice that it references `/dc1/child1` with the predicate `someRel` that we defined in the direct container:

    curl localhost:9001/dc1

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


## TODO
A lot.

* Add validation to make sure the data in the root node matches the URL (host:port) where the server is running.

* Support isMemberOfRelation in Direct Containers.

* Support Indirect Containers.


## LDP Test Suite
The W3C provides a test suite to make sure LDP server implementations meet a minimum criteria. The test suite can be found at http://w3c.github.io/ldp-testsuite/

In order to run the suite against this repo you need to do the following:

  1. Clone the ldp-testsuite repo: `git clone https://github.com/w3c/ldp-testsuite`
  1. Download maven from http://maven.apache.org/download.cgi
  1. Unzip maven to the `ldp-testsuite` folder
  1. `cd ldp-testsuite`
  1. Run `bin/mvn package`

...and then you can run the following command against the LDP Server to execute an individual test:

    java -jar target/ldp-testsuite-0.2.0-SNAPSHOT-shaded.jar --server http://localhost:9001 --test name_of_test_goes_here --basic

...or as follow to run all basic container tests (including support for non-RDF):

    java -jar target/ldp-testsuite-0.2.0-SNAPSHOT-shaded.jar --server http://localhost:9001 --basic --non-rdf

As of 1/9/2016 these are the results of all basic container tests (including support for non-RDF):

    LDP Test Suite
    Total tests run: 112, Failures: 4, Skips: 28


TODO: Document how to test DC and the results 97/5/27
