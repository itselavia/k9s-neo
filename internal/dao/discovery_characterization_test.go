// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/watch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery/cached/disk"
	restclient "k8s.io/client-go/rest"
)

func newCharacterizationFactory(t *testing.T) *watch.Factory {
	t.Helper()

	server := newCharacterizationDiscoveryServer(t)
	t.Cleanup(server.Close)

	conn := characterizationConnection{
		discovery: newCharacterizationCachedDiscovery(t, server.URL),
		cfg:       newCharacterizationConfig(),
	}

	return watch.NewFactory(conn)
}

func initCharacterizationConfig(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	t.Setenv(config.K9sEnvConfigDir, dir)
	if err := config.InitLocs(); err != nil {
		t.Fatalf("init config locations: %v", err)
	}

	return filepath.Join(dir, "context-aliases.yaml")
}

type characterizationConnection struct {
	client.Connection
	discovery *disk.CachedDiscoveryClient
	cfg       *client.Config
}

func (c characterizationConnection) Config() *client.Config {
	return c.cfg
}

func (c characterizationConnection) ConnectionOK() bool {
	return true
}

func (c characterizationConnection) CachedDiscovery() (*disk.CachedDiscoveryClient, error) {
	return c.discovery, nil
}

func newCharacterizationCachedDiscovery(t *testing.T, host string) *disk.CachedDiscoveryClient {
	t.Helper()

	cacheDir := t.TempDir()
	cfg := &restclient.Config{
		Host: host,
	}

	discovery, err := disk.NewCachedDiscoveryClientForConfig(cfg, cacheDir, "", time.Minute)
	if err != nil {
		t.Fatalf("create cached discovery client: %v", err)
	}

	return discovery
}

func newCharacterizationConfig() *client.Config {
	cfg := client.NewConfig(genericclioptions.NewConfigFlags(false))
	cfg.SetSkipCRDAugment(true)

	return cfg
}

func newStaticCoreRegistryConfig() *client.Config {
	cfg := client.NewConfig(genericclioptions.NewConfigFlags(false))
	cfg.SetStaticCoreRegistry(true)

	return cfg
}

func newStaticCoreRegistryFactory() Factory {
	return registryTestFactory{
		conn: registryTestConnection{cfg: newStaticCoreRegistryConfig()},
	}
}

func newCharacterizationDiscoveryServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/api", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, metav1.APIVersions{
			Versions: []string{"v1"},
		})
	})
	mux.HandleFunc("/api/v1", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, metav1.APIResourceList{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{
					Name:         "pods",
					SingularName: "pod",
					Namespaced:   true,
					Kind:         "Pod",
					ShortNames:   []string{"po"},
				},
				{
					Name:         "services",
					SingularName: "service",
					Namespaced:   true,
					Kind:         "Service",
					ShortNames:   []string{"svc"},
				},
				{
					Name:         "namespaces",
					SingularName: "namespace",
					Namespaced:   false,
					Kind:         "Namespace",
					ShortNames:   []string{"ns"},
				},
				{
					Name:         "nodes",
					SingularName: "node",
					Namespaced:   false,
					Kind:         "Node",
				},
			},
		})
	})
	mux.HandleFunc("/apis", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, metav1.APIGroupList{
			Groups: []metav1.APIGroup{
				{
					Name: "apps",
					Versions: []metav1.GroupVersionForDiscovery{
						{GroupVersion: "apps/v1", Version: "v1"},
					},
					PreferredVersion: metav1.GroupVersionForDiscovery{
						GroupVersion: "apps/v1",
						Version:      "v1",
					},
				},
				{
					Name: "batch",
					Versions: []metav1.GroupVersionForDiscovery{
						{GroupVersion: "batch/v1", Version: "v1"},
					},
					PreferredVersion: metav1.GroupVersionForDiscovery{
						GroupVersion: "batch/v1",
						Version:      "v1",
					},
				},
				{
					Name: "events.k8s.io",
					Versions: []metav1.GroupVersionForDiscovery{
						{GroupVersion: "events.k8s.io/v1", Version: "v1"},
					},
					PreferredVersion: metav1.GroupVersionForDiscovery{
						GroupVersion: "events.k8s.io/v1",
						Version:      "v1",
					},
				},
			},
		})
	})
	mux.HandleFunc("/apis/apps/v1", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, metav1.APIResourceList{
			GroupVersion: "apps/v1",
			APIResources: []metav1.APIResource{
				{
					Name:         "deployments",
					SingularName: "deployment",
					Namespaced:   true,
					Kind:         "Deployment",
					ShortNames:   []string{"deploy"},
				},
				{
					Name:         "replicasets",
					SingularName: "replicaset",
					Namespaced:   true,
					Kind:         "ReplicaSet",
					ShortNames:   []string{"rs"},
				},
				{
					Name:         "daemonsets",
					SingularName: "daemonset",
					Namespaced:   true,
					Kind:         "DaemonSet",
					ShortNames:   []string{"ds"},
				},
				{
					Name:         "statefulsets",
					SingularName: "statefulset",
					Namespaced:   true,
					Kind:         "StatefulSet",
					ShortNames:   []string{"sts"},
				},
			},
		})
	})
	mux.HandleFunc("/apis/batch/v1", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, metav1.APIResourceList{
			GroupVersion: "batch/v1",
			APIResources: []metav1.APIResource{
				{
					Name:         "jobs",
					SingularName: "job",
					Namespaced:   true,
					Kind:         "Job",
				},
				{
					Name:         "cronjobs",
					SingularName: "cronjob",
					Namespaced:   true,
					Kind:         "CronJob",
					ShortNames:   []string{"cj"},
				},
			},
		})
	})
	mux.HandleFunc("/apis/events.k8s.io/v1", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, metav1.APIResourceList{
			GroupVersion: "events.k8s.io/v1",
			APIResources: []metav1.APIResource{
				{
					Name:         "events",
					SingularName: "event",
					Namespaced:   true,
					Kind:         "Event",
				},
			},
		})
	})

	return httptest.NewServer(mux)
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encode json response: %v", err)
	}
}
