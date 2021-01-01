package rvnodegen

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
)

type NodeBuilder struct {
	lister Lister
}

func NewNodeBuilder(lister Lister) *NodeBuilder {
	n := &NodeBuilder{
		lister: lister,
	}
	return n
}

func (n *NodeBuilder) Build(namespace string) ([]GraphNode, error) {
	objects, err := n.lister.
		ByNamespace(namespace).
		List(podGVK, labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	resourceVisitors := ResourceVisitorsFactory(n.lister)
	emitter := NewNodeEmitter()
	visitor := NewVisitor(emitter, n.lister, resourceVisitors...)

	if err := visitor.Visit(false, objects...); err != nil {
		return nil, fmt.Errorf("visit objects: %w", err)
	}

	return emitter.Nodes(), nil

}
