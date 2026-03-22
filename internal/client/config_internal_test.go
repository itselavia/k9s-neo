// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/derailed/k9s/internal/perftrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	restclient "k8s.io/client-go/rest"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestComposeWrapTransportPreservesExistingOrder(t *testing.T) {
	order := make([]string, 0, 3)
	cfg := &restclient.Config{
		WrapTransport: func(rt http.RoundTripper) http.RoundTripper {
			return roundTripFunc(func(req *http.Request) (*http.Response, error) {
				order = append(order, "old")
				return rt.RoundTrip(req)
			})
		},
	}

	composeWrapTransport(cfg, func(rt http.RoundTripper) http.RoundTripper {
		return roundTripFunc(func(req *http.Request) (*http.Response, error) {
			order = append(order, "new")
			return rt.RoundTrip(req)
		})
	})

	base := roundTripFunc(func(*http.Request) (*http.Response, error) {
		order = append(order, "base")
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	})

	rt := cfg.WrapTransport(base)
	_, err := rt.RoundTrip((&http.Request{}))
	assert.NoError(t, err)
	assert.Equal(t, []string{"old", "new", "base"}, order)
}

func TestRESTConfigForAddsPerfWrapTransport(t *testing.T) {
	flags := genericclioptions.NewConfigFlags(false)
	kubeConfig := "./testdata/config"
	flags.KubeConfig = &kubeConfig

	trace, err := perftrace.NewSession(perftrace.Options{File: t.TempDir() + "/trace.jsonl"})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, trace.Close(nil))
	}()

	cfg := NewConfig(flags)
	cfg.SetPerfTrace(trace)

	rc, err := cfg.RESTConfigFor("core")
	require.NoError(t, err)
	require.True(t, rc.WrapTransport != nil)
}
