// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/perftrace"
	"github.com/derailed/tcell/v2"
)

type tracedComponent interface {
	model.Component
	SetViewSeq(int64)
	ViewSeq() int64
	TraceViewMeta() perftrace.Event
}

type usefulRowTraceComponent interface {
	tracedComponent
	ConsumePendingUsefulRow() (perftrace.Event, bool)
}

func (a *App) perfTrace() *perftrace.Session {
	if a == nil || a.Conn() == nil || a.Conn().Config() == nil {
		return nil
	}

	return a.Conn().Config().PerfTrace()
}

func (a *App) activateComponentTrace(c model.Component) {
	trace := a.perfTrace()
	tc, ok := c.(tracedComponent)
	if trace == nil || !ok {
		return
	}

	meta := tc.TraceViewMeta()
	seq := trace.ActivateView(meta.ViewName, meta.GVR, meta.Namespace, meta.Path)
	tc.SetViewSeq(seq)
}

func currentTracedComponent(a *App) tracedComponent {
	if a == nil || a.Content == nil {
		return nil
	}
	top := a.Content.Top()
	tc, ok := top.(tracedComponent)
	if !ok {
		return nil
	}

	return tc
}

func keyName(evt *tcell.EventKey) string {
	if evt == nil {
		return ""
	}
	if name := evt.Name(); name != "" {
		return name
	}
	if r := evt.Rune(); r != 0 {
		return string(r)
	}

	return ""
}

func detailKind(title string) string {
	return strings.ToLower(strings.TrimSpace(title))
}

func (a *App) recordAfterDraw() {
	if !a.IsRunning() {
		return
	}
	if page, _ := a.Main.GetFrontPage(); page != "main" {
		return
	}
	trace := a.perfTrace()
	tc := currentTracedComponent(a)
	if trace == nil || tc == nil || tc.ViewSeq() == 0 {
		return
	}
	trace.MarkViewOnce(tc.ViewSeq(), perftrace.MarkerFirstRenderCommitted, tc.TraceViewMeta())
	if uc, ok := tc.(usefulRowTraceComponent); ok {
		if ev, ok := uc.ConsumePendingUsefulRow(); ok {
			trace.MarkViewOnce(tc.ViewSeq(), perftrace.MarkerFirstUsefulRow, ev)
		}
	}
}

func markDetailOpenStart(app *App, kind string, gvr *client.GVR, path string) {
	if app == nil {
		return
	}
	trace := app.perfTrace()
	if trace == nil {
		return
	}
	ns, _ := client.Namespaced(path)
	ev := perftrace.Event{
		DetailKind: kind,
		Namespace:  ns,
		Path:       path,
	}
	if gvr != nil {
		ev.GVR = gvr.String()
	}
	trace.Mark(perftrace.MarkerDetailOpenStart, ev)
}

func hasLogContent(lines [][]byte) bool {
	for _, line := range lines {
		if strings.TrimSpace(string(line)) != "" {
			return true
		}
	}

	return false
}
