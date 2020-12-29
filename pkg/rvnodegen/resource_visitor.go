package rvnodegen

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
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
		NewReplicaSetResourceVisitor(lister),
	}
}

// ReplicaSetResourceVisitor visits a replica set.
type ReplicaSetResourceVisitor struct {
	lister Lister
}

var _ ResourceVisitor = &ReplicaSetResourceVisitor{}

// NewReplicaSetResourceVisitor creates an instance of ReplicaSetResourceVisitor.
func NewReplicaSetResourceVisitor(lister Lister) *ReplicaSetResourceVisitor {
	r := &ReplicaSetResourceVisitor{
		lister: lister,
	}
	return r
}

// Name is the name of the resource visitor.
func (r *ReplicaSetResourceVisitor) Name() string {
	return "ReplicaSet"
}

// Matches returns a group/version/kind that this resource visitor pertains to.
func (r *ReplicaSetResourceVisitor) Matches(gvk schema.GroupVersionKind) bool {
	return replicaSetGVK.String() == gvk.String()
}

// Visit visits a replica set resource.
func (r *ReplicaSetResourceVisitor) Visit(object *unstructured.Unstructured, node GraphNode, visitor *Visitor) (GraphNode, error) {
	selector := labels.SelectorFromSet(object.GetLabels())
	objects, err := r.lister.ByNamespace(object.GetNamespace()).List(podGVK, selector)
	if err != nil {
		return GraphNode{}, err
	}

	node.Extra["podsOk"] = len(objects)

	if len(objects) > 0 {
		pod := objects[0]

		serviceAccountName, _, err := unstructured.NestedString(pod.Object, "spec", "serviceAccount")

		serviceAccount, err := r.lister.
			ByNamespace(object.GetNamespace()).
			Get(serviceAccountGVK, serviceAccountName)
		if err != nil {
			return GraphNode{}, err
		}
		node.Targets = append(node.Targets, string(serviceAccount.GetUID()))

		if err := visitor.Visit(serviceAccount); err != nil {
			return GraphNode{}, err
		}
	}

	return node, nil
}
