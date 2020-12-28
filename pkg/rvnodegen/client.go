package rvnodegen

import (
	"fmt"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	restclient "k8s.io/client-go/rest"

	// import all the Kubernetes auth packages.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// Client is a client for communicating with a Kubernetes cluster.
type Client struct {
	discoveryClient discovery.DiscoveryInterface
	dynamicClient   dynamic.Interface
}

// NewClient creates an instance of Client.
func NewClient(restConfig *restclient.Config, options ...Option) (*Client, error) {
	oc := buildOptionConfig(options...)

	discoveryClient, err := disk.NewCachedDiscoveryClientForConfig(
		restConfig,
		oc.discoveryCacheDir,
		oc.httpCacheDir,
		oc.discoveryTTL)
	if err != nil {
		return nil, fmt.Errorf("create discovery client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("create dynamic client: %w", err)
	}

	c := &Client{
		discoveryClient: discoveryClient,
		dynamicClient:   dynamicClient,
	}

	return c, nil
}

// DiscoveryClient returns a discovery client.
func (c *Client) DiscoveryClient() discovery.DiscoveryInterface {
	return c.discoveryClient
}

// DynamicClient returns a dynamic client.
func (c *Client) DynamicClient() dynamic.Interface {
	return c.dynamicClient
}
