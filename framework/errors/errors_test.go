package errors

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	err := New("something went wrong")
	require.Error(t, err)
	assert.Equal(t, "something went wrong", err.Error())
	assert.NotEmpty(t, GetStack(err))
	assert.Empty(t, GetCode(err))
	assert.NotNil(t, GetMetadata(err))
}

func TestWrap(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		assert.NoError(t, Wrap(nil, "context"))
	})

	t.Run("wraps plain error", func(t *testing.T) {
		base := stderrors.New("base failure")
		wrapped := Wrap(base, "loading user")
		require.Error(t, wrapped)
		assert.Equal(t, "loading user: base failure", wrapped.Error())
		assert.True(t, Is(wrapped, base), "wrapped error must unwrap to the original")
	})

	t.Run("wrapping a framework Error preserves code and stack", func(t *testing.T) {
		inner := WithCode(New("inner"), CodeNotFound)
		outer := Wrap(inner, "outer")
		assert.Equal(t, "outer: inner", outer.Error())
		assert.Equal(t, CodeNotFound, GetCode(outer))
		assert.Equal(t, GetStack(inner), GetStack(outer))
	})
}

func TestWithCode(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		assert.NoError(t, WithCode(nil, CodeInternal))
	})

	t.Run("attaches code to plain error", func(t *testing.T) {
		err := WithCode(stderrors.New("missing"), CodeNotFound)
		assert.Equal(t, CodeNotFound, GetCode(err))
	})

	t.Run("does not mutate the input error", func(t *testing.T) {
		original := New("shared")
		_ = WithCode(original, CodeConflict)
		assert.Empty(t, GetCode(original), "WithCode must clone, not mutate")
	})
}

func TestWithMetadata(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		assert.NoError(t, WithMetadata(nil, "k", "v"))
	})

	t.Run("attaches metadata", func(t *testing.T) {
		err := WithMetadata(New("boom"), "user_id", "u-1")
		assert.Equal(t, "u-1", GetMetadata(err)["user_id"])
	})

	t.Run("does not mutate the input error's metadata", func(t *testing.T) {
		original := WithMetadata(New("boom"), "a", 1)
		_ = WithMetadata(original, "b", 2)
		_, exists := GetMetadata(original)["b"]
		assert.False(t, exists, "WithMetadata must clone the metadata map")
	})
}

func TestShorthandConstructors(t *testing.T) {
	base := stderrors.New("cause")

	tests := []struct {
		name     string
		build    func(error, string) error
		wantCode string
	}{
		{"NewInternal", NewInternal, CodeInternal},
		{"NewNotFound", NewNotFound, CodeNotFound},
		{"NewInvalidInput", NewInvalidInput, CodeInvalidInput},
		{"NewUnauthorized", NewUnauthorized, CodeUnauthorized},
		{"NewForbidden", NewForbidden, CodeForbidden},
		{"NewConflict", NewConflict, CodeConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.build(base, "message")
			assert.Equal(t, tt.wantCode, GetCode(err))
			assert.Equal(t, "message: cause", err.Error())
			assert.True(t, Is(err, base))
		})
	}
}

func TestToHTTPCode(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		{CodeNotFound, 404},
		{CodeInvalidInput, 400},
		{CodeValidation, 400},
		{CodeUnauthorized, 401},
		{CodeForbidden, 403},
		{CodeConflict, 409},
		{CodeAlreadyExists, 409},
		{CodeTimeout, 408},
		{CodeDeadlineExceeded, 408},
		{CodeUnavailable, 503},
		{CodeNotImplemented, 501},
		{CodeInternal, 500},
		{"", 500},
	}

	for _, tt := range tests {
		name := tt.code
		if name == "" {
			name = "uncoded"
		}
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, ToHTTPCode(WithCode(New("x"), tt.code)))
		})
	}

	t.Run("plain error maps to 500", func(t *testing.T) {
		assert.Equal(t, 500, ToHTTPCode(stderrors.New("plain")))
	})
}

func TestToGRPCCode(t *testing.T) {
	tests := []struct {
		code string
		want uint32
	}{
		{CodeNotFound, 5},         // NotFound
		{CodeInvalidInput, 3},     // InvalidArgument
		{CodeValidation, 3},       // InvalidArgument
		{CodeUnauthorized, 16},    // Unauthenticated
		{CodeForbidden, 7},        // PermissionDenied
		{CodeConflict, 6},         // AlreadyExists
		{CodeAlreadyExists, 6},    // AlreadyExists
		{CodeTimeout, 4},          // DeadlineExceeded
		{CodeDeadlineExceeded, 4}, // DeadlineExceeded
		{CodeUnavailable, 14},     // Unavailable
		{CodeNotImplemented, 12},  // Unimplemented
		{CodeCanceled, 1},         // Canceled
		{CodeInternal, 13},        // Internal
		{"", 13},                  // Internal
	}

	for _, tt := range tests {
		name := tt.code
		if name == "" {
			name = "uncoded"
		}
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, ToGRPCCode(WithCode(New("x"), tt.code)))
		})
	}
}

func TestIsAndAs(t *testing.T) {
	sentinel := ErrNotFound
	wrapped := Wrap(sentinel, "outer")

	assert.True(t, Is(wrapped, sentinel))
	assert.False(t, Is(wrapped, ErrConflict))

	var target *Error
	require.True(t, As(wrapped, &target))
	assert.Equal(t, "outer: not found", target.Message)
}

func TestStackCapture(t *testing.T) {
	err := New("stack me")
	stack := GetStack(err)
	assert.Contains(t, stack, "errors_test.go", "stack must include the caller frame")
	assert.NotContains(t, stack, "runtime/", "runtime frames must be filtered")
}

func TestGettersOnNilAndPlainErrors(t *testing.T) {
	assert.Empty(t, GetCode(nil))
	assert.Nil(t, GetMetadata(nil))
	assert.Empty(t, GetStack(nil))

	plain := stderrors.New("plain")
	assert.Empty(t, GetCode(plain))
	assert.Nil(t, GetMetadata(plain))
	assert.Empty(t, GetStack(plain))
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New("test error")
	}
}

func BenchmarkWrap(b *testing.B) {
	err := stderrors.New("original error")
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
