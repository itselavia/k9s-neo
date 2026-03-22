// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package perftrace

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestNonStreamingRequestEmitsComplete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	s, err := NewSession(Options{File: path})
	require.NoError(t, err)

	body := `{"kind":"PodList","items":[{},{}]}`
	rt := s.WrapTransport("core")(roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			ContentLength: int64(len(body)),
			Header:        http.Header{"Content-Type": []string{"application/json"}},
			Body:          io.NopCloser(strings.NewReader(body)),
		}, nil
	}))

	req, err := http.NewRequest(http.MethodGet, "https://example.com/api/v1/namespaces/big/pods", nil)
	require.NoError(t, err)
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.NoError(t, s.Close(nil))

	events := readEvents(t, path)
	require.Len(t, events, 2)
	ev := events[0]
	assert.Equal(t, EventKubeRequestComplete, ev.Type)
	assert.Equal(t, int64(len(body)), ev.ResponseBytes)
	assert.Equal(t, 2, ev.ItemCount)
	assert.Equal(t, "PodList", ev.ResponseKind)
	assert.Equal(t, "big", ev.Namespace)
	assert.Equal(t, "pods", ev.Resource)
	assert.Equal(t, "list", ev.KubeVerb)
}

func TestStreamingRequestEmitsOpenAndClose(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	s, err := NewSession(Options{File: path})
	require.NoError(t, err)

	body := "hello"
	rt := s.WrapTransport("logs")(roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			ContentLength: int64(len(body)),
			Header:        http.Header{"Content-Type": []string{"text/plain"}},
			Body:          io.NopCloser(strings.NewReader(body)),
		}, nil
	}))

	req, err := http.NewRequest(http.MethodGet, "https://example.com/api/v1/namespaces/big/pods/pod-a/log?follow=true", nil)
	require.NoError(t, err)
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.NoError(t, s.Close(nil))

	events := readEvents(t, path)
	require.Len(t, events, 3)
	assert.Equal(t, EventKubeStreamOpen, events[0].Type)
	assert.Equal(t, EventKubeStreamClose, events[1].Type)
	assert.Equal(t, int64(len(body)), events[1].ResponseBytes)
	assert.True(t, events[1].Streaming)
}

func TestOversizedJSONBodySkipsInspection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	s, err := NewSession(Options{File: path})
	require.NoError(t, err)

	body := `{"kind":"PodList","items":["` + strings.Repeat("a", maxInspectBytes+10) + `"]}`
	rt := s.WrapTransport("core")(roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			ContentLength: int64(len(body)),
			Header:        http.Header{"Content-Type": []string{"application/json"}},
			Body:          io.NopCloser(strings.NewReader(body)),
		}, nil
	}))

	req, err := http.NewRequest(http.MethodGet, "https://example.com/api/v1/pods", nil)
	require.NoError(t, err)
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.NoError(t, s.Close(nil))

	events := readEvents(t, path)
	require.Len(t, events, 2)
	ev := events[0]
	assert.Equal(t, int64(len(body)), ev.ResponseBytes)
	assert.False(t, ev.BodyInspected)
	assert.Empty(t, ev.ResponseKind)
	assert.Zero(t, ev.ItemCount)
}
