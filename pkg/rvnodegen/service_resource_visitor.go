package rvnodegen

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ServiceResourceVisitor visits a service resource.
type ServiceResourceVisitor struct {
	lister Lister
}

var _ ResourceVisitor = &ServiceResourceVisitor{}

// NewServiceResourceVisitor creates an instance of ServiceResourceVisitor.
func NewServiceResourceVisitor(lister Lister) *ServiceResourceVisitor {
	s := &ServiceResourceVisitor{
		lister: lister,
	}
	return s
}

// Name is the name of the resource visitor.
func (s *ServiceResourceVisitor) Name() string {
	return "Service"
}

// Matches returns a group/version/kind that this resource visitor matches.
func (s *ServiceResourceVisitor) Matches(gvk schema.GroupVersionKind) bool {
	return serviceGVK.String() == gvk.String()
}

// Visit visits a service resource.
func (s *ServiceResourceVisitor) Visit(object *unstructured.Unstructured, node GraphNode, visitor *Visitor) (GraphNode, error) {
	serviceSelector, _, err := unstructured.NestedStringMap(object.Object, "spec", "selector")
	if err != nil {
		return GraphNode{}, err
	}

	set := labels.SelectorFromSet(serviceSelector)

	pods, err := s.lister.ByNamespace(object.GetNamespace()).List(podGVK, set)
	if err != nil {
		return GraphNode{}, err
	}

	ownersByID := map[string]*unstructured.Unstructured{}

	for _, pod := range pods {
		for _, ref := range pod.GetOwnerReferences() {
			gv, err := schema.ParseGroupVersion(ref.APIVersion)
			if err != nil {
				return GraphNode{}, err
			}

			gvk := schema.GroupVersionKind{
				Group:   gv.Group,
				Version: gv.Version,
				Kind:    ref.Kind,
			}

			owner, err := s.lister.ByNamespace(object.GetNamespace()).Get(gvk, ref.Name)
			if err != nil {
				return GraphNode{}, err
			}

			ownersByID[string(owner.GetUID())] = owner
		}
	}

	var owners []*unstructured.Unstructured

	for id, owner := range ownersByID {
		node.Targets = append(node.Targets, id)
		owners = append(owners, owner)
	}

	if err := visitor.Visit(owners...); err != nil {
		return GraphNode{}, err
	}

	return node, nil
}
