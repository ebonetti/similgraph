package similgraph

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	g, new2Old, err := New(sampleEdges(), 0, 0, 0)

	gTest := SimilGraph{[]implicitEdge{{2, 0.6}, {3, 0.8}, {3, 0.8}, {4, 0.6}, {0, 0.6}, {0, 0.8}, {1, 0.8}, {1, 0.6}}, []uint32{0, 2, 4, 5, 7, 8}, 2}
	new2OldTest := []uint32{0, 1}

	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*g, gTest) {
		t.Error("G is not built in the correct way")
	}

	if !reflect.DeepEqual(new2Old, new2OldTest) {
		t.Error("G is not indexed in the correct way")
	}
}

func TestEdgeIterator(t *testing.T) {
	g, _, _ := New(sampleEdges(), 0, 0, 0)
	smallerIt, biggerIt, err := g.EdgeIterator(0)

	if err != nil {
		t.Error(err)
	}

	if e, ok := smallerIt(); ok {
		t.Errorf("smallerIterator should be empty, but it contains %v", e)
	}

	e, ok := biggerIt()
	eTest := Edge{0, 1, 0.64000005}
	e1, ok1 := biggerIt()
	switch {
	case !ok:
		t.Error("biggerIterator should contain the edge between 0 and 1")
	case e != eTest:
		t.Errorf("biggerIterator should be %v, but it's %v", eTest, e)
	case ok1:
		t.Errorf("biggerIterator should be empty after one iteration, but it contains %v", e1)
	}
}

func TestVertexCount(t *testing.T) {
	g, _, _ := New(sampleEdges(), 0, 0, 0)
	if g.VertexCount() != g.vertexCount {
		t.Error("G.VertexCount() doesn't return the correct edge count")
	}
}

func sampleEdges() func() (Edge, bool) {
	edges := []Edge{{0, 0, 3}, {0, 1, 4}, {1, 1, 4}, {1, 2, 3}}
	return func() (e Edge, ok bool) {
		if len(edges) == 0 {
			return
		}

		e, edges, ok = edges[0], edges[1:], true
		return
	}
}
