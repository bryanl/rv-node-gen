package rvnodegen

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ServiceAccountVisitor visits service account resource.
type ServiceAccountVisitor struct {
	lister Lister
}

var _ ResourceVisitor = &ServiceAccountVisitor{}

// NewServiceAccountVisitor creates an instance of ServiceAccountVisitor.
func NewServiceAccountVisitor(lister Lister) *ServiceAccountVisitor {
	s := &ServiceAccountVisitor{
		lister: lister,
	}
	return s
}

// Name is the name of the resource visitor.
func (s *ServiceAccountVisitor) Name() string {
	return "ServiceAccount"
}

// Matches returns a group/version/kind that this resource visitor matches.
func (s *ServiceAccountVisitor) Matches(gvk schema.GroupVersionKind) bool {
	return serviceAccountGVK.String() == gvk.String()
}

// Visit visits a service account resource.
func (s *ServiceAccountVisitor) Visit(object *unstructured.Unstructured, node GraphNode, visitor *Visitor) (GraphNode, error) {
	secrets, found, err := unstructured.NestedSlice(object.Object, "secrets")
	if err != nil {
		return GraphNode{}, err
	}

	if !found {
		return node, nil
	}

	for i := range secrets {
		secretConfig := secrets[i].(map[string]interface{})
		name, _, err := unstructured.NestedString(secretConfig, "name")
		if err != nil {
			return GraphNode{}, err
		}

		secret, err := s.lister.
			ByNamespace(object.GetNamespace()).
			Get(secretGVK, name)
		if err != nil {
			return GraphNode{}, err
		}

		if err := visitor.Visit(false, secret); err != nil {
			return GraphNode{}, err
		}

		node.Targets = append(node.Targets, string(secret.GetUID()))
	}

	return node, nil
}
