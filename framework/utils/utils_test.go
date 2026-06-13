package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTitleCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"single word", "order", "Order"},
		{"two words", "order item", "Order Item"},
		{"already cased", "Order", "Order"},
		{"empty", "", ""},
		{"extra spaces collapse", "  order   item ", "Order Item"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, TitleCase(tt.input))
		})
	}
}

func TestGenerateUUIDAndRandomString(t *testing.T) {
	first, second := GenerateUUID(), GenerateUUID()
	assert.Len(t, first, 36)
	assert.NotEqual(t, first, second)

	s, err := GenerateRandomString(16)
	require.NoError(t, err)
	assert.Len(t, s, 16)
}

func TestConversions(t *testing.T) {
	t.Run("int round trip", func(t *testing.T) {
		n, err := StringToInt("42")
		require.NoError(t, err)
		assert.Equal(t, 42, n)
		assert.Equal(t, "42", IntToString(42))

		_, err = StringToInt("not-a-number")
		assert.Error(t, err)
	})

	t.Run("float round trip", func(t *testing.T) {
		f, err := StringToFloat("3.5")
		require.NoError(t, err)
		assert.Equal(t, 3.5, f)
		assert.Equal(t, "3.5", FloatToString(3.5))
	})

	t.Run("bool round trip", func(t *testing.T) {
		b, err := StringToBool("true")
		require.NoError(t, err)
		assert.True(t, b)
		assert.Equal(t, "false", BoolToString(false))
	})

	t.Run("time round trip", func(t *testing.T) {
		stamp := "2026-06-12T10:00:00Z"
		parsed, err := StringToTime(stamp, time.RFC3339)
		require.NoError(t, err)
		assert.Equal(t, stamp, TimeToString(parsed, time.RFC3339))
	})
}

func TestStructMapRoundTrip(t *testing.T) {
	type payload struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	m, err := StructToMap(payload{Name: "x", Count: 2})
	require.NoError(t, err)
	assert.Equal(t, "x", m["name"])

	var out payload
	require.NoError(t, MapToStruct(m, &out))
	assert.Equal(t, payload{Name: "x", Count: 2}, out)
}

func TestValidators(t *testing.T) {
	tests := []struct {
		name string
		fn   func(string) bool
		in   string
		want bool
	}{
		{"valid email", IsEmail, "user@example.com", true},
		{"invalid email", IsEmail, "not-an-email", false},
		{"valid url", IsURL, "https://example.com/path", true},
		{"invalid url", IsURL, "example", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.fn(tt.in))
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want bool
	}{
		{"nil", nil, true},
		{"empty string", "", true},
		{"string", "x", false},
		{"zero int", 0, true},
		{"int", 1, false},
		{"empty slice", []string{}, true},
		{"slice", []string{"a"}, false},
		{"false bool", false, true},
		{"nil pointer", (*int)(nil), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsEmpty(tt.in))
		})
	}
}

func TestStringHelpers(t *testing.T) {
	assert.Equal(t, "abc", Truncate("abcdef", 3))
	assert.Equal(t, "abc", Truncate("abc", 10))
	assert.True(t, Contains("hello world", "world"))
	assert.Equal(t, "a,b", Join([]string{"a", "b"}, ","))
	assert.Equal(t, []string{"a", "b"}, Split("a,b", ","))
	assert.Equal(t, "x", TrimSpace("  x  "))
}

func TestFormatters(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"bytes", FormatBytes(512), "512 B"},
		{"kilobytes", FormatBytes(2048), "2.0 KB"},
		{"megabytes", FormatBytes(5 * 1024 * 1024), "5.0 MB"},
		{"millis", FormatDuration(250 * time.Millisecond), "250 ms"},
		{"seconds", FormatDuration(2500 * time.Millisecond), "2.5 s"},
		{"minutes", FormatDuration(90 * time.Second), "1.5 m"},
		{"hours", FormatDuration(90 * time.Minute), "1.5 h"},
		{"days", FormatDuration(36 * time.Hour), "1.5 d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

func TestRetry(t *testing.T) {
	t.Run("succeeds after failures", func(t *testing.T) {
		calls := 0
		err := Retry(3, time.Millisecond, func() error {
			calls++
			if calls < 3 {
				return assert.AnError
			}
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 3, calls)
	})

	t.Run("returns last error when exhausted", func(t *testing.T) {
		err := Retry(2, time.Millisecond, func() error { return assert.AnError })
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestPointerHelpers(t *testing.T) {
	p := Pointer(42)
	require.NotNil(t, p)
	assert.Equal(t, 42, *p)
	assert.Equal(t, 42, Dereference(p, 0))
	assert.Equal(t, 7, Dereference[int](nil, 7))
}
