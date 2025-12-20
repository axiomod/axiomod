package errors

import (
	"errors"
	"testing"
)

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New("test error")
	}
}

func BenchmarkWrap(b *testing.B) {
	err := errors.New("original error")
	for i := 0; i < b.N; i++ {
		_ = Wrap(err, "wrapped error")
	}
}

func BenchmarkWithCode(b *testing.B) {
	err := New("test error")
	for i := 0; i < b.N; i++ {
		_ = WithCode(err, CodeInternal)
	}
}
