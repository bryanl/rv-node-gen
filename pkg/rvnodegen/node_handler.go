package rvnodegen

import (
	"net/http"

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

// NodeHandler is a HTTP handler for generating nodes.
type NodeHandler struct {
	lister Lister
}

var _ http.Handler = &NodeHandler{}

// NewNodeHandler creates an instance of NodeHandler.
func NewNodeHandler(lister Lister) *NodeHandler {
	nh := &NodeHandler{
		lister: lister,
	}

	return nh
}

func (nh *NodeHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	nb := NewNodeBuilder(nh.lister)
	nodes, err := nb.Build("default")
	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	resp := nodesResponse{Nodes: nodes}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resp)
}
