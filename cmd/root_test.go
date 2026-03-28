// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestPerfTraceFlagsAreHidden(t *testing.T) {
	for _, name := range []string{
		"perf-trace-file",
		"perf-trace-scenario",
		"perf-trace-run-id",
		"perf-skip-crd-augment",
		"perf-static-core-registry",
		"perf-skip-namespace-validation",
	} {
		flag := rootCmd.Flags().Lookup(name)
		require.NotNil(t, flag)
		assert.True(t, flag.Hidden)
	}
}

func TestNewPerfTraceSessionFailsFast(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "not-a-dir")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0o644))

	prev := perfTraceFile
	t.Cleanup(func() {
		perfTraceFile = prev
	})
	perfTraceFile = filepath.Join(file, "trace.jsonl")

	_, err := newPerfTraceSession()
	require.Error(t, err)
}

func TestLoadConfigurationWithoutTrace(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/apis":
			_, _ = io.WriteString(w, `{"kind":"APIGroupList","apiVersion":"v1","groups":[]}`)
		case "/version":
			_, _ = io.WriteString(w, `{"major":"1","minor":"30","gitVersion":"v1.30.0"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tmp := t.TempDir()
	t.Setenv(config.K9sEnvConfigDir, tmp)
	require.NoError(t, config.InitLocs())
	require.NoError(t, config.NewConfig(client.NewConfig(genericclioptions.NewConfigFlags(false))).SaveFile(config.AppConfigFile))

	kubeconfig := filepath.Join(tmp, "kubeconfig")
	require.NoError(t, os.WriteFile(kubeconfig, []byte(fmt.Sprintf(`
apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: %s
  name: test
contexts:
- context:
    cluster: test
    user: test
    namespace: default
  name: test
current-context: test
kind: Config
users:
- name: test
  user:
    token: fake
`, server.URL)), 0o644))

	prevK8sFlags, prevK9sFlags := k8sFlags, k9sFlags
	t.Cleanup(func() {
		k8sFlags = prevK8sFlags
		k9sFlags = prevK9sFlags
	})

	k8sFlags = genericclioptions.NewConfigFlags(false)
	context := "test"
	cluster := "test"
	user := "test"
	namespace := "default"
	k8sFlags.KubeConfig = &kubeconfig
	k8sFlags.Context = &context
	k8sFlags.ClusterName = &cluster
	k8sFlags.AuthInfoName = &user
	k8sFlags.Namespace = &namespace
	k9sFlags = config.NewFlags()

	cfg, err := loadConfiguration(nil)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.GetConnection())
	assert.Equal(t, "test", cfg.K9s.ActiveContextName())
}
