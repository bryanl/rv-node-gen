package rvnodegen

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	daemonSetGVK      = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"}
	deploymentGVK     = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
	podGVK            = schema.GroupVersionKind{Version: "v1", Kind: "Pod"}
	secretGVK         = schema.GroupVersionKind{Version: "v1", Kind: "Secret"}
	serviceAccountGVK = schema.GroupVersionKind{Version: "v1", Kind: "ServiceAccount"}
	serviceGVK        = schema.GroupVersionKind{Version: "v1", Kind: "Service"}
	statefulSetGVK    = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"}
)

func isPod(object *unstructured.Unstructured) bool {
	return object.GroupVersionKind().String() == podGVK.String()
}

func isDeployment(object *unstructured.Unstructured) bool {
	return object.GroupVersionKind().String() == deploymentGVK.String()
}

func isDaemonSet(object *unstructured.Unstructured) bool {
	return object.GroupVersionKind().String() == daemonSetGVK.String()
}

func isStatefulSet(object *unstructured.Unstructured) bool {
	return object.GroupVersionKind().String() == statefulSetGVK.String()
}
