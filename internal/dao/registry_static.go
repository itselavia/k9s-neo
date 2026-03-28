// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"github.com/derailed/k9s/internal/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type staticResourceMeta struct {
	gvr  *client.GVR
	meta metav1.APIResource
}

func loadStaticCoreResources(m ResourceMetas) {
	loadStaticResources(m, staticCoreResources)
}

func loadStaticAgonesAllowlist(m ResourceMetas) {
	loadStaticResources(m, staticAgonesResources)
}

func loadStaticResources(m ResourceMetas, entries []staticResourceMeta) {
	for _, entry := range entries {
		meta := entry.meta
		m[entry.gvr] = &meta
	}
}

func readOnlyVerbs() []string {
	return []string{client.GetVerb, client.ListVerb, client.WatchVerb}
}

var staticCoreResources = []staticResourceMeta{
	{
		gvr: client.PodGVR,
		meta: metav1.APIResource{
			Name:         "pods",
			SingularName: "pod",
			Namespaced:   true,
			Kind:         "Pod",
			Group:        "",
			Version:      "v1",
			ShortNames:   []string{"po"},
			Verbs:        readOnlyVerbs(),
		},
	},
	{
		gvr: client.NodeGVR,
		meta: metav1.APIResource{
			Name:         "nodes",
			SingularName: "node",
			Namespaced:   false,
			Kind:         "Node",
			Group:        "",
			Version:      "v1",
			Verbs:        readOnlyVerbs(),
		},
	},
	{
		gvr: client.NsGVR,
		meta: metav1.APIResource{
			Name:         "namespaces",
			SingularName: "namespace",
			Namespaced:   false,
			Kind:         "Namespace",
			Group:        "",
			Version:      "v1",
			ShortNames:   []string{"ns"},
			Verbs:        readOnlyVerbs(),
		},
	},
	{
		gvr: client.EvGVR,
		meta: metav1.APIResource{
			Name:         "events",
			SingularName: "event",
			Namespaced:   true,
			Kind:         "Event",
			Group:        "events.k8s.io",
			Version:      "v1",
			Verbs:        readOnlyVerbs(),
		},
	},
	{
		gvr: client.DpGVR,
		meta: metav1.APIResource{
			Name:         "deployments",
			SingularName: "deployment",
			Namespaced:   true,
			Kind:         "Deployment",
			Group:        "apps",
			Version:      "v1",
			ShortNames:   []string{"deploy"},
			Verbs:        readOnlyVerbs(),
			Categories:   []string{scaleCat},
		},
	},
	{
		gvr: client.RsGVR,
		meta: metav1.APIResource{
			Name:         "replicasets",
			SingularName: "replicaset",
			Namespaced:   true,
			Kind:         "ReplicaSet",
			Group:        "apps",
			Version:      "v1",
			ShortNames:   []string{"rs"},
			Verbs:        readOnlyVerbs(),
			Categories:   []string{scaleCat},
		},
	},
	{
		gvr: client.DsGVR,
		meta: metav1.APIResource{
			Name:         "daemonsets",
			SingularName: "daemonset",
			Namespaced:   true,
			Kind:         "DaemonSet",
			Group:        "apps",
			Version:      "v1",
			ShortNames:   []string{"ds"},
			Verbs:        readOnlyVerbs(),
		},
	},
	{
		gvr: client.StsGVR,
		meta: metav1.APIResource{
			Name:         "statefulsets",
			SingularName: "statefulset",
			Namespaced:   true,
			Kind:         "StatefulSet",
			Group:        "apps",
			Version:      "v1",
			ShortNames:   []string{"sts"},
			Verbs:        readOnlyVerbs(),
			Categories:   []string{scaleCat},
		},
	},
	{
		gvr: client.JobGVR,
		meta: metav1.APIResource{
			Name:         "jobs",
			SingularName: "job",
			Namespaced:   true,
			Kind:         "Job",
			Group:        "batch",
			Version:      "v1",
			Verbs:        readOnlyVerbs(),
		},
	},
	{
		gvr: client.CjGVR,
		meta: metav1.APIResource{
			Name:         "cronjobs",
			SingularName: "cronjob",
			Namespaced:   true,
			Kind:         "CronJob",
			Group:        "batch",
			Version:      "v1",
			ShortNames:   []string{"cj"},
			Verbs:        readOnlyVerbs(),
		},
	},
	{
		gvr: client.SvcGVR,
		meta: metav1.APIResource{
			Name:         "services",
			SingularName: "service",
			Namespaced:   true,
			Kind:         "Service",
			Group:        "",
			Version:      "v1",
			ShortNames:   []string{"svc"},
			Verbs:        readOnlyVerbs(),
		},
	},
}

var staticAgonesResources = []staticResourceMeta{
	{
		gvr: client.GameServerGVR,
		meta: metav1.APIResource{
			Name:         "gameservers",
			SingularName: "gameserver",
			Namespaced:   true,
			Kind:         "GameServer",
			Group:        "agones.dev",
			Version:      "v1",
			Verbs:        readOnlyVerbs(),
			Categories:   []string{crdCat},
		},
	},
	{
		gvr: client.FleetGVR,
		meta: metav1.APIResource{
			Name:         "fleets",
			SingularName: "fleet",
			Namespaced:   true,
			Kind:         "Fleet",
			Group:        "agones.dev",
			Version:      "v1",
			Verbs:        readOnlyVerbs(),
			Categories:   []string{crdCat},
		},
	},
	{
		gvr: client.GameServerAllocationGVR,
		meta: metav1.APIResource{
			Name:         "gameserverallocations",
			SingularName: "gameserverallocation",
			Namespaced:   true,
			Kind:         "GameServerAllocation",
			Group:        "allocation.agones.dev",
			Version:      "v1",
			Verbs:        readOnlyVerbs(),
			Categories:   []string{crdCat},
		},
	},
	{
		gvr: client.FleetAutoscalerGVR,
		meta: metav1.APIResource{
			Name:         "fleetautoscalers",
			SingularName: "fleetautoscaler",
			Namespaced:   true,
			Kind:         "FleetAutoscaler",
			Group:        "autoscaling.agones.dev",
			Version:      "v1",
			Verbs:        readOnlyVerbs(),
			Categories:   []string{crdCat},
		},
	},
}
