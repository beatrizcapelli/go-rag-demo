package rag

import "testing"

func TestInMemoryStore_AddAndSearch(t *testing.T) {
	store := NewInMemoryStore()

	// 2D toy embeddings so we can reason easily
	chunks := []Chunk{
		{ID: "1", Content: "A", Embedding: []float64{1, 0}},
		{ID: "2", Content: "B", Embedding: []float64{0, 1}},
	}
	store.Add(chunks...)

	// query close to {1,0}
	query := []float64{0.9, 0.1}
	results := store.Search(query, 1)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Chunk.ID != "1" {
		t.Fatalf("expected best match to be chunk 1, got %s", results[0].Chunk.ID)
	}
}

func TestCosineSimilarity_Basic(t *testing.T) {
	a := []float64{1, 0}
	b := []float64{1, 0}
	c := []float64{0, 1}

	if got := cosine(a, b); got < 0.99 {
		t.Fatalf("expected cosine(a,b) ~ 1, got %f", got)
	}

	if got := cosine(a, c); got > 0.01 {
		t.Fatalf("expected cosine(a,c) ~ 0, got %f", got)
	}
}

func TestInMemoryStore_SearchTopKBounds(t *testing.T) {
	store := NewInMemoryStore()
	store.Add(
		Chunk{ID: "1", Embedding: []float64{1, 0}},
		Chunk{ID: "2", Embedding: []float64{0, 1}},
	)

	res := store.Search([]float64{1, 0}, 10)
	if len(res) != 2 {
		t.Fatalf("expected 2 results when topK > len(chunks), got %d", len(res))
	}
}
