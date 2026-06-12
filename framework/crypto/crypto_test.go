package crypto

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashSHA256(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "hello",
			input:    "hello",
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, HashSHA256(tt.input))
		})
	}
}

func TestHashSHA512(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
		},
		{
			name:     "hello",
			input:    "hello",
			expected: "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, HashSHA512(tt.input))
		})
	}
}

func TestHashDeterminism(t *testing.T) {
	assert.Equal(t, HashSHA256("input"), HashSHA256("input"))
	assert.NotEqual(t, HashSHA256("input"), HashSHA256("other"))
	assert.Equal(t, HashSHA512("input"), HashSHA512("input"))
	assert.NotEqual(t, HashSHA512("input"), HashSHA512("other"))
}

func TestEncryptDecryptAES_Roundtrip(t *testing.T) {
	tests := []struct {
		name      string
		keySize   int
		plaintext []byte
	}{
		{"AES-128 short text", 16, []byte("hello world")},
		{"AES-192 short text", 24, []byte("hello world")},
		{"AES-256 short text", 32, []byte("hello world")},
		{"AES-256 single byte", 32, []byte{0x00}},
		{"AES-256 binary data", 32, []byte{0xff, 0x00, 0xaa, 0x55, 0x01}},
		{"AES-256 large payload", 32, make([]byte, 64*1024)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GenerateRandomBytes(tt.keySize)
			require.NoError(t, err)

			encrypted, err := EncryptAES(tt.plaintext, key)
			require.NoError(t, err)
			require.NotEmpty(t, encrypted)

			decrypted, err := DecryptAES(encrypted, key)
			require.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestEncryptAES_NonceUniqueness(t *testing.T) {
	key, err := GenerateRandomBytes(32)
	require.NoError(t, err)

	first, err := EncryptAES([]byte("same plaintext"), key)
	require.NoError(t, err)
	second, err := EncryptAES([]byte("same plaintext"), key)
	require.NoError(t, err)

	// A random nonce must yield different ciphertexts for the same input.
	assert.NotEqual(t, first, second)
}

func TestEncryptAES_InvalidInputs(t *testing.T) {
	validKey := make([]byte, 32)

	tests := []struct {
		name      string
		plaintext []byte
		key       []byte
		expected  error
	}{
		{"nil key", []byte("data"), nil, ErrInvalidKey},
		{"empty key", []byte("data"), []byte{}, ErrInvalidKey},
		{"key too short", []byte("data"), make([]byte, 15), ErrInvalidKey},
		{"key invalid middle size", []byte("data"), make([]byte, 20), ErrInvalidKey},
		{"key too long", []byte("data"), make([]byte, 33), ErrInvalidKey},
		{"empty plaintext", []byte{}, validKey, ErrInvalidData},
		{"nil plaintext", nil, validKey, ErrInvalidData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := EncryptAES(tt.plaintext, tt.key)
			assert.ErrorIs(t, err, tt.expected)
			assert.Empty(t, encrypted)
		})
	}
}

func TestDecryptAES_WrongKey(t *testing.T) {
	key, err := GenerateRandomBytes(32)
	require.NoError(t, err)
	otherKey, err := GenerateRandomBytes(32)
	require.NoError(t, err)

	encrypted, err := EncryptAES([]byte("secret"), key)
	require.NoError(t, err)

	decrypted, err := DecryptAES(encrypted, otherKey)
	assert.ErrorIs(t, err, ErrDecryptionFailed)
	assert.Nil(t, decrypted)
}

func TestDecryptAES_TamperedCiphertext(t *testing.T) {
	key, err := GenerateRandomBytes(32)
	require.NoError(t, err)

	encrypted, err := EncryptAES([]byte("secret"), key)
	require.NoError(t, err)

	raw, err := base64.StdEncoding.DecodeString(encrypted)
	require.NoError(t, err)

	// Flip a bit in the last byte (part of the GCM auth tag).
	raw[len(raw)-1] ^= 0x01
	tampered := base64.StdEncoding.EncodeToString(raw)

	decrypted, err := DecryptAES(tampered, key)
	assert.ErrorIs(t, err, ErrDecryptionFailed)
	assert.Nil(t, decrypted)
}

func TestDecryptAES_InvalidInputs(t *testing.T) {
	validKey := make([]byte, 32)
	shortCiphertext := base64.StdEncoding.EncodeToString([]byte{0x01, 0x02})

	tests := []struct {
		name      string
		encrypted string
		key       []byte
		expected  error
	}{
		{"invalid key length", "irrelevant", make([]byte, 5), ErrInvalidKey},
		{"not base64", "not-valid-base64!!!", validKey, ErrInvalidData},
		{"ciphertext shorter than nonce", shortCiphertext, validKey, ErrInvalidData},
		{"empty ciphertext", "", validKey, ErrInvalidData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted, err := DecryptAES(tt.encrypted, tt.key)
			assert.ErrorIs(t, err, tt.expected)
			assert.Nil(t, decrypted)
		})
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"zero length", 0},
		{"sixteen bytes", 16},
		{"thirty-two bytes", 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes, err := GenerateRandomBytes(tt.length)
			require.NoError(t, err)
			assert.Len(t, bytes, tt.length)
		})
	}

	t.Run("uniqueness", func(t *testing.T) {
		first, err := GenerateRandomBytes(32)
		require.NoError(t, err)
		second, err := GenerateRandomBytes(32)
		require.NoError(t, err)
		assert.NotEqual(t, first, second)
	})
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"zero length", 0},
		{"short string", 8},
		{"long string", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, err := GenerateRandomString(tt.length)
			require.NoError(t, err)
			assert.Len(t, str, tt.length)
		})
	}

	t.Run("uniqueness", func(t *testing.T) {
		first, err := GenerateRandomString(32)
		require.NoError(t, err)
		second, err := GenerateRandomString(32)
		require.NoError(t, err)
		assert.NotEqual(t, first, second)
	})
}
