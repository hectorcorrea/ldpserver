package rdf

type SubjectNode struct {
	value      string
	predicates []*PredicateNode
}

type PredicateNode struct {
	value   string
	objects []string
}

func NewSubjectNode(value string) SubjectNode {
	return SubjectNode{value: value}
}

func NewPredicateNode(value string) PredicateNode {
	return PredicateNode{value: value}
}

func (subject *SubjectNode) AddPredicate(value string) *PredicateNode {
	predicate := PredicateNode{value: value}
	subject.predicates = append(subject.predicates, &predicate)
	return &predicate
}

func (predicate *PredicateNode) AddObject(object string) {
	predicate.objects = append(predicate.objects, object)
}

func (subject *SubjectNode) Render() string {
	triples := ""
	for _, predicate := range subject.predicates {
		for _, object := range predicate.objects {
			triples += subject.value + " " + predicate.value + " " + object + " .\n"
		}
	}
	return triples
}

func (subject *SubjectNode) RenderTriples() []Triple {
	triples := []Triple{}
	for _, predicate := range subject.predicates {
		for _, object := range predicate.objects {
			triple := NewTriple(subject.value, predicate.value, object)
			triples = append(triples, triple)
		}
	}
	return triples
}
