package embedding

import "testing"

func TestVectorRoundTrip(t *testing.T) {
	want := []float32{0.25, -1.5, 3}
	got, err := Decode(Encode(want))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("vector[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestCosineSimilarity(t *testing.T) {
	got, err := CosineSimilarity([]float32{1, 0}, []float32{1, 0})
	if err != nil || got != 1 {
		t.Fatalf("same vector similarity = %v, error = %v", got, err)
	}
	got, err = CosineSimilarity([]float32{1, 0}, []float32{0, 1})
	if err != nil || got != 0 {
		t.Fatalf("orthogonal vector similarity = %v, error = %v", got, err)
	}
	if _, err := CosineSimilarity([]float32{1}, []float32{1, 2}); err == nil {
		t.Fatal("dimension mismatch was accepted")
	}
}

func TestContentHashChangesWithContent(t *testing.T) {
	if ContentHash("title", "one", "work") == ContentHash("title", "two", "work") {
		t.Fatal("content changes did not change hash")
	}
}
