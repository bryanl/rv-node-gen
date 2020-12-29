package rvnodegen

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
)

// Lister is an interface for listing objects.
type Lister interface {
	// List will return all objects across namespaces
	List(gvk schema.GroupVersionKind, selector labels.Selector) ([]*unstructured.Unstructured, error)
	// Get will attempt to retrieve assuming that name==key
	Get(gvk schema.GroupVersionKind, name string) (*unstructured.Unstructured, error)
	// ByNamespace will give you a GenericNamespaceLister for one namespace
	ByNamespace(namespace string) NamespaceLister
}

type lister struct {
	informerManager *InformerManager
}

var _ Lister = &lister{}

func newLister(informerManager *InformerManager) *lister {
	l := &lister{
		informerManager: informerManager,
	}
	return l
}

func (l *lister) List(gvk schema.GroupVersionKind, selector labels.Selector) ([]*unstructured.Unstructured, error) {
	informer, err := fetchInformer(l.informerManager, gvk)
	if err != nil {
		return nil, err
	}

	list, err := informer.Lister().List(selector)
	if err != nil {
		return nil, err
	}

	return toUnstructuredSlice(list)
}

func (l *lister) Get(gvk schema.GroupVersionKind, name string) (*unstructured.Unstructured, error) {
	informer, err := fetchInformer(l.informerManager, gvk)
	if err != nil {
		return nil, err
	}

	item, err := informer.Lister().Get(name)
	if err != nil {
		return nil, err
	}

	return toUnstructured(item)
}

func (l *lister) ByNamespace(namespace string) NamespaceLister {
	return newNamespaceLister(l.informerManager, namespace)
}

// NamespaceLister is a lister for listing namespace scoped objects.
type NamespaceLister interface {
	// List will return all objects in this namespace
	List(gvk schema.GroupVersionKind, selector labels.Selector) (ret []*unstructured.Unstructured, err error)
	// Get will attempt to retrieve by namespace and name
	Get(gvk schema.GroupVersionKind, name string) (*unstructured.Unstructured, error)
}

type namespaceLister struct {
	namespace       string
	informerManager *InformerManager
}

var _ NamespaceLister = &namespaceLister{}

func newNamespaceLister(informerManager *InformerManager, namespace string) *namespaceLister {
	n := &namespaceLister{
		namespace:       namespace,
		informerManager: informerManager,
	}
	return n
}

func (n *namespaceLister) List(gvk schema.GroupVersionKind, selector labels.Selector) (ret []*unstructured.Unstructured, err error) {
	informer, err := fetchInformer(n.informerManager, gvk)
	if err != nil {
		return nil, err
	}

	list, err := informer.Lister().ByNamespace(n.namespace).List(selector)
	if err != nil {
		return nil, err
	}

	return toUnstructuredSlice(list)
}

func (n *namespaceLister) Get(gvk schema.GroupVersionKind, name string) (*unstructured.Unstructured, error) {
	informer, err := fetchInformer(n.informerManager, gvk)
	if err != nil {
		return nil, err
	}

	item, err := informer.Lister().ByNamespace(n.namespace).Get(name)
	if err != nil {
		return nil, err
	}

	return toUnstructured(item)
}

func toUnstructuredSlice(in []runtime.Object) ([]*unstructured.Unstructured, error) {
	var out []*unstructured.Unstructured

	for i := range in {
		object, ok := in[i].(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("object is a %T", in[i])
		}

		out = append(out, object)
	}

	return out, nil
}

func toUnstructured(in runtime.Object) (*unstructured.Unstructured, error) {
	object, ok := in.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("object is a %T", in)
	}
	return object, nil
}

func fetchInformer(informerManager *InformerManager, gvk schema.GroupVersionKind) (informers.GenericInformer, error) {
	resource, err := informerManager.Resource(gvk)
	if err != nil {
		return nil, err
	}

	informer := informerManager.factory.ForResource(resource)
	return informer, nil
}
