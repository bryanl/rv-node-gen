package rvnodegen

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
}
