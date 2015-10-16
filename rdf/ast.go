package rdf

type Tree struct {
	nodes []*Node
}

type Node struct {
	value    string
	children []*Node
}

func NewTree() Tree {
	return Tree{}
}

func NewNode(value string) Node {
	return Node{value: value}
}

func (tree *Tree) AddNode(value string) *Node {
	node := Node{value: value}
	tree.nodes = append(tree.nodes, &node)
	return &node
}

func (node *Node) AddChild(value string) *Node {
	child := Node{value: value}
	node.children = append(node.children, &child)
	return &child
}

func (subject *Node) Render() string {
	triples := ""
	for _, predicate := range subject.children {
		for _, object := range predicate.children {
			triples += subject.value + " " + predicate.value + " " + object.value + " .\n"
		}
	}
	return triples
}

func (subject *Node) RenderTriples() []Triple {
	triples := []Triple{}
	for _, predicate := range subject.children {
		for _, object := range predicate.children {
			triple := NewTriple(subject.value, predicate.value, object.value)
			triples = append(triples, triple)
		}
	}
	return triples
}

func (tree *Tree) Render() string {
	triples := ""
	for _, node := range tree.nodes {
		triples += node.Render()
	}
	return triples
}
