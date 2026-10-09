package routing

import "testing"

func TestRouter(t *testing.T) {
	router := NewRouter([]string{
		"node1:50051",
		"node2:50051",
		"node3:50051",
	})

	node := router.Route("product")

	if node == "" {
		t.Fatal("expected a node")
	}

	t.Logf("product routed to %s", node)
}
