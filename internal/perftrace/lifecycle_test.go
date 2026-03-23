// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package perftrace

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLifecycleMarkersAreViewScopedAndOneShot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	s, err := NewSession(Options{File: path})
	require.NoError(t, err)

	seq := s.ActivateView("Pod", "v1/pods", "big", "big/pod-a")
	require.NotZero(t, seq)
	assert.True(t, s.MarkViewOnce(seq, MarkerFirstModelBuilt, Event{RowsTotal: 100}))
	assert.False(t, s.MarkViewOnce(seq, MarkerFirstModelBuilt, Event{RowsTotal: 200}))
	assert.True(t, s.MarkViewOnce(seq, MarkerFirstUsefulRow, Event{
		RowsTotal:    100,
		RowsVisible:  10,
		SelectedPath: "big/pod-a",
	}))
	assert.True(t, s.MarkFirstKeyAfterRender(seq, "Ctrl-R"))
	assert.False(t, s.MarkFirstKeyAfterRender(seq, "Ctrl-R"))
	require.NoError(t, s.Close(nil))

	events := readEvents(t, path)
	require.Len(t, events, 5)
	assert.Equal(t, EventLifecycleMark, events[0].Type)
	assert.Equal(t, MarkerViewActivate, events[0].Marker)
	assert.Equal(t, seq, events[0].ViewSeq)
	assert.Equal(t, MarkerFirstModelBuilt, events[1].Marker)
	assert.Equal(t, MarkerFirstUsefulRow, events[2].Marker)
	assert.Equal(t, MarkerFirstKeyAfterRender, events[3].Marker)
	assert.Equal(t, EventSessionEnd, events[4].Type)
}

func TestFirstKeyAfterRenderRequiresUsefulRow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	s, err := NewSession(Options{File: path})
	require.NoError(t, err)

	seq := s.ActivateView("Pod", "v1/pods", "big", "big/pod-a")
	assert.False(t, s.MarkFirstKeyAfterRender(seq, "Ctrl-R"))
	assert.True(t, s.MarkViewOnce(seq, MarkerFirstUsefulRow, Event{SelectedPath: "big/pod-a", RowsVisible: 1}))
	assert.True(t, s.MarkFirstKeyAfterRender(seq, "Ctrl-R"))
	require.NoError(t, s.Close(nil))
}
