// Package similgraph provides primitives for computing a cosine similarity graph whose edges are computed on the fly.
package similgraph

import (
	"sort"

	"github.com/pkg/errors"
)

//go:generate gorewrite

//SimilGraph represents a cosine similarity graph whose edges are computed on the fly
type SimilGraph struct {
	bigraphEdges []implicitEdge
	vertexSlices []uint32
	vertexCount  uint32
}

//VertexCount return the count of the vertex in the graph
func (g SimilGraph) VertexCount() uint32 {
	return g.vertexCount
}

//AreCorrelated return the cosine similarity graph whose edges are computed on the fly
/*func (g SimilGraph) AreCorrelated(v1, v2 uint32) (weight float32, err error) {
	if v1 >= g.vertexCount {
		return 0, errors.Errorf("similgraph: v1 (%v) is an invalid vertex, vertex limit is %v.", v1, g.vertexCount)
	}
	if v2 >= g.vertexCount {
		return 0, errors.Errorf("similgraph: v2 (%v) is an invalid vertex, vertex limit is %v.", v2, g.vertexCount)
	}
	if v1 == v2 {
		return 0, errors.Errorf("similgraph: v1 and v2 are the same vertex (%v).", v1)
	}

	edgegroup1 := g.bigraphEdges[g.vertexSlices[v1]:g.vertexSlices[v1+1]]
	edgegroup2 := g.bigraphEdges[g.vertexSlices[v2]:g.vertexSlices[v2+1]]
	f64weight := float64(0)
	for {
		edge1 := edgegroup1[0]
		edge2 := edgegroup2[0]
		switch {
		case edge1.Vertex < edge2.Vertex:
			if len(edgegroup1) > 1 {
				edgegroup1 = edgegroup1[1:]
			} else {
				return float32(f64weight), nil
			}
		case edge1.Vertex > edge2.Vertex:
			if len(edgegroup2) > 1 {
				edgegroup2 = edgegroup2[1:]
			} else {
				return float32(f64weight), nil
			}
		default:
			f64weight += float64(edge1.Weight) * float64(edge2.Weight)
			if len(edgegroup1) > 1 && len(edgegroup2) > 1 {
				edgegroup1 = edgegroup1[1:]
				edgegroup2 = edgegroup2[1:]
			} else {
				return float32(f64weight), nil
			}
		}
	}
}*/

//EdgeIterator iterate over the edges of the cosine similarity graph, computed on the fly.
func (g SimilGraph) EdgeIterator(v uint32) (smallerIterator, biggerIterator func() (info Edge, ok bool), err error) {
	if v >= g.vertexCount {
		return nil, nil, errors.Errorf("similgraph: v (%v) is an invalid vertex, vertex limit is %v.", v, g.vertexCount)
	}

	smallerh := make([]pivotedEdgesSpan, 0, 4)
	biggerh := make([]pivotedEdgesSpan, 0, 4)
	for _, e := range g.bigraphEdges[g.vertexSlices[v]:g.vertexSlices[v+1]] {
		edgegroup := g.bigraphEdges[g.vertexSlices[e.Vertex]:g.vertexSlices[e.Vertex+1]]
		index := sort.Search(len(edgegroup), func(i int) bool { return edgegroup[i].Vertex >= v })
		if index >= len(edgegroup) || edgegroup[index].Vertex != v {
			return nil, nil, errors.Errorf("similgraph internal error: vertex (%v) was not found where it should be.", v)
		}
		if index > 0 {
			smallerh = append(smallerh, pivotedEdgesSpan{edgegroup[index], edgegroup[:index]})
		}
		if index+1 < len(edgegroup) {
			biggerh = append(biggerh, pivotedEdgesSpan{edgegroup[index], edgegroup[index+1:]})
		}
	}
	smallerIterator = edgeIterMergeFrom(pES2EdgeIter(smallerh...)...).NextSum
	biggerIterator = edgeIterMergeFrom(pES2EdgeIter(biggerh...)...).NextSum
	return smallerIterator, biggerIterator, nil
}

//Edge represent a weighted edge
type Edge struct {
	VertexA uint32
	VertexB uint32
	Weight  float32
}

//Less implement the Lexicographical order between two edges.
func (ei Edge) Less(ej Edge) bool {
	if ei.VertexA < ej.VertexA {
		return true
	}
	if ei.VertexA == ej.VertexA && ei.VertexB < ej.VertexB {
		return true
	}
	return false
}

type implicitEdge struct {
	Vertex uint32
	Weight float32
}

type pivotedEdgesSpan struct {
	Pivot implicitEdge
	Edges []implicitEdge
}

func pES2EdgeIter(pes ...pivotedEdgesSpan) (nexts []func() (Edge, bool)) {
	for _, s := range pes {
		s := s
		nexts = append(nexts, func() (e Edge, ok bool) {
			if len(s.Edges) == 0 {
				return
			}
			p, ie := s.Pivot, s.Edges[0]
			s.Edges = s.Edges[1:]
			return Edge{p.Vertex, ie.Vertex, p.Weight * ie.Weight}, true
		})
	}
	return
}

func (m *edgeIterMerge) NextSum() (e Edge, ok bool) {
	e, ok = m.Next()
	if !ok {
		return
	}

	f64eWeight := float64(e.Weight)
	for e1, ok := m.Peek(); ok && equalIndexes(e, e1); e1, ok = m.Peek() {
		e, _ := m.Next()
		f64eWeight += float64(e.Weight)
	}
	e.Weight = float32(f64eWeight)
	return
}

func equalIndexes(a, b Edge) bool {
	return a.VertexA == b.VertexA && a.VertexB == b.VertexB
}
