package rvnodegen

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
)

// Visitor visits resources and emits data.
type Visitor struct {
	emitter          Emitter
	lister           Lister
	resourceVisitors []ResourceVisitor
	visitedCache     map[types.UID]bool
}

// NewVisitor creates an instance of a Visitor.
func NewVisitor(emitter Emitter, lister Lister, resourceVisitors ...ResourceVisitor) *Visitor {
	v := &Visitor{
		emitter:          emitter,
		lister:           lister,
		resourceVisitors: resourceVisitors,
		visitedCache:     map[types.UID]bool{},
	}
	return v
}

// Visit visits a set of objects. If the visit fails, it returns an error.
func (v *Visitor) Visit(objects ...*unstructured.Unstructured) error {
	for i := range objects {
		object := objects[i].DeepCopy()

		if _, ok := v.visitedCache[object.GetUID()]; ok {
			continue
		}

		v.visitedCache[object.GetUID()] = true

		var group *string

		if g := object.GroupVersionKind().Group; g != "" {
			group = pointer.StringPtr(g)
		}

		var parent *string

		extra := map[string]interface{}{
			"complex": map[string]interface{}{
				"number": 1,
				"array":  []string{"1", "2", "3"},
				"bool":   false,
			},
		}
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

		node, err = v.checkForOwnedPods(object, node)
		if err != nil {
			return fmt.Errorf("check for owned pods for (%s) %s %s: %w",
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

func (v *Visitor) checkForOwnedPods(object *unstructured.Unstructured, node GraphNode) (GraphNode, error) {
	pods, err := v.lister.ByNamespace(object.GetNamespace()).List(podGVK, labels.Everything())
	if err != nil {
		return GraphNode{}, err
	}

	var controlledPods []*unstructured.Unstructured
	for _, pod := range pods {
		if metav1.IsControlledBy(pod, object) {
			controlledPods = append(controlledPods, pod)
		}
	}

	if len(controlledPods) > 0 {
		node.Extra["podsOk"] = len(controlledPods)
		node.Keywords = append(node.Keywords, "podSummary")
		pod := controlledPods[0]

		serviceAccountName, _, err := unstructured.NestedString(pod.Object, "spec", "serviceAccount")

		serviceAccount, err := v.lister.
			ByNamespace(object.GetNamespace()).
			Get(serviceAccountGVK, serviceAccountName)
		if err != nil {
			return GraphNode{}, err
		}
		node.Targets = append(node.Targets, string(serviceAccount.GetUID()))

		if err := v.Visit(serviceAccount); err != nil {
			return GraphNode{}, err
		}
	}

	return node, nil
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

		node = setParent(owner, node)
		node = setTarget(owner, node)
	}

	return node, nil
}

func ownsPods(owner *unstructured.Unstructured) bool {
	return isDeployment(owner) || isDaemonSet(owner) || isStatefulSet(owner)
}

func setParent(owner *unstructured.Unstructured, node GraphNode) GraphNode {
	if ownsPods(owner) {
		node.Parent = pointer.StringPtr(string(owner.GetUID()))
		node.Keywords = append(node.Keywords, "workloadOwner")
	}

	return node
}

func setTarget(owner *unstructured.Unstructured, node GraphNode) GraphNode {
	if !ownsPods(owner) {
		node.Targets = append(node.Targets, string(owner.GetUID()))
	}

	return node
}
