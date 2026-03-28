// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAliasEnsureCharacterizesSupportedCoreCommands(t *testing.T) {
	oldMeta := MetaAccess
	MetaAccess = NewMeta()
	t.Cleanup(func() {
		MetaAccess = oldMeta
	})

	alias := NewAlias(newCharacterizationFactory(t))

	_, err := alias.Ensure(initCharacterizationConfig(t))
	require.NoError(t, err)

	uu := map[string]*client.GVR{
		"po":           client.PodGVR,
		"pod":          client.PodGVR,
		"pods":         client.PodGVR,
		"svc":          client.SvcGVR,
		"service":      client.SvcGVR,
		"services":     client.SvcGVR,
		"deploy":       client.DpGVR,
		"deployment":   client.DpGVR,
		"deployments":  client.DpGVR,
		"ds":           client.DsGVR,
		"daemonset":    client.DsGVR,
		"daemonsets":   client.DsGVR,
		"sts":          client.StsGVR,
		"statefulset":  client.StsGVR,
		"statefulsets": client.StsGVR,
		"rs":           client.RsGVR,
		"replicaset":   client.RsGVR,
		"replicasets":  client.RsGVR,
		"job":          client.JobGVR,
		"jobs":         client.JobGVR,
		"cj":           client.CjGVR,
		"cronjob":      client.CjGVR,
		"cronjobs":     client.CjGVR,
		"ns":           client.NsGVR,
		"namespace":    client.NsGVR,
		"namespaces":   client.NsGVR,
		"node":         client.NodeGVR,
		"nodes":        client.NodeGVR,
		"event":        client.EvGVR,
		"events":       client.EvGVR,
	}

	for aliasText, expected := range uu {
		t.Run(aliasText, func(t *testing.T) {
			resolved, ok := alias.Resolve(cmd.NewInterpreter(aliasText))
			require.True(t, ok)
			assert.Equal(t, expected, resolved)
		})
	}
}

func TestAliasEnsureStaticCoreRegistryResolvesAgonesAllowlist(t *testing.T) {
	oldMeta := MetaAccess
	MetaAccess = NewMeta()
	t.Cleanup(func() {
		MetaAccess = oldMeta
	})

	alias := NewAlias(newStaticCoreRegistryFactory())

	_, err := alias.Ensure(initCharacterizationConfig(t))
	require.NoError(t, err)

	uu := map[string]*client.GVR{
		"gameserver":                client.GsGVR,
		"gameservers":               client.GsGVR,
		"agones.dev/v1/gameservers": client.GsGVR,
		"gameservers.agones.dev":    client.GsGVR,
		"fleet":                     client.FltGVR,
		"fleets":                    client.FltGVR,
		"allocation.agones.dev/v1/gameserverallocations": client.GsaGVR,
		"gameserverallocations.allocation.agones.dev":    client.GsaGVR,
		"autoscaling.agones.dev/v1/fleetautoscalers":     client.FasGVR,
		"fleetautoscalers.autoscaling.agones.dev":        client.FasGVR,
	}

	for aliasText, expected := range uu {
		t.Run(aliasText, func(t *testing.T) {
			resolved, ok := alias.Resolve(cmd.NewInterpreter(aliasText))
			require.True(t, ok)
			assert.Equal(t, expected, resolved)
		})
	}
}

func TestAliasEnsureStaticCoreRegistryResolvesSupportedCoreCommands(t *testing.T) {
	oldMeta := MetaAccess
	MetaAccess = NewMeta()
	t.Cleanup(func() {
		MetaAccess = oldMeta
	})

	alias := NewAlias(newStaticCoreRegistryFactory())

	_, err := alias.Ensure(initCharacterizationConfig(t))
	require.NoError(t, err)

	uu := map[string]*client.GVR{
		"po":           client.PodGVR,
		"pod":          client.PodGVR,
		"pods":         client.PodGVR,
		"svc":          client.SvcGVR,
		"service":      client.SvcGVR,
		"services":     client.SvcGVR,
		"deploy":       client.DpGVR,
		"deployment":   client.DpGVR,
		"deployments":  client.DpGVR,
		"ds":           client.DsGVR,
		"daemonset":    client.DsGVR,
		"daemonsets":   client.DsGVR,
		"sts":          client.StsGVR,
		"statefulset":  client.StsGVR,
		"statefulsets": client.StsGVR,
		"rs":           client.RsGVR,
		"replicaset":   client.RsGVR,
		"replicasets":  client.RsGVR,
		"job":          client.JobGVR,
		"jobs":         client.JobGVR,
		"cj":           client.CjGVR,
		"cronjob":      client.CjGVR,
		"cronjobs":     client.CjGVR,
		"ns":           client.NsGVR,
		"namespace":    client.NsGVR,
		"namespaces":   client.NsGVR,
		"node":         client.NodeGVR,
		"nodes":        client.NodeGVR,
		"event":        client.EvGVR,
		"events":       client.EvGVR,
	}

	for aliasText, expected := range uu {
		t.Run(aliasText, func(t *testing.T) {
			resolved, ok := alias.Resolve(cmd.NewInterpreter(aliasText))
			require.True(t, ok)
			assert.Equal(t, expected, resolved)
		})
	}
}
