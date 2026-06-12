package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryCache(t *testing.T) {
	tests := []struct {
		name          string
		maxItems      int
		janitorWanted bool
	}{
		{"bounded cache starts janitor", 10, true},
		{"unbounded cache has no janitor", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewMemoryCache(tt.maxItems)
			require.NotNil(t, c)
			assert.Equal(t, tt.janitorWanted, c.janitorOn)
		})
	}
}

func TestMemoryCache_SetGet(t *testing.T) {
	c := NewMemoryCache(0)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "key", []byte("value"), 0))

	got, err := c.Get(ctx, "key")
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), got)
}

func TestMemoryCache_GetMissingKey(t *testing.T) {
	c := NewMemoryCache(0)

	got, err := c.Get(context.Background(), "missing")
	assert.ErrorIs(t, err, ErrKeyNotFound)
	assert.Nil(t, got)
}

func TestMemoryCache_Overwrite(t *testing.T) {
	c := NewMemoryCache(0)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "key", []byte("first"), 0))
	require.NoError(t, c.Set(ctx, "key", []byte("second"), 0))

	got, err := c.Get(ctx, "key")
	require.NoError(t, err)
	assert.Equal(t, []byte("second"), got)
}

func TestMemoryCache_TTLExpiry(t *testing.T) {
	c := NewMemoryCache(0)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "expiring", []byte("value"), 100*time.Millisecond))

	// Available before expiry.
	got, err := c.Get(ctx, "expiring")
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), got)

	// Generous margin to avoid timing flakiness.
	time.Sleep(400 * time.Millisecond)

	got, err = c.Get(ctx, "expiring")
	assert.ErrorIs(t, err, ErrKeyNotFound)
	assert.Nil(t, got)

	// The expired entry must have been removed lazily on Get.
	c.mu.RLock()
	_, stillThere := c.items["expiring"]
	c.mu.RUnlock()
	assert.False(t, stillThere, "expired item should be deleted on access")
}

func TestMemoryCache_ZeroTTLNeverExpires(t *testing.T) {
	c := NewMemoryCache(0)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "forever", []byte("value"), 0))
	time.Sleep(150 * time.Millisecond)

	got, err := c.Get(ctx, "forever")
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), got)
}

func TestMemoryCache_FullCache(t *testing.T) {
	c := NewMemoryCache(2)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "a", []byte("1"), 0))
	require.NoError(t, c.Set(ctx, "b", []byte("2"), 0))

	// A new key does not fit.
	err := c.Set(ctx, "c", []byte("3"), 0)
	assert.ErrorIs(t, err, ErrCacheFull)

	// Overwriting an existing key is still allowed at capacity.
	require.NoError(t, c.Set(ctx, "a", []byte("updated"), 0))
	got, err := c.Get(ctx, "a")
	require.NoError(t, err)
	assert.Equal(t, []byte("updated"), got)

	// Deleting frees a slot for a new key.
	require.NoError(t, c.Delete(ctx, "b"))
	require.NoError(t, c.Set(ctx, "c", []byte("3"), 0))
}

func TestMemoryCache_Delete(t *testing.T) {
	c := NewMemoryCache(0)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "key", []byte("value"), 0))
	require.NoError(t, c.Delete(ctx, "key"))

	_, err := c.Get(ctx, "key")
	assert.ErrorIs(t, err, ErrKeyNotFound)

	// Deleting a missing key is a no-op.
	assert.NoError(t, c.Delete(ctx, "missing"))
}

func TestMemoryCache_Clear(t *testing.T) {
	c := NewMemoryCache(0)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "a", []byte("1"), 0))
	require.NoError(t, c.Set(ctx, "b", []byte("2"), 0))

	require.NoError(t, c.Clear(ctx))

	_, err := c.Get(ctx, "a")
	assert.ErrorIs(t, err, ErrKeyNotFound)
	_, err = c.Get(ctx, "b")
	assert.ErrorIs(t, err, ErrKeyNotFound)
}

func TestMemoryCache_DefensiveCopies(t *testing.T) {
	c := NewMemoryCache(0)
	ctx := context.Background()

	t.Run("mutating the input after Set does not affect the cache", func(t *testing.T) {
		input := []byte("original")
		require.NoError(t, c.Set(ctx, "in", input, 0))

		input[0] = 'X'

		got, err := c.Get(ctx, "in")
		require.NoError(t, err)
		assert.Equal(t, []byte("original"), got)
	})

	t.Run("mutating the output of Get does not affect the cache", func(t *testing.T) {
		require.NoError(t, c.Set(ctx, "out", []byte("original"), 0))

		first, err := c.Get(ctx, "out")
		require.NoError(t, err)
		first[0] = 'X'

		second, err := c.Get(ctx, "out")
		require.NoError(t, err)
		assert.Equal(t, []byte("original"), second)
	})
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	c := NewMemoryCache(0)
	ctx := context.Background()

	const goroutines = 8
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := fmt.Sprintf("key-%d-%d", id, i)
				assert.NoError(t, c.Set(ctx, key, []byte("value"), time.Minute))

				got, err := c.Get(ctx, key)
				assert.NoError(t, err)
				assert.Equal(t, []byte("value"), got)

				if i%3 == 0 {
					assert.NoError(t, c.Delete(ctx, key))
				}
			}
		}(g)
	}
	wg.Wait()
}

func TestMemoryCache_ImplementsCacheInterface(t *testing.T) {
	var _ Cache = NewMemoryCache(0)
}
