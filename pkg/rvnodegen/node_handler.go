package rvnodegen

import (
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"
)

type errorMessage struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type errorResponse struct {
	Error errorMessage `json:"error"`
}

type nodesResponse struct {
	Nodes []GraphNode `json:"nodes"`
}

func respondWithError(w http.ResponseWriter, err error, code int) {
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	resp := errorResponse{
		Error: errorMessage{
			Message: err.Error(),
			Status:  code,
		},
	}
	_ = enc.Encode(resp)
}

type NodeHandler struct {
	informerManager *InformerManager
}

var _ http.Handler = &NodeHandler{}

func NewNodeHandler(informerManager *InformerManager) *NodeHandler {
	nh := &NodeHandler{
		informerManager: informerManager,
	}

	return nh
}

func (nh NodeHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	gvk := schema.GroupVersionKind{
		Version: "v1",
		Kind:    "Pod",
	}

	gvr, err := nh.informerManager.Resource(gvk)
	if err != nil {
		respondWithError(w, fmt.Errorf("get resource for gvk (%s): %w", gvk, err), http.StatusInternalServerError)
		return
	}

	objects, err := nh.informerManager.Lister(gvr).ByNamespace("default").List(labels.Everything())
	if err != nil {
		respondWithError(w, fmt.Errorf("list pods: %w", err), http.StatusInternalServerError)
		return
	}

	list, err := toUnstructuredSlice(objects)
	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}

	emitter := NewNodeEmitter()
	visitor := NewVisitor(emitter, nh.informerManager)

	if err := visitor.Visit(list...); err != nil {
		respondWithError(w, fmt.Errorf("visit objects: %w", err), http.StatusInternalServerError)
		return
	}

	resp := nodesResponse{Nodes: emitter.Nodes()}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resp)
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
