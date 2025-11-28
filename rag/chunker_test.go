package rag

import "testing"

type fakeEmbedder struct{}

func (f *fakeEmbedder) Embed(text string) []float64 {
	return []float64{1} // dummy
}

func TestChunkText_SplitsSentences(t *testing.T) {
	text := "Sentence one. Sentence two. Sentence three. Sentence four."

	e := &fakeEmbedder{}
	chunks := ChunkText(text, "test-doc", e)

	if len(chunks) == 0 {
		t.Fatalf("expected at least one chunk")
	}

	// With max 3 sentences per chunk, this should be 2 chunks
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}

	if chunks[0].Content == "" || chunks[1].Content == "" {
		t.Fatalf("expected chunk content to be non-empty")
	}

	if chunks[0].Source != "test-doc" {
		t.Fatalf("expected source to be 'test-doc', got %s", chunks[0].Source)
	}
}

func TestChunkText_EmptyInput(t *testing.T) {
	e := &fakeEmbedder{}
	chunks := ChunkText("", "empty", e)

	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks for empty input, got %d", len(chunks))
	}
}
