// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package perftrace

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyRequest(t *testing.T) {
	tests := map[string]struct {
		method string
		rawURL string
		check  func(t *testing.T, info requestInfo)
	}{
		"namespaced-list": {
			method: "GET",
			rawURL: "https://example.com/api/v1/namespaces/big/pods",
			check: func(t *testing.T, info requestInfo) {
				assert.Equal(t, "list", info.kubeVerb)
				assert.Equal(t, "v1", info.apiVersion)
				assert.Equal(t, "pods", info.resource)
				assert.Equal(t, "big", info.namespace)
			},
		},
		"namespaced-get": {
			method: "GET",
			rawURL: "https://example.com/apis/apps/v1/namespaces/big/deployments/api",
			check: func(t *testing.T, info requestInfo) {
				assert.Equal(t, "get", info.kubeVerb)
				assert.Equal(t, "apps", info.apiGroup)
				assert.Equal(t, "v1", info.apiVersion)
				assert.Equal(t, "deployments", info.resource)
				assert.Equal(t, "api", info.name)
			},
		},
		"cluster-scoped-get": {
			method: "GET",
			rawURL: "https://example.com/api/v1/nodes/node-a",
			check: func(t *testing.T, info requestInfo) {
				assert.Equal(t, "get", info.kubeVerb)
				assert.Equal(t, "nodes", info.resource)
				assert.Equal(t, "node-a", info.name)
			},
		},
		"discovery-version": {
			method: "GET",
			rawURL: "https://example.com/version",
			check: func(t *testing.T, info requestInfo) {
				assert.Equal(t, "get", info.kubeVerb)
				assert.Empty(t, info.resource)
			},
		},
		"discovery-apis": {
			method: "GET",
			rawURL: "https://example.com/apis",
			check: func(t *testing.T, info requestInfo) {
				assert.Equal(t, "get", info.kubeVerb)
				assert.Empty(t, info.resource)
			},
		},
		"auth-review-post": {
			method: "POST",
			rawURL: "https://example.com/apis/authorization.k8s.io/v1/selfsubjectaccessreviews",
			check: func(t *testing.T, info requestInfo) {
				assert.Equal(t, "create", info.kubeVerb)
				assert.Equal(t, "authorization.k8s.io", info.apiGroup)
				assert.Equal(t, "selfsubjectaccessreviews", info.resource)
			},
		},
		"pod-log-follow": {
			method: "GET",
			rawURL: "https://example.com/api/v1/namespaces/big/pods/pod-a/log?follow=true",
			check: func(t *testing.T, info requestInfo) {
				assert.True(t, info.streaming)
				assert.True(t, info.follow)
				assert.Equal(t, "log", info.subresource)
			},
		},
		"watch-request": {
			method: "GET",
			rawURL: "https://example.com/api/v1/namespaces/big/pods?watch=true",
			check: func(t *testing.T, info requestInfo) {
				assert.True(t, info.streaming)
				assert.True(t, info.watch)
				assert.Equal(t, "watch", info.kubeVerb)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			u, err := url.Parse(tc.rawURL)
			require.NoError(t, err)
			tc.check(t, classifyRequest(tc.method, u))
		})
	}
}
