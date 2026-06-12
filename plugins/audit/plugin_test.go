package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/platform/observability"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger(t *testing.T) *observability.Logger {
	t.Helper()
	logger, err := observability.NewLogger(&config.Config{})
	require.NoError(t, err)
	return logger
}

func TestPlugin_Lifecycle(t *testing.T) {
	p := &Plugin{}

	assert.Equal(t, "auditing", p.Name())

	err := p.Initialize(map[string]interface{}{}, newTestLogger(t), nil, nil, nil)
	assert.NoError(t, err)
	assert.NoError(t, p.Start())
	assert.NoError(t, p.Stop())
}

func TestPlugin_RecordWithoutFileSink(t *testing.T) {
	p := &Plugin{}
	require.NoError(t, p.Initialize(map[string]interface{}{}, newTestLogger(t), nil, nil, nil))
	require.NoError(t, p.Start())
	defer func() { assert.NoError(t, p.Stop()) }()

	err := p.Record(Event{Actor: "user-1", Action: "login", Resource: "session", Result: "success"})
	assert.NoError(t, err)
}

func TestPlugin_RecordToFileSink(t *testing.T) {
	auditPath := filepath.Join(t.TempDir(), "audit.jsonl")

	p := &Plugin{}
	require.NoError(t, p.Initialize(map[string]interface{}{"filePath": auditPath}, newTestLogger(t), nil, nil, nil))
	require.NoError(t, p.Start())

	events := []Event{
		{Actor: "user-1", Action: "create", Resource: "orders/42", Result: "success", TenantID: "acme"},
		{Actor: "user-2", Action: "delete", Resource: "orders/42", Result: "denied",
			Metadata: map[string]interface{}{"reason": "forbidden"}},
	}
	for _, e := range events {
		require.NoError(t, p.Record(e))
	}
	require.NoError(t, p.Stop())

	file, err := os.Open(auditPath)
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	var got []Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var e Event
		require.NoError(t, json.Unmarshal(scanner.Bytes(), &e))
		got = append(got, e)
	}
	require.NoError(t, scanner.Err())

	require.Len(t, got, 2)
	assert.Equal(t, "user-1", got[0].Actor)
	assert.Equal(t, "create", got[0].Action)
	assert.Equal(t, "acme", got[0].TenantID)
	assert.False(t, got[0].Timestamp.IsZero(), "timestamp must be filled in")
	assert.Equal(t, "denied", got[1].Result)
	assert.Equal(t, "forbidden", got[1].Metadata["reason"])
}

func TestPlugin_RecordPreservesExplicitTimestamp(t *testing.T) {
	p := &Plugin{}
	require.NoError(t, p.Initialize(map[string]interface{}{}, newTestLogger(t), nil, nil, nil))
	require.NoError(t, p.Start())
	defer func() { assert.NoError(t, p.Stop()) }()

	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	e := Event{Timestamp: ts, Actor: "user-1", Action: "read", Resource: "doc/1", Result: "success"}
	assert.NoError(t, p.Record(e))
}

func TestPlugin_StartFailsOnUnwritablePath(t *testing.T) {
	p := &Plugin{}
	badPath := filepath.Join(t.TempDir(), "missing-dir", "audit.jsonl")
	require.NoError(t, p.Initialize(map[string]interface{}{"filePath": badPath}, newTestLogger(t), nil, nil, nil))

	err := p.Start()
	assert.Error(t, err)
}
