package rdf

import (
	// "errors"
	"log"
)

type Node struct {
	value    string
	children []*Node
}

func NewNode(value string) Node {
	return Node{value: value}
}

func (node *Node) AddChild(value string) *Node {
	child := NewNode(value)
	node.children = append(node.children, &child)
	return &child
}

func (subject *Node) Render() string {
	triples := ""
	log.Printf("1. s=%s", subject.value)
	for _, predicate := range subject.children {
		log.Printf("2. p=%s", predicate.value)
		for _, object := range predicate.children {
			log.Printf("3. o=%s", object.value)
			triples += subject.value + " " + predicate.value + " " + object.value + " .\n"
		}
	}
	return triples
}
