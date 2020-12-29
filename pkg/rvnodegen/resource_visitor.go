package rvnodegen

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResourceVisitor is a resource specific visitor.
type ResourceVisitor interface {
	// Name is the name of the resource visitor.
	Name() string
	// Matches returns a group/version/kind that this resource visitor pertains to.
	Matches(gvk schema.GroupVersionKind) bool
	// Visit visits an a resource.
	Visit(object *unstructured.Unstructured, node GraphNode, visitor *Visitor) (GraphNode, error)
}

// ResourceVisitorsFactory creates a slice of ResourceVisitors.
func ResourceVisitorsFactory(lister Lister) []ResourceVisitor {
	return []ResourceVisitor{
		NewPodResourceVisitor(lister),
		NewServiceAccountVisitor(lister),
		NewServiceResourceVisitor(lister),
	}
}
