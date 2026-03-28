package view

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/k9s/internal/watch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery/cached/disk"
	restclient "k8s.io/client-go/rest"
)

func Test_viewMetaFor(t *testing.T) {
	uu := map[string]struct {
		cmd string
		gvr *client.GVR
		p   *cmd.Interpreter
		err error
	}{
		"empty": {
			cmd: "",
			gvr: client.PodGVR,
			err: errors.New("`` command not found"),
		},

		"toast": {
			cmd: "v1/pd",
			gvr: client.PodGVR,
			err: errors.New("`v1/pd` command not found"),
		},

		"gvr": {
			cmd: "v1/pods",
			gvr: client.PodGVR,
			p:   cmd.NewInterpreter("v1/pods"),
			err: errors.New("blah"),
		},

		"short-name": {
			cmd: "po",
			gvr: client.PodGVR,
			p:   cmd.NewInterpreter("v1/pods", "po"),
			err: errors.New("blee"),
		},

		"custom-alias": {
			cmd: "pdl",
			gvr: client.PodGVR,
			p:   cmd.NewInterpreter("v1/pods @fred 'app=blee' default", "pdl"),
			err: errors.New("blee"),
		},

		"inception": {
			cmd: "pdal blee",
			gvr: client.PodGVR,
			p:   cmd.NewInterpreter("v1/pods @fred 'app=blee' blee", "pdal", "pod"),
			err: errors.New("blee"),
		},
	}

	c := &Command{
		alias: &dao.Alias{
			Aliases: config.NewAliases(),
		},
	}
	c.alias.Define(client.PodGVR, "po", "pod", "pods", client.PodGVR.String())
	c.alias.Define(client.NewGVR("pod default"), "pd")
	c.alias.Define(client.NewGVR("pod @fred 'app=blee' default"), "pdl")
	c.alias.Define(client.NewGVR("pdl"), "pdal")

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			p := cmd.NewInterpreter(u.cmd)
			gvr, _, acmd, err := c.viewMetaFor(p)
			if err != nil {
				assert.Equal(t, u.err.Error(), err.Error())
			} else {
				assert.Equal(t, u.gvr, gvr)
				assert.Equal(t, u.p, acmd)
			}
		})
	}
}

func TestCommandInitCharacterizesSupportedCoreCommands(t *testing.T) {
	oldMeta := dao.MetaAccess
	dao.MetaAccess = dao.NewMeta()
	t.Cleanup(func() {
		dao.MetaAccess = oldMeta
	})

	dir := t.TempDir()
	t.Setenv(config.K9sEnvConfigDir, dir)
	require.NoError(t, config.InitLocs())

	server := newCommandDiscoveryServer(t)
	t.Cleanup(server.Close)

	app := &App{
		factory: watch.NewFactory(commandCharacterizationConnection{
			discovery: newCommandCachedDiscovery(t, server.URL),
			cfg:       newCommandCharacterizationConfig(),
		}),
	}

	c := NewCommand(app)
	require.NoError(t, c.Init(filepath.Join(dir, "context-aliases.yaml")))

	uu := map[string]*client.GVR{
		"po":     client.PodGVR,
		"svc":    client.SvcGVR,
		"deploy": client.DpGVR,
		"ds":     client.DsGVR,
		"sts":    client.StsGVR,
		"rs":     client.RsGVR,
		"job":    client.JobGVR,
		"cj":     client.CjGVR,
		"ns":     client.NsGVR,
		"nodes":  client.NodeGVR,
		"events": client.EvGVR,
	}

	for commandText, expected := range uu {
		t.Run(commandText, func(t *testing.T) {
			gvr, _, _, err := c.viewMetaFor(cmd.NewInterpreter(commandText))
			require.NoError(t, err)
			assert.Equal(t, expected, gvr)
		})
	}
}

func TestCommandInitStaticCoreRegistrySupportsAgonesAllowlist(t *testing.T) {
	oldMeta := dao.MetaAccess
	dao.MetaAccess = dao.NewMeta()
	t.Cleanup(func() {
		dao.MetaAccess = oldMeta
	})

	dir := t.TempDir()
	t.Setenv(config.K9sEnvConfigDir, dir)
	require.NoError(t, config.InitLocs())

	app := &App{
		factory: watch.NewFactory(commandCharacterizationConnection{
			cfg: newCommandStaticCoreRegistryConfig(),
		}),
	}

	c := NewCommand(app)
	require.NoError(t, c.Init(filepath.Join(dir, "context-aliases.yaml")))

	uu := map[string]*client.GVR{
		"gameservers.agones.dev":                      client.GsGVR,
		"agones.dev/v1/gameservers":                   client.GsGVR,
		"fleets.agones.dev":                           client.FltGVR,
		"gameserverallocations.allocation.agones.dev": client.GsaGVR,
		"fleetautoscalers.autoscaling.agones.dev":     client.FasGVR,
	}

	for commandText, expected := range uu {
		t.Run(commandText, func(t *testing.T) {
			gvr, _, _, err := c.viewMetaFor(cmd.NewInterpreter(commandText))
			require.NoError(t, err)
			assert.Equal(t, expected, gvr)
		})
	}
}

