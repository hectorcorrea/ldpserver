package rdf

import (
// "errors"
// "log"
)

type Node struct {
	value    string
	children []Node
}

type AST struct {
	subjects []Node
}

func NewAST() AST {
	var tree AST
	return tree
}
