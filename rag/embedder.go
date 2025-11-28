package rag

// Embedder is an interface so later you can swap implementation
type Embedder interface {
	Embed(text string) []float64
}

// Simple deterministic fake embedder based on rune counts
type SimpleEmbedder struct{}

func NewSimpleEmbedder() *SimpleEmbedder {
	return &SimpleEmbedder{}
}

func (e *SimpleEmbedder) Embed(text string) []float64 {
	// fake 4D vector: length, vowels, consonants, spaces
	var length, vowels, consonants, spaces float64
	for _, r := range text {
		length++
		switch {
		case r == 'a' || r == 'e' || r == 'i' || r == 'o' || r == 'u' ||
			r == 'A' || r == 'E' || r == 'I' || r == 'O' || r == 'U':
			vowels++
		case r == ' ':
			spaces++
		default:
			consonants++
		}
	}
	return []float64{length, vowels, consonants, spaces}
}
