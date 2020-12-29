package rvnodegen

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/pointer"
)

// Visitor visits resources and emits data.
type Visitor struct {
	emitter          Emitter
	lister           Lister
	resourceVisitors []ResourceVisitor
}

// NewVisitor creates an instance of a Visitor.
func NewVisitor(emitter Emitter, lister Lister, resourceVisitors ...ResourceVisitor) *Visitor {
	v := &Visitor{
		emitter:          emitter,
		lister:           lister,
		resourceVisitors: resourceVisitors,
	}
	return v
}

// Visit visits a set of objects. If the visit fails, it returns an error.
func (v *Visitor) Visit(objects ...*unstructured.Unstructured) error {
	for i := range objects {
		object := objects[i].DeepCopy()

		var group *string

		if g := object.GroupVersionKind().Group; g != "" {
			group = pointer.StringPtr(g)
		}

		var parent *string

		extra := map[string]interface{}{}
		var targets []string

		node := GraphNode{
			ID:      string(object.GetUID()),
			Name:    object.GetName(),
			Group:   group,
			Version: object.GroupVersionKind().Version,
			Kind:    object.GroupVersionKind().Kind,
			Parent:  parent,
			Extra:   extra,
			Targets: targets,
		}

		node, err := v.visitOwners(object, node)
		if err != nil {
			return fmt.Errorf("unable to visit owner for (%s) %s %s: %w",
				object.GetNamespace(), object.GroupVersionKind(), object.GetName(), err)
		}

		for _, resourceVisitor := range v.resourceVisitors {
			if resourceVisitor.Matches(object.GroupVersionKind()) {
				node, err = resourceVisitor.Visit(object, node, v)
				if err != nil {
					return fmt.Errorf("resource visitor %s: %w", resourceVisitor.Name(), err)
				}
			}
		}

		if err := v.emitter.Emit(object, node); err != nil {
			return fmt.Errorf("emit node: %w", err)
		}
	}

	return nil
}

func (v *Visitor) visitOwners(object *unstructured.Unstructured, node GraphNode) (GraphNode, error) {
	for _, ref := range object.GetOwnerReferences() {
		gv, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			return GraphNode{}, fmt.Errorf("parse API version %q: %w", ref.APIVersion, err)
		}

		gvk := schema.GroupVersionKind{
			Group:   gv.Group,
			Version: gv.Version,
			Kind:    ref.Kind,
		}

		owner, err := v.lister.ByNamespace(object.GetNamespace()).Get(gvk, ref.Name)
		if err != nil {
			return GraphNode{}, fmt.Errorf("get owner: %w", err)
		}

		if err := v.Visit(owner); err != nil {
			return GraphNode{}, err
		}

		switch {
		case isDeployment(owner):
			node.Parent = pointer.StringPtr(string(owner.GetUID()))
		case isDaemonSet(owner):
			node.Parent = pointer.StringPtr(string(owner.GetUID()))
		default:
			node.Targets = append(node.Targets, string(ref.UID))
		}
	}

	return node, nil
}
