package rvnodegen

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GraphNode is a graph node that be used to create a graph.
type GraphNode struct {
	// ID is the id of the node.
	ID string `json:"id,omitempty"`
	// Name is the name of the node.
	Name string `json:"name,omitempty"`
	// Group is the node's group. It is optional.
	Group *string `json:"group,omitempty"`
	// Version is the node's version.
	Version string `json:"version,omitempty"`
	// Kind is the node's version.
	Kind string `json:"kind,omitempty"`
	// Parent is the node's parent. It is optional.
	Parent *string `json:"parent,omitempty"`
	// Extra is extra data for the node.
	Extra map[string]interface{} `json:"extra,omitempty"`
	// Targets are ids this node points to.
	Targets []string `json:"targets,omitempty"`
	// Keywords are keywords for the node.
	Keywords []string `json:"keywords,omitempty"`
	// IsGroup sets this node as a group.
	IsGroup *string `json:"isGroup,omitempty"`
	// NodeType is the the node type.
	NodeType NodeType `json:"nodeType,omitempty"`
}

// NodeType is the type of node.
type NodeType string

const (
	// NodeTypeWorkload is a workload node.
	NodeTypeWorkload NodeType = "workload"
	// NodeTypeNetworking is a networking node.
	NodeTypeNetworking NodeType = "networking"
	// NodeTypeConfiguration is a configuration node.
	NodeTypeConfiguration NodeType = "configuration"
	// NodeTypeCustomResource is a custom resource node
	NodeTypeCustomResource NodeType = "custom-resource"
)

func detectNodeType(lister Lister, object runtime.Object) (NodeType, error) {
	if lister == nil {
		panic("lister is nil")
	}

	if object == nil {
		return "", fmt.Errorf("object nil")
	}

	groupKind := object.GetObjectKind().GroupVersionKind().GroupKind()

	if isGroupKindMatch(groupKind, []schema.GroupVersionKind{daemonSetGVK, cronJobGVK, deploymentGVK,
		jobGVK, podGVK, replicaSetGVK, replicationControllerGVK, statefulSetGVK}) {
		return NodeTypeWorkload, nil
	}

	if isGroupKindMatch(groupKind, []schema.GroupVersionKind{ingressGVK, serviceGVK}) {
		return NodeTypeNetworking, nil
	}

	if isGroupKindMatch(groupKind, []schema.GroupVersionKind{configMapGVK, secretGVK, serviceAccountGVK}) {
		return NodeTypeConfiguration, nil
	}

	customResourceDefinitions, err := lister.List(crdGVK, labels.Everything())
	if err != nil {
		return "", fmt.Errorf("list custom resource definitions: %w", err)
	}

	for _, customResourceDefinition := range customResourceDefinitions {
		group, _, err := unstructured.NestedString(customResourceDefinition.Object, "spec", "group")
		if err != nil {
			return "", fmt.Errorf("get custom resource definition group: %w", err)
		}

		kind, _, err := unstructured.NestedString(customResourceDefinition.Object, "spec", "names", "kind")
		if err != nil {
			return "", fmt.Errorf("get custom resource definition kind: %w", err)
		}

		crdGroupKind := schema.GroupKind{Group: group, Kind: kind}
		if crdGroupKind.String() != groupKind.String() {
			continue
		}

		return NodeTypeCustomResource, nil
	}

	return "", fmt.Errorf("unknown group kind %s", groupKind)
}
