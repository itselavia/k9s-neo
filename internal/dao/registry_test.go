// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery/cached/disk"
)

func TestLoadResourcesCharacterizesSupportedV0Metas(t *testing.T) {
	m := NewMeta()

	require.NoError(t, m.LoadResources(newCharacterizationFactory(t)))

	uu := map[string]struct {
		gvr        *client.GVR
		name       string
		singular   string
		kind       string
		namespaced bool
		scalable   bool
	}{
		"pods": {
			gvr:        client.PodGVR,
			name:       "pods",
			singular:   "pod",
			kind:       "Pod",
			namespaced: true,
		},
		"nodes": {
			gvr:        client.NodeGVR,
			name:       "nodes",
			singular:   "node",
			kind:       "Node",
			namespaced: false,
		},
		"namespaces": {
			gvr:        client.NsGVR,
			name:       "namespaces",
			singular:   "namespace",
			kind:       "Namespace",
			namespaced: false,
		},
		"events": {
			gvr:        client.EvGVR,
			name:       "events",
			singular:   "event",
			kind:       "Event",
			namespaced: true,
		},
		"deployments": {
			gvr:        client.DpGVR,
			name:       "deployments",
			singular:   "deployment",
			kind:       "Deployment",
			namespaced: true,
			scalable:   true,
		},
		"replicasets": {
			gvr:        client.RsGVR,
			name:       "replicasets",
			singular:   "replicaset",
			kind:       "ReplicaSet",
			namespaced: true,
			scalable:   true,
		},
		"daemonsets": {
			gvr:        client.DsGVR,
			name:       "daemonsets",
			singular:   "daemonset",
			kind:       "DaemonSet",
			namespaced: true,
		},
		"statefulsets": {
			gvr:        client.StsGVR,
			name:       "statefulsets",
			singular:   "statefulset",
			kind:       "StatefulSet",
			namespaced: true,
			scalable:   true,
		},
		"jobs": {
			gvr:        client.JobGVR,
			name:       "jobs",
			singular:   "job",
			kind:       "Job",
			namespaced: true,
		},
		"cronjobs": {
			gvr:        client.CjGVR,
			name:       "cronjobs",
			singular:   "cronjob",
			kind:       "CronJob",
			namespaced: true,
		},
		"services": {
			gvr:        client.SvcGVR,
			name:       "services",
			singular:   "service",
			kind:       "Service",
			namespaced: true,
		},
	}

	for name, tc := range uu {
		t.Run(name, func(t *testing.T) {
			meta, err := m.MetaFor(tc.gvr)
			require.NoError(t, err)

			assert.Equal(t, tc.name, meta.Name)
			assert.Equal(t, tc.singular, meta.SingularName)
			assert.Equal(t, tc.kind, meta.Kind)
			assert.Equal(t, tc.namespaced, meta.Namespaced)
			assert.Equal(t, tc.scalable, IsScalable(meta))
		})
	}
}

func TestLoadResourcesCharacterizesSupportedV0Namespacedness(t *testing.T) {
	m := NewMeta()

	require.NoError(t, m.LoadResources(newCharacterizationFactory(t)))

	uu := map[string]struct {
		gvr        *client.GVR
		namespaced bool
	}{
		"pods": {
			gvr:        client.PodGVR,
			namespaced: true,
		},
		"services": {
			gvr:        client.SvcGVR,
			namespaced: true,
		},
		"deployments": {
			gvr:        client.DpGVR,
			namespaced: true,
		},
		"events": {
			gvr:        client.EvGVR,
			namespaced: true,
		},
		"nodes": {
			gvr:        client.NodeGVR,
			namespaced: false,
		},
		"namespaces": {
			gvr:        client.NsGVR,
			namespaced: false,
		},
	}

	for name, tc := range uu {
		t.Run(name, func(t *testing.T) {
			namespaced, err := m.IsNamespaced(tc.gvr)
			require.NoError(t, err)
			assert.Equal(t, tc.namespaced, namespaced)
		})
	}
}

