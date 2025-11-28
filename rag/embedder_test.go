package rag

import "testing"

func TestSimpleEmbedder_Deterministic(t *testing.T) {
	e := NewSimpleEmbedder()

	text := "Go is great for AI."
	v1 := e.Embed(text)
	v2 := e.Embed(text)

	if len(v1) == 0 {
		t.Fatalf("expected non-empty embedding")
	}
	if len(v1) != len(v2) {
		t.Fatalf("embeddings length mismatch: %d vs %d", len(v1), len(v2))
	}
	for i := range v1 {
		if v1[i] != v2[i] {
			t.Fatalf("embeddings not deterministic at index %d: %v vs %v", i, v1[i], v2[i])
		}
	}
}

func TestSimpleEmbedder_DifferentTextsDiffer(t *testing.T) {
	e := NewSimpleEmbedder()

	v1 := e.Embed("short")
	v2 := e.Embed("a much longer string")

	if len(v1) != len(v2) {
		t.Fatalf("expected same dimension, got %d vs %d", len(v1), len(v2))
	}

	// they should differ in at least one dimension
	different := false
	for i := range v1 {
		if v1[i] != v2[i] {
			different = true
			break
		}
	}
	if !different {
		t.Fatalf("expected different embeddings for different texts")
	}
}
