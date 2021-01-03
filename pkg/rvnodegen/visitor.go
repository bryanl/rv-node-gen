package rvnodegen

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
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
	healthStatus     HealthStatuser
}

// NewVisitor creates an instance of a Visitor.
func NewVisitor(emitter Emitter, lister Lister, resourceVisitors []ResourceVisitor, options ...Option) (*Visitor, error) {
	opts := buildOptionConfig(options...)

	hs, err := opts.healthStatuserFactory(lister)
	if err != nil {
		return nil, fmt.Errorf("health status factor: %w", err)
	}

	v := &Visitor{
		emitter:          emitter,
		lister:           lister,
		resourceVisitors: resourceVisitors,
		visitedCache:     map[types.UID]bool{},
		healthStatus:     hs,
	}
	return v, nil
}

// Visit visits a set of objects. If the visit fails, it returns an error.
func (v *Visitor) Visit(isGroup bool, objects ...*unstructured.Unstructured) error {
	for i := range objects {
		object := objects[i].DeepCopy()

		if _, ok := v.visitedCache[object.GetUID()]; ok {
			continue
		}

		v.visitedCache[object.GetUID()] = true

		var parent *string

		var targets []string

		var ig *string
		if isGroup {
			ig = pointer.StringPtr("yes")
		}

		nodeType, err := detectNodeType(v.lister, object)
		if err != nil {
			return fmt.Errorf("detect node type: %w", err)
		}

		healthStatus, err := v.healthStatus.HealthStatus(object)
		if err != nil {
			return fmt.Errorf("health status: %w", err)
		}

		node := GraphNode{
			ID:           string(object.GetUID()),
			Label:        object.GetName(),
			Parent:       parent,
			Targets:      targets,
			IsGroup:      ig,
			NodeType:     nodeType,
			HealthStatus: healthStatus,
		}

		node, err = v.visitOwners(object, node)
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return err
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
		// node.Extra["podsOk"] = len(controlledPods)
		// node.Keywords = append(node.Keywords, "podSummary")
		pod := controlledPods[0]

		serviceAccountName, _, err := unstructured.NestedString(pod.Object, "spec", "serviceAccount")

		serviceAccount, err := v.lister.
			ByNamespace(object.GetNamespace()).
			Get(serviceAccountGVK, serviceAccountName)
		if err != nil {
			return GraphNode{}, err
		}
		node.Targets = append(node.Targets, string(serviceAccount.GetUID()))

		if err := v.Visit(false, serviceAccount); err != nil {
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

		node = setTarget(owner, node)
		n, isGroup := setParent(owner, node)
		node = n

		if err := v.Visit(isGroup, owner); err != nil {
			return GraphNode{}, err
		}
	}

	return node, nil
}

func ownsPods(owner *unstructured.Unstructured) bool {
	return isDeployment(owner) || isDaemonSet(owner) || isStatefulSet(owner)
}

func setParent(owner *unstructured.Unstructured, node GraphNode) (GraphNode, bool) {
	isOwner := ownsPods(owner)

	if isOwner {
		node.Parent = pointer.StringPtr(string(owner.GetUID()))
	}

	return node, isOwner
}

func setTarget(owner *unstructured.Unstructured, node GraphNode) GraphNode {
	if !ownsPods(owner) {
		node.Targets = append(node.Targets, string(owner.GetUID()))
	}

	return node
}
