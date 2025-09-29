package printer

import (
	"context"
	"io"
	"log"

	"entgo.io/ent/entc/gen"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

// Graphviz writes a DOT graph visualization of the ent graph to the given writer.
func Graphviz(ctx context.Context, w io.Writer, g *gen.Graph) {
	gv, err := graphviz.New(ctx)
	if err != nil {
		log.Fatalf("failed to create graphviz: %v", err)
	}

	graph, err := gv.Graph()
	if err != nil {
		log.Fatalf("failed to create graph: %v", err)
	}
	defer func() {
		if err := graph.Close(); err != nil {
			log.Printf("failed to close graph: %v", err)
		}
		gv.Close()
	}()

	// Map to keep track of node names to cgraph.Node
	nodeMap := make(map[string]*cgraph.Node)

	// Add nodes
	for _, n := range g.Nodes {
		node, err := graph.CreateNodeByName(n.Name)
		if err != nil {
			log.Fatalf("failed to create node %q: %v", n.Name, err)
		}
		nodeMap[n.Name] = node
	}

	// Add edges
	for _, n := range g.Nodes {
		for _, e := range n.Edges {
			// Only draw edge if the target node exists
			if target, ok := nodeMap[e.Type.Name]; ok {
				_, err := graph.CreateEdgeByName(e.Name, nodeMap[n.Name], target)
				if err != nil {
					log.Fatalf("failed to create edge %q -> %q: %v", n.Name, e.Type.Name, err)
				}
				// TODO: Customize edge appearance based on relation type
				// ce.SetLabel(e.Label())
			}
		}
	}

	// Write DOT output
	if err := gv.Render(ctx, graph, graphviz.XDOT, w); err != nil {
		log.Fatalf("failed to render graph: %v", err)
	}
}