func TestCommandInitStaticCoreRegistrySupportsCoreCommands(t *testing.T) {
	oldMeta := dao.MetaAccess
	dao.MetaAccess = dao.NewMeta()
	t.Cleanup(func() {
		dao.MetaAccess = oldMeta
	})

	dir := t.TempDir()
	t.Setenv(config.K9sEnvConfigDir, dir)
	require.NoError(t, config.InitLocs())

	app := &App{
		factory: watch.NewFactory(commandCharacterizationConnection{
			cfg: newCommandStaticCoreRegistryConfig(),
		}),
	}

	c := NewCommand(app)
	require.NoError(t, c.Init(filepath.Join(dir, "context-aliases.yaml")))

	uu := map[string]*client.GVR{
		"po":     client.PodGVR,
		"svc":    client.SvcGVR,
		"deploy": client.DpGVR,
		"ds":     client.DsGVR,
		"sts":    client.StsGVR,
		"rs":     client.RsGVR,
		"job":    client.JobGVR,
		"cj":     client.CjGVR,
		"ns":     client.NsGVR,
		"nodes":  client.NodeGVR,
		"events": client.EvGVR,
	}

	for commandText, expected := range uu {
		t.Run(commandText, func(t *testing.T) {
			gvr, _, _, err := c.viewMetaFor(cmd.NewInterpreter(commandText))
			require.NoError(t, err)
			assert.Equal(t, expected, gvr)
		})
	}
}

func TestCommandInitStaticCoreRegistryRejectsUnsupportedGenericCRD(t *testing.T) {
	oldMeta := dao.MetaAccess
	dao.MetaAccess = dao.NewMeta()
	t.Cleanup(func() {
		dao.MetaAccess = oldMeta
	})

	dir := t.TempDir()
	t.Setenv(config.K9sEnvConfigDir, dir)
	require.NoError(t, config.InitLocs())

	app := &App{
		factory: watch.NewFactory(commandCharacterizationConnection{
			cfg: newCommandStaticCoreRegistryConfig(),
		}),
	}

	c := NewCommand(app)
	require.NoError(t, c.Init(filepath.Join(dir, "context-aliases.yaml")))

	_, _, _, err := c.viewMetaFor(cmd.NewInterpreter("widgets.example.dev"))
	require.EqualError(t, err, "`widgets.example.dev` command not found")
}

type commandCharacterizationConnection struct {
	client.Connection
	discovery *disk.CachedDiscoveryClient
	cfg       *client.Config
}

func (c commandCharacterizationConnection) Config() *client.Config {
	return c.cfg
}

func (c commandCharacterizationConnection) ConnectionOK() bool {
	return true
}

func (c commandCharacterizationConnection) CachedDiscovery() (*disk.CachedDiscoveryClient, error) {
	return c.discovery, nil
}

func newCommandCachedDiscovery(t *testing.T, host string) *disk.CachedDiscoveryClient {
	t.Helper()

	discovery, err := disk.NewCachedDiscoveryClientForConfig(&restclient.Config{Host: host}, t.TempDir(), "", time.Minute)
	require.NoError(t, err)

	return discovery
}

func newCommandCharacterizationConfig() *client.Config {
	cfg := client.NewConfig(genericclioptions.NewConfigFlags(false))
	cfg.SetSkipCRDAugment(true)

	return cfg
}

func newCommandStaticCoreRegistryConfig() *client.Config {
	cfg := client.NewConfig(genericclioptions.NewConfigFlags(false))
	cfg.SetStaticCoreRegistry(true)

	return cfg
}

func newCommandDiscoveryServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/api", func(w http.ResponseWriter, _ *http.Request) {
		writeCommandJSON(t, w, metav1.APIVersions{Versions: []string{"v1"}})
	})
	mux.HandleFunc("/api/v1", func(w http.ResponseWriter, _ *http.Request) {
		writeCommandJSON(t, w, metav1.APIResourceList{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", SingularName: "pod", Namespaced: true, Kind: "Pod", ShortNames: []string{"po"}},
				{Name: "services", SingularName: "service", Namespaced: true, Kind: "Service", ShortNames: []string{"svc"}},
				{Name: "namespaces", SingularName: "namespace", Namespaced: false, Kind: "Namespace", ShortNames: []string{"ns"}},
				{Name: "nodes", SingularName: "node", Namespaced: false, Kind: "Node"},
			},
		})
	})
	mux.HandleFunc("/apis", func(w http.ResponseWriter, _ *http.Request) {
		writeCommandJSON(t, w, metav1.APIGroupList{
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
		writeCommandJSON(t, w, metav1.APIResourceList{
			GroupVersion: "apps/v1",
			APIResources: []metav1.APIResource{
				{Name: "deployments", SingularName: "deployment", Namespaced: true, Kind: "Deployment", ShortNames: []string{"deploy"}},
				{Name: "replicasets", SingularName: "replicaset", Namespaced: true, Kind: "ReplicaSet", ShortNames: []string{"rs"}},
				{Name: "daemonsets", SingularName: "daemonset", Namespaced: true, Kind: "DaemonSet", ShortNames: []string{"ds"}},
				{Name: "statefulsets", SingularName: "statefulset", Namespaced: true, Kind: "StatefulSet", ShortNames: []string{"sts"}},
			},
		})
	})
	mux.HandleFunc("/apis/batch/v1", func(w http.ResponseWriter, _ *http.Request) {
		writeCommandJSON(t, w, metav1.APIResourceList{
			GroupVersion: "batch/v1",
			APIResources: []metav1.APIResource{
				{Name: "jobs", SingularName: "job", Namespaced: true, Kind: "Job"},
				{Name: "cronjobs", SingularName: "cronjob", Namespaced: true, Kind: "CronJob", ShortNames: []string{"cj"}},
			},
		})
	})
	mux.HandleFunc("/apis/events.k8s.io/v1", func(w http.ResponseWriter, _ *http.Request) {
		writeCommandJSON(t, w, metav1.APIResourceList{
			GroupVersion: "events.k8s.io/v1",
			APIResources: []metav1.APIResource{
				{Name: "events", SingularName: "event", Namespaced: true, Kind: "Event"},
			},
		})
	})

	return httptest.NewServer(mux)
}

func writeCommandJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(v))
}
