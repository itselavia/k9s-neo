// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package perftrace

import "time"

const (
	schemaVersion = 1

	EventSessionStart        = "session_start"
	EventConfigSnapshot      = "config_snapshot"
	EventKubeRequestComplete = "kube_request_complete"
	EventKubeStreamOpen      = "kube_stream_open"
	EventKubeStreamClose     = "kube_stream_close"
	EventSessionEnd          = "session_end"
)

// Options represents runtime perf trace settings.
type Options struct {
	File     string
	Scenario string
	RunID    string
	App      string
	Version  string
	Commit   string
}

// Event represents one structured perf trace event.
type Event struct {
	SchemaVersion int       `json:"schema_version"`
	Seq           int64     `json:"seq"`
	TS            time.Time `json:"ts"`
	Type          string    `json:"type"`
	RunID         string    `json:"run_id,omitempty"`
	Scenario      string    `json:"scenario,omitempty"`

	App            string `json:"app,omitempty"`
	Version        string `json:"version,omitempty"`
	Commit         string `json:"commit,omitempty"`
	PID            int    `json:"pid,omitempty"`
	Kubeconfig     string `json:"kubeconfig,omitempty"`
	ContextFlag    string `json:"context_flag,omitempty"`
	ClusterFlag    string `json:"cluster_flag,omitempty"`
	UserFlag       string `json:"user_flag,omitempty"`
	NamespaceFlag  string `json:"namespace_flag,omitempty"`
	AllNamespaces  bool   `json:"all_namespaces,omitempty"`
	Command        string `json:"command,omitempty"`
	ReadOnly       bool   `json:"readonly,omitempty"`
	RequestTimeout string `json:"request_timeout,omitempty"`

	RequestID     string  `json:"request_id,omitempty"`
	ClientRole    string  `json:"client_role,omitempty"`
	HTTPMethod    string  `json:"http_method,omitempty"`
	KubeVerb      string  `json:"kube_verb,omitempty"`
	Host          string  `json:"host,omitempty"`
	Path          string  `json:"path,omitempty"`
	Query         string  `json:"query,omitempty"`
	APIGroup      string  `json:"api_group,omitempty"`
	APIVersion    string  `json:"api_version,omitempty"`
	Resource      string  `json:"resource,omitempty"`
	Subresource   string  `json:"subresource,omitempty"`
	Namespace     string  `json:"namespace,omitempty"`
	Name          string  `json:"name,omitempty"`
	StatusCode    int     `json:"status_code,omitempty"`
	DurationMS    float64 `json:"duration_ms,omitempty"`
	ResponseBytes int64   `json:"response_bytes,omitempty"`
	ContentLength int64   `json:"content_length,omitempty"`
	ContentType   string  `json:"content_type,omitempty"`
	Streaming     bool    `json:"streaming,omitempty"`
	Watch         bool    `json:"watch,omitempty"`
	Follow        bool    `json:"follow,omitempty"`
	ResponseKind  string  `json:"response_kind,omitempty"`
	ItemCount     int     `json:"item_count,omitempty"`
	BodyInspected bool    `json:"body_inspected,omitempty"`
	Error         string  `json:"error,omitempty"`
}

func durationMS(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}
