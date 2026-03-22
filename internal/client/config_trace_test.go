// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client_test

import (
	"net/http"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/perftrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestConfigSetPerfTraceAndRESTConfigFor(t *testing.T) {
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
	}

	trace, err := perftrace.NewSession(perftrace.Options{File: filepath.Join(t.TempDir(), "trace.jsonl")})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, trace.Close(nil))
	}()

	cfg := client.NewConfig(&flags)
	cfg.SetPerfTrace(trace)
	proxyURL, err := url.Parse("http://localhost:1234")
	require.NoError(t, err)
	cfg.SetProxy(func(*http.Request) (*url.URL, error) {
		return proxyURL, nil
	})

	rc, err := cfg.RESTConfigFor("core")
	require.NoError(t, err)
	gotURL, err := rc.Proxy((&http.Request{}))
	require.NoError(t, err)
	assert.Equal(t, proxyURL.String(), gotURL.String())
}
