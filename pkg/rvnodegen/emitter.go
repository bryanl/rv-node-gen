package rvnodegen

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Emitter is an interface that provides a way to emit data. It is used with Visitor.
type Emitter interface {
	// Emit emits an object.
	Emit(object *unstructured.Unstructured, graphNode GraphNode) error
}

// NodeEmitter is an emitter that contains graph nodes.
type NodeEmitter struct {
	nodes []GraphNode
}

var _ Emitter = &NodeEmitter{}

// NewNodeEmitter creates an instance of NodeEmitter.
func NewNodeEmitter() *NodeEmitter {
	n := &NodeEmitter{}
	return n
}

// Emit emits a graph node for an object.
func (n *NodeEmitter) Emit(object *unstructured.Unstructured, graphNode GraphNode) error {
	if isPod(object) {
		// TODO search existing nodes for a pod with the same selector
		return nil
	}

	n.nodes = append(n.nodes, graphNode)

	return nil
}

// Nodes returns the graph nodes.
func (n *NodeEmitter) Nodes() []GraphNode {
	return n.nodes
}

func isPod(object *unstructured.Unstructured) bool {
	podGroupVersionKind := schema.GroupVersionKind{Version: "v1", Kind: "Pod"}
	return object.GroupVersionKind().String() == podGroupVersionKind.String()
}
