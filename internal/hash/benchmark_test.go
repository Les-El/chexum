package hash

import (
	"testing"
)

func BenchmarkComputeBytesSHA256_1KB(b *testing.B) {
	computer, _ := NewComputer(AlgorithmSHA256)
	data := make([]byte, 1024) // 1KB
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		computer.ComputeBytes(data)
	}
}

func BenchmarkComputeBytesSHA256_1MB(b *testing.B) {
	computer, _ := NewComputer(AlgorithmSHA256)
	data := make([]byte, 1024*1024) // 1MB
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		computer.ComputeBytes(data)
	}
}

func BenchmarkComputeBytesMD5_1KB(b *testing.B) {
	computer, _ := NewComputer(AlgorithmMD5)
	data := make([]byte, 1024) // 1KB
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		computer.ComputeBytes(data)
	}
}