func TestLoadResourcesStaticCoreRegistryLoadsCuratedResources(t *testing.T) {
	m := NewMeta()

	require.NoError(t, m.LoadResources(newStaticCoreRegistryFactory()))

	for name, gvr := range map[string]*client.GVR{
		"pods":                  client.PodGVR,
		"nodes":                 client.NodeGVR,
		"namespaces":            client.NsGVR,
		"events":                client.EvGVR,
		"deployments":           client.DpGVR,
		"replicasets":           client.RsGVR,
		"daemonsets":            client.DsGVR,
		"statefulsets":          client.StsGVR,
		"jobs":                  client.JobGVR,
		"cronjobs":              client.CjGVR,
		"services":              client.SvcGVR,
		"gameservers":           client.GsGVR,
		"fleets":                client.FltGVR,
		"gameserverallocations": client.GsaGVR,
		"fleetautoscalers":      client.FasGVR,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := m.MetaFor(gvr)
			require.NoError(t, err)
		})
	}

	_, err := m.MetaFor(client.CmGVR)
	require.Error(t, err)

	_, err = m.MetaFor(client.NewGVR("example.dev/v1/widgets"))
	require.Error(t, err)
}

func TestLoadResourcesStaticCoreRegistrySkipsDiscovery(t *testing.T) {
	cfg := newStaticCoreRegistryConfig()
	conn := &staticCoreDiscoveryCountConnection{cfg: cfg}
	factory := registryTestFactory{conn: conn}
	m := NewMeta()

	require.NoError(t, m.LoadResources(factory))
	assert.Zero(t, conn.cachedDiscoveryCalls)
}

func TestMetaFor(t *testing.T) {
	uu := map[string]struct {
		gvr *client.GVR
		err error
		e   metav1.APIResource
	}{
		"xray-gvr": {
			gvr: client.XGVR,
			e: metav1.APIResource{
				Name:         "xrays",
				Kind:         "XRays",
				SingularName: "xray",
				Categories:   []string{k9sCat},
			},
		},

		"xray": {
			gvr: client.NewGVR("xrays"),
			e: metav1.APIResource{
				Name:         "xrays",
				Kind:         "XRays",
				SingularName: "xray",
				Categories:   []string{k9sCat},
			},
		},

		"policy": {
			gvr: client.NewGVR("policy"),
			e: metav1.APIResource{
				Name:       "policies",
				Kind:       "Rules",
				Namespaced: true,
				Categories: []string{k9sCat},
			},
		},

		"helm": {
			gvr: client.NewGVR("helm"),
			e: metav1.APIResource{
				Name:       "helm",
				Kind:       "Helm",
				Namespaced: true,
				Verbs:      []string{"delete"},
				Categories: []string{helmCat},
			},
		},

		"toast": {
			gvr: client.NewGVR("blah"),
			err: errors.New("no resource meta defined for\n \"blah\""),
		},
	}

	m := NewMeta()
	require.NoError(t, m.LoadResources(nil))
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			meta, err := m.MetaFor(u.gvr)
			assert.Equal(t, u.err, err)
			if err == nil {
				assert.Equal(t, &u.e, meta)
			}
		})
	}
}

func TestShouldAugmentCRDs(t *testing.T) {
	flags := genericclioptions.NewConfigFlags(false)
	cfg := client.NewConfig(flags)

	uu := map[string]struct {
		factory Factory
		want    bool
	}{
		"nil-factory": {
			want: true,
		},
		"nil-connection": {
			factory: registryTestFactory{},
			want:    true,
		},
		"default": {
			factory: registryTestFactory{
				conn: registryTestConnection{cfg: cfg},
			},
			want: true,
		},
		"skip-enabled": {
			factory: registryTestFactory{
				conn: registryTestConnection{cfg: func() *client.Config {
					cfg := client.NewConfig(genericclioptions.NewConfigFlags(false))
					cfg.SetSkipCRDAugment(true)
					return cfg
				}()},
			},
			want: false,
		},
	}

	for name, tc := range uu {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, shouldAugmentCRDs(tc.factory))
		})
	}
}

type registryTestFactory struct {
	Factory
	conn client.Connection
}

func (f registryTestFactory) Client() client.Connection {
	return f.conn
}

type registryTestConnection struct {
	client.Connection
	cfg *client.Config
}

func (c registryTestConnection) Config() *client.Config {
	return c.cfg
}

type staticCoreDiscoveryCountConnection struct {
	client.Connection
	cfg                  *client.Config
	cachedDiscoveryCalls int
}

func (c *staticCoreDiscoveryCountConnection) Config() *client.Config {
	return c.cfg
}

func (c *staticCoreDiscoveryCountConnection) CachedDiscovery() (*disk.CachedDiscoveryClient, error) {
	c.cachedDiscoveryCalls++
	return nil, errors.New("cached discovery should not be called in static core mode")
}
