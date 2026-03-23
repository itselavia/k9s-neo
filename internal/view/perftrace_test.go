// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/perftrace"
	viewcmd "github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
	version "k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	disk "k8s.io/client-go/discovery/cached/disk"
	dynamic "k8s.io/client-go/dynamic"
	kubernetes "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

func TestRecordAfterDrawMarksFirstUsefulRowAfterFirstRender(t *testing.T) {
	tracePath := filepath.Join(t.TempDir(), "trace.jsonl")
	trace, err := perftrace.NewSession(perftrace.Options{File: tracePath})
	require.NoError(t, err)

	flags := genericclioptions.NewConfigFlags(false)
	kcfg := client.NewConfig(flags)
	kcfg.SetPerfTrace(trace)

	cfg := mock.NewMockConfig(t)
	cfg.SetConnection(traceConnection{cfg: kcfg})

	app := NewApp(cfg)
	ctx := context.WithValue(context.Background(), internal.KeyApp, app)
	require.NoError(t, app.Content.Init(ctx))
	app.Main.AddPage("main", tview.NewBox(), true, true)
	app.SetRunning(true)

	component := &traceTestComponent{
		Box: tview.NewBox(),
		meta: perftrace.Event{
			ViewName:  "Pod",
			GVR:       "v1/pods",
			Namespace: "big",
			Path:      "big/pod-a",
		},
		pending: &perftrace.Event{
			SelectedPath: "big/pod-a",
			RowsTotal:    100,
			RowsVisible:  10,
		},
	}
	app.Content.Push(component)

	app.recordAfterDraw()
	require.NoError(t, trace.Close(nil))

	events := readTraceEvents(t, tracePath)
	markers := lifecycleMarkers(events)
	require.Equal(t, []string{
		perftrace.MarkerViewActivate,
		perftrace.MarkerFirstRenderCommitted,
		perftrace.MarkerFirstUsefulRow,
	}, markers)
}

func TestMarkDetailOpenStartEmitsActionBoundaryMarker(t *testing.T) {
	tracePath := filepath.Join(t.TempDir(), "trace.jsonl")
	trace, err := perftrace.NewSession(perftrace.Options{File: tracePath})
	require.NoError(t, err)

	flags := genericclioptions.NewConfigFlags(false)
	kcfg := client.NewConfig(flags)
	kcfg.SetPerfTrace(trace)

	cfg := mock.NewMockConfig(t)
	cfg.SetConnection(traceConnection{cfg: kcfg})

	app := NewApp(cfg)
	markDetailOpenStart(app, "logs", client.PodGVR, "big/pod-a")
	require.NoError(t, trace.Close(nil))

	events := readTraceEvents(t, tracePath)
	require.Len(t, events, 2)
	assert.Equal(t, perftrace.EventLifecycleMark, events[0].Type)
	assert.Equal(t, perftrace.MarkerDetailOpenStart, events[0].Marker)
	assert.Equal(t, "logs", events[0].DetailKind)
	assert.Equal(t, "big/pod-a", events[0].Path)
}

type traceConnection struct {
	cfg *client.Config
}

func (c traceConnection) CanI(string, *client.GVR, string, []string) (bool, error) { return true, nil }
func (c traceConnection) Config() *client.Config                                   { return c.cfg }
func (traceConnection) ConnectionOK() bool                                         { return false }
func (traceConnection) Dial() (kubernetes.Interface, error)                        { return nil, nil }
func (traceConnection) DialLogs() (kubernetes.Interface, error)                    { return nil, nil }
func (traceConnection) SwitchContext(string) error                                 { return nil }
func (traceConnection) CachedDiscovery() (*disk.CachedDiscoveryClient, error)      { return nil, nil }
func (traceConnection) RestConfig() (*restclient.Config, error)                    { return nil, nil }
func (traceConnection) MXDial() (*versioned.Clientset, error)                      { return nil, nil }
func (traceConnection) DynDial() (dynamic.Interface, error)                        { return nil, nil }
func (traceConnection) HasMetrics() bool                                           { return false }
func (traceConnection) ValidNamespaceNames() (client.NamespaceNames, error)        { return nil, nil }
func (traceConnection) IsValidNamespace(string) bool                               { return true }
func (traceConnection) ServerVersion() (*version.Info, error)                      { return nil, nil }
func (traceConnection) CheckConnectivity() bool                                    { return false }
func (traceConnection) ActiveContext() string                                      { return "" }
func (traceConnection) ActiveNamespace() string                                    { return "" }
func (traceConnection) IsActiveNamespace(string) bool                              { return false }

type traceTestComponent struct {
	*tview.Box
	meta    perftrace.Event
	viewSeq int64
	pending *perftrace.Event
}

func (*traceTestComponent) Init(context.Context) error             { return nil }
func (*traceTestComponent) Start()                                 {}
func (*traceTestComponent) Stop()                                  {}
func (*traceTestComponent) Hints() model.MenuHints                 { return nil }
func (*traceTestComponent) ExtraHints() map[string]string          { return nil }
func (*traceTestComponent) InCmdMode() bool                        { return false }
func (*traceTestComponent) SetFilter(string, bool)                 {}
func (*traceTestComponent) SetLabelSelector(labels.Selector, bool) {}
func (*traceTestComponent) SetCommand(*viewcmd.Interpreter)        {}
func (c *traceTestComponent) Name() string                         { return c.meta.ViewName }
func (c *traceTestComponent) SetViewSeq(seq int64)                 { c.viewSeq = seq }
func (c *traceTestComponent) ViewSeq() int64                       { return c.viewSeq }
func (c *traceTestComponent) TraceViewMeta() perftrace.Event       { return c.meta }
func (c *traceTestComponent) ConsumePendingUsefulRow() (perftrace.Event, bool) {
	if c.pending == nil {
		return perftrace.Event{}, false
	}
	ev := *c.pending
	c.pending = nil
	return ev, true
}

func readTraceEvents(t *testing.T, path string) []perftrace.Event {
	t.Helper()

	bb, err := os.ReadFile(path)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(bb)), "\n")
	events := make([]perftrace.Event, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var ev perftrace.Event
		require.NoError(t, json.Unmarshal([]byte(line), &ev))
		events = append(events, ev)
	}

	return events
}

func lifecycleMarkers(events []perftrace.Event) []string {
	markers := make([]string, 0, len(events))
	for _, ev := range events {
		if ev.Type == perftrace.EventLifecycleMark {
			markers = append(markers, ev.Marker)
		}
	}

	return markers
}
