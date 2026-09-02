// Package embedding은 노트 임베딩의 직렬화와 유사도 계산을 제공한다.
package embedding

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
)

// Record 저장된 노트 임베딩.
type Record struct {
	NoteID      int
	Model       string
	Dimensions  int
	Vector      []float32
	ContentHash string
	UpdatedAt   string
}

func Encode(vector []float32) []byte {
	data := make([]byte, len(vector)*4)
	for i, value := range vector {
		binary.LittleEndian.PutUint32(data[i*4:], math.Float32bits(value))
	}
	return data
}

func Decode(data []byte) ([]float32, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("임베딩 데이터 길이가 올바르지 않습니다")
	}
	vector := make([]float32, len(data)/4)
	for i := range vector {
		vector[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[i*4:]))
	}
	return vector, nil
}

func ContentHash(title, content, category string) string {
	hash := sha256.Sum256([]byte(title + "\n" + content + "\n" + category))
	return fmt.Sprintf("%x", hash[:])
}

// CosineSimilarity 두 벡터의 코사인 유사도를 반환한다.
func CosineSimilarity(left, right []float32) (float32, error) {
	if len(left) == 0 || len(left) != len(right) {
		return 0, fmt.Errorf("임베딩 차원이 일치하지 않습니다")
	}
	var dot, leftNorm, rightNorm float64
	for i := range left {
		l, r := float64(left[i]), float64(right[i])
		dot += l * r
		leftNorm += l * l
		rightNorm += r * r
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0, nil
	}
	return float32(dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm))), nil
}
