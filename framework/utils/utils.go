package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GenerateUUID generates a new UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateRandomString generates a random string of the specified length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// StringToInt converts a string to an int
func StringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// IntToString converts an int to a string
func IntToString(i int) string {
	return strconv.Itoa(i)
}

// StringToFloat converts a string to a float64
func StringToFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// FloatToString converts a float64 to a string
func FloatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// StringToBool converts a string to a bool
func StringToBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

// BoolToString converts a bool to a string
func BoolToString(b bool) string {
	return strconv.FormatBool(b)
}

// StringToTime converts a string to a time.Time
func StringToTime(s string, layout string) (time.Time, error) {
	return time.Parse(layout, s)
}

// TimeToString converts a time.Time to a string
func TimeToString(t time.Time, layout string) string {
	return t.Format(layout)
}

// StructToMap converts a struct to a map
func StructToMap(obj interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// MapToStruct converts a map to a struct
func MapToStruct(m map[string]interface{}, obj interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, obj)
}

// IsEmail checks if a string is a valid email
func IsEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// IsURL checks if a string is a valid URL
func IsURL(url string) bool {
	re := regexp.MustCompile(`^(http|https)://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(:[0-9]+)?(/.*)?$`)
	return re.MatchString(url)
}

// IsEmpty checks if a value is empty
func IsEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	return false
}

// Truncate truncates a string to the specified length
func Truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length]
}

// Contains checks if a string contains a substring
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Join joins strings with a separator
func Join(elems []string, sep string) string {
	return strings.Join(elems, sep)
}

// Split splits a string by a separator
func Split(s, sep string) []string {
	return strings.Split(s, sep)
}

// TrimSpace trims whitespace from a string
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// FormatBytes formats bytes to a human-readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration formats a duration to a human-readable string
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%d ms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1f s", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1f m", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1f h", d.Hours())
	}
	return fmt.Sprintf("%.1f d", d.Hours()/24)
}

// Retry retries a function until it succeeds or reaches the maximum number of retries
func Retry(attempts int, sleep time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(sleep)
		sleep *= 2 // Exponential backoff
	}
	return err
}

// Pointer returns a pointer to the value
func Pointer[T any](v T) *T {
	return &v
}

// Dereference returns the value of a pointer or a default value if the pointer is nil
func Dereference[T any](v *T, defaultValue T) T {
	if v == nil {
		return defaultValue
	}
	return *v
}
