package similgraph

import (
	"math"

	"github.com/pkg/errors"
)

// New creates a new similgraph from the edges iterator; vertexACount, vertexBCount and edgeCount contain size estimates, used for efficient memory allocation.
func New(edges func() (info Edge, ok bool), vertexACount, vertexBCount, edgeCount int) (g *SimilGraph, newoldVertexA []uint32, err error) {

	g = &SimilGraph{make([]implicitEdge, 0, 2*edgeCount), make([]uint32, 0, vertexACount+vertexBCount+1), uint32(0)}
	newoldVertexA = make([]uint32, 0, vertexACount)

	if err := normalizeEdgeWeight(edges, g, &newoldVertexA); err != nil {
		return &SimilGraph{}, nil, err
	}

	if g.vertexCount > 0 {
		addInvertedEdges(g)
	}

	return g, newoldVertexA, nil
}

func normalizeEdgeWeight(edges func() (info Edge, ok bool), g *SimilGraph, newoldVertexA *[]uint32) error {
	e0, ok := edges()
	if !ok {
		return nil
	}

	if e0.Weight == 0 {
		return errors.Errorf("similgraph: input edge with zero weight e%v = %v", len(g.bigraphEdges), e0)
	}

	*newoldVertexA = append(*newoldVertexA, e0.VertexA)
	g.bigraphEdges = append(g.bigraphEdges, implicitEdge{e0.VertexB, e0.Weight})
	g.vertexSlices = append(g.vertexSlices, 0)
	groupQNorm := float64(e0.Weight) * float64(e0.Weight)

	for e1, ok := edges(); ok; e1, ok = edges() {
		if e0.VertexA > e1.VertexA || (e0.VertexA == e1.VertexA && e0.VertexB >= e1.VertexB) {
			return errors.Errorf("similgraph: unsorted input edges, should be e%v < e%v, but %v >= %v", len(g.bigraphEdges)-1, len(g.bigraphEdges), e0, e1)
		}
		if e1.Weight == 0 {
			return errors.Errorf("similgraph: input edge with zero weight e%v = %v", len(g.bigraphEdges), e1)
		}
		if e0.VertexA != e1.VertexA {

			lastGroup := g.bigraphEdges[g.vertexSlices[len(g.vertexSlices)-1]:]
			updateWeight(lastGroup, groupQNorm)
			g.vertexSlices = append(g.vertexSlices, uint32(len(g.bigraphEdges)))

			*newoldVertexA = append(*newoldVertexA, e1.VertexA)
			groupQNorm = 0
		}
		g.bigraphEdges = append(g.bigraphEdges, implicitEdge{e1.VertexB, e1.Weight})
		groupQNorm += float64(e1.Weight) * float64(e1.Weight)
		e0 = e1
	}
	lastGroup := g.bigraphEdges[g.vertexSlices[len(g.vertexSlices)-1]:]
	updateWeight(lastGroup, groupQNorm)
	g.vertexSlices = append(g.vertexSlices, uint32(len(g.bigraphEdges)))
	g.vertexCount = uint32(len(g.vertexSlices) - 1)
	return nil
}

func addInvertedEdges(g *SimilGraph) {
	//eventual copy into a bigger array to permit id refactor and add inverterted graph at the same time
	if 2*len(g.bigraphEdges) > cap(g.bigraphEdges) {
		newbigraphEdges := make([]implicitEdge, 0, 2*len(g.bigraphEdges))
		g.bigraphEdges = append(newbigraphEdges, g.bigraphEdges...)
	}

	pes := make([]pivotedEdgesSpan, 0, g.vertexCount)
	for v := uint32(0); v < g.vertexCount; v++ {
		group := g.bigraphEdges[g.vertexSlices[v]:g.vertexSlices[v+1]]
		pes = append(pes, pivotedEdgesSpan{implicitEdge{v, 1}, group})
	}

	h := redgeIterMergeFrom(pES2REdgeIter(pes...)...)
	for re, ok := h.Peek(); ok; re, ok = h.Peek() {
		oldID, newID := re.Edge.Vertex, uint32(len(g.vertexSlices)-1)
		for re, ok := h.Peek(); ok && re.Edge.Vertex == oldID; re, ok = h.Peek() {
			h.Next()
			re.Edge.Vertex = newID
			g.bigraphEdges = append(g.bigraphEdges, implicitEdge{re.Pivot.Vertex, re.Derived.Weight})
		}
		g.vertexSlices = append(g.vertexSlices, uint32(len(g.bigraphEdges)))
	}
}

func updateWeight(group []implicitEdge, qNorm float64) {
	norm := float32(math.Sqrt(qNorm))
	for i := range group {
		group[i].Weight /= norm
	}
}

func pES2REdgeIter(pes ...pivotedEdgesSpan) (nexts []func() (rEdge, bool)) {
	for _, s := range pes {
		s := s
		nexts = append(nexts, func() (e rEdge, ok bool) {
			if len(s.Edges) == 0 {
				return
			}
			p, ie := &s.Pivot, &s.Edges[0]
			s.Edges = s.Edges[1:]
			return rEdge{Edge{ie.Vertex, p.Vertex, ie.Weight * p.Weight}, p, ie}, true
		})
	}
	return
}

type rEdge struct {
	Derived     Edge
	Pivot, Edge *implicitEdge
}

func (ei rEdge) Less(ej rEdge) bool {
	return ei.Derived.Less(ej.Derived)
}
