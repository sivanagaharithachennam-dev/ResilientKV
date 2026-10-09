package routing

import "fmt"

// Router selects a server for a key.
type Router struct {
	nodes []string
}

// NewRouter creates a router.
func NewRouter(nodes []string) *Router {
	return &Router{
		nodes: nodes,
	}
}

// Route selects a node based on the key.
func (r *Router) Route(key string) string {
	if len(r.nodes) == 0 {
		return ""
	}

	var hash uint32

	for i := 0; i < len(key); i++ {
		hash = hash*31 + uint32(key[i])
	}

	index := int(hash % uint32(len(r.nodes)))

	return r.nodes[index]
}

// Describe returns the selected node.
func (r *Router) Describe(key string) string {
	return fmt.Sprintf(
		"key=%s routed_to=%s",
		key,
		r.Route(key),
	)
}
