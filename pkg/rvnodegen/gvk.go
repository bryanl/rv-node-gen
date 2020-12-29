package rvnodegen

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
	podGVK            = schema.GroupVersionKind{Version: "v1", Kind: "Pod"}
	replicaSetGVK     = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "ReplicaSet"}
	serviceAccountGVK = schema.GroupVersionKind{Version: "v1", Kind: "ServiceAccount"}
	serviceGVK        = schema.GroupVersionKind{Version: "v1", Kind: "Service"}
)
