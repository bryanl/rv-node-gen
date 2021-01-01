package rvnodegen

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PodResourceVisitor visits a pod.
type PodResourceVisitor struct {
	lister    Lister
	seenCache map[string]bool
}

var _ ResourceVisitor = &PodResourceVisitor{}

// NewPodResourceVisitor creates an instance of PodResourceVisitor.
func NewPodResourceVisitor(lister Lister) *PodResourceVisitor {
	p := &PodResourceVisitor{
		lister:    lister,
		seenCache: map[string]bool{},
	}
	return p
}

// Name is the name of the resource visitor.
func (p *PodResourceVisitor) Name() string {
	return "Pod"
}

// Matches returns a group/version/kind that this resource visitor matches.
func (p *PodResourceVisitor) Matches(gvk schema.GroupVersionKind) bool {
	return podGVK.String() == gvk.String()
}

// Visit visits a pod resource.
func (p *PodResourceVisitor) Visit(object *unstructured.Unstructured, node GraphNode, visitor *Visitor) (GraphNode, error) {
	hash, err := stringMapHash(object.GetLabels())
	if err != nil {
		return GraphNode{}, err
	}

	if _, ok := p.seenCache[hash]; ok {
		// a similar pod has been processed
		return node, nil
	}

	p.seenCache[hash] = true

	services, err := p.lister.ByNamespace(object.GetNamespace()).List(serviceGVK, labels.Everything())
	if err != nil {
		return GraphNode{}, err
	}

	podLabels := labels.Set(object.GetLabels())

	for _, service := range services {
		serviceSelector, found, err := unstructured.NestedStringMap(service.Object, "spec", "selector")
		if err != nil {
			return GraphNode{}, err
		}

		if !found {
			continue
		}

		set := labels.SelectorFromSet(serviceSelector)

		if !set.Matches(podLabels) {
			continue
		}

		if err := visitor.Visit(false, service); err != nil {
			return GraphNode{}, err
		}
	}

	return node, nil
}

func stringMapHash(stringMap map[string]string) (string, error) {
	h := sha256.New()

	if err := json.NewEncoder(h).Encode(stringMap); err != nil {
		return "", err
	}

	return fmt.Sprint(hex.EncodeToString(h.Sum(nil))), nil

}
