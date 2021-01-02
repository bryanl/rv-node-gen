package rvnodegen

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	configMapGVK             = schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}
	crdGVK                   = schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"}
	cronJobGVK               = schema.GroupVersionKind{Group: "batch", Version: "v1beta1", Kind: "CronJob"}
	daemonSetGVK             = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"}
	deploymentGVK            = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
	ingressGVK               = schema.GroupVersionKind{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress"}
	jobGVK                   = schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"}
	podGVK                   = schema.GroupVersionKind{Version: "v1", Kind: "Pod"}
	replicaSetGVK            = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "ReplicaSet"}
	replicationControllerGVK = schema.GroupVersionKind{Version: "v1", Kind: "ReplicationController"}
	secretGVK                = schema.GroupVersionKind{Version: "v1", Kind: "Secret"}
	serviceAccountGVK        = schema.GroupVersionKind{Version: "v1", Kind: "ServiceAccount"}
	serviceGVK               = schema.GroupVersionKind{Version: "v1", Kind: "Service"}
	statefulSetGVK           = schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"}
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

func isGroupKindMatch(groupKind schema.GroupKind, list []schema.GroupVersionKind) bool {
	for i := range list {
		if list[i].GroupKind().String() == groupKind.String() {
			return true
		}
	}

	return false
}
