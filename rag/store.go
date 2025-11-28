package rag

import (
	"math"
	"sync"
)

type InMemoryStore struct {
	mu     sync.RWMutex
	chunks []Chunk
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		chunks: []Chunk{},
	}
}

func (s *InMemoryStore) Add(chunks ...Chunk) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chunks = append(s.chunks, chunks...)
}

// naive cosine similarity
func cosine(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func (s *InMemoryStore) Search(queryEmbedding []float64, topK int) []SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]SearchResult, 0, len(s.chunks))
	for _, ch := range s.chunks {
		score := cosine(queryEmbedding, ch.Embedding)
		results = append(results, SearchResult{
			Chunk: ch,
			Score: score,
		})
	}

	// sort by score desc (simple bubble-ish, OK for small demo)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if topK > len(results) {
		topK = len(results)
	}
	return results[:topK]
}

func (s *InMemoryStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chunks = nil
}