// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package perftrace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSessionDisabledNoop(t *testing.T) {
	s, err := NewSession(Options{})
	require.NoError(t, err)
	assert.False(t, s.Enabled())
	s.Emit(Event{Type: EventSessionStart})
	require.NoError(t, s.Close(nil))
}

func TestNewSessionInvalidPath(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "not-a-dir")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0o644))

	_, err := NewSession(Options{File: filepath.Join(file, "trace.jsonl")})
	require.Error(t, err)
}

func TestEmitWritesJSONLAndSequences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	s, err := NewSession(Options{
		File:     path,
		RunID:    "run-1",
		Scenario: "startup",
	})
	require.NoError(t, err)

	s.Emit(Event{Type: EventSessionStart})
	s.Emit(Event{Type: EventConfigSnapshot})
	require.NoError(t, s.Close(nil))

	events := readEvents(t, path)
	require.Len(t, events, 3)
	assert.Equal(t, int64(1), events[0].Seq)
	assert.Equal(t, int64(2), events[1].Seq)
	assert.Equal(t, int64(3), events[2].Seq)
	assert.Equal(t, EventSessionEnd, events[2].Type)
	assert.Equal(t, "run-1", events[0].RunID)
	assert.Equal(t, "startup", events[0].Scenario)
}

func TestWriteFailureDisablesFutureTraceWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	s, err := NewSession(Options{File: path})
	require.NoError(t, err)

	require.NoError(t, s.file.Close())
	s.Emit(Event{Type: EventSessionStart})
	assert.False(t, s.Enabled())
	_ = s.Close(nil)
}

func readEvents(t *testing.T, path string) []Event {
	t.Helper()

	bb, err := os.ReadFile(path)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(bb)), "\n")
	events := make([]Event, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var ev Event
		require.NoError(t, json.Unmarshal([]byte(line), &ev))
		events = append(events, ev)
	}

	return events
}
