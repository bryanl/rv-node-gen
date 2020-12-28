package rvnodegen

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

var (
	// BannedResources are resources that will not be used in node generation.
	BannedResources = []schema.GroupVersionResource{
		{Group: "extensions", Version: "v1beta1", Resource: "ingresses"},
	}
)

// ResourceLister is an interface for listing resources.
type ResourceLister interface {
	// Lister returns a lister given group/version/resource.
	Lister(resource schema.GroupVersionResource) cache.GenericLister
	// Resource returns a resource given a group/version/kind.
	Resource(gvk schema.GroupVersionKind) (schema.GroupVersionResource, error)
}

// InformerManager in a manager for multiple informers.
type InformerManager struct {
	factory dynamicinformer.DynamicSharedInformerFactory
	mapping map[schema.GroupVersionKind]schema.GroupVersionResource
}

var _ ResourceLister = &InformerManager{}

// NewInformerManager creates an instance of InformerManager.
func NewInformerManager(client *Client) (*InformerManager, error) {
	defaultResync := 180 * time.Second

	informerFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		client.DynamicClient(),
		defaultResync,
		"", // this doesn't seem to matter
		nil)
	mapping, err := primeFactory(client.DiscoveryClient(), informerFactory)
	if err != nil {
		return nil, fmt.Errorf("prime informer factory: %w", err)
	}

	stopCh := make(chan struct{}, 1)
	informerFactory.Start(stopCh)

	i := &InformerManager{
		factory: informerFactory,
		mapping: mapping,
	}

	informerFactory.Start(stopCh)
	informerFactory.WaitForCacheSync(stopCh)

	return i, nil
}

// Lister returns a lister given a resource.
func (im *InformerManager) Lister(resource schema.GroupVersionResource) cache.GenericLister {
	return im.factory.ForResource(resource).Lister()
}

// Resource returns a resource given a group/version/kind.
func (im *InformerManager) Resource(gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	resource, ok := im.mapping[gvk]
	if !ok {
		return schema.GroupVersionResource{}, fmt.Errorf("not found")
	}

	return resource, nil
}

func primeFactory(
	discoveryClient discovery.DiscoveryInterface,
	factory dynamicinformer.DynamicSharedInformerFactory) (map[schema.GroupVersionKind]schema.GroupVersionResource, error) {
	mapping := map[schema.GroupVersionKind]schema.GroupVersionResource{}

	resourceLists, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("get server preferred resources: %w", err)
	}

	for _, resourceList := range resourceLists {
		groupVersion, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			return nil, fmt.Errorf("invalid group version %q: %w", resourceList.GroupVersion, err)
		}

		for _, apiResource := range resourceList.APIResources {
			if !stringsIncludes("watch", apiResource.Verbs) || !stringsIncludes("list", apiResource.Verbs) {
				continue
			}

			if !apiResource.Namespaced {
				continue
			}

			resource := schema.GroupVersionResource{
				Group:    groupVersion.Group,
				Version:  groupVersion.Version,
				Resource: apiResource.Name,
			}

			if isBanned(resource) {
				continue
			}

			factory.ForResource(resource)

			gvk := schema.GroupVersionKind{
				Group:   resource.Group,
				Version: resource.Version,
				Kind:    apiResource.Kind,
			}

			mapping[gvk] = resource
		}
	}

	return mapping, nil
}

func stringsIncludes(s string, sl []string) bool {
	for i := range sl {
		if s == sl[i] {
			return true
		}
	}

	return false
}

func isBanned(resource schema.GroupVersionResource) bool {
	for i := range BannedResources {
		if BannedResources[i].String() == resource.String() {
			return true
		}
	}

	return false
}
