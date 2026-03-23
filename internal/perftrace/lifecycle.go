// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package perftrace

import (
	"sync"
	"time"
)

var processStart = time.Now()

type viewState struct {
	meta Event
	seen map[string]struct{}
}

type lifecycleTracker struct {
	mu       sync.Mutex
	nextView int64
	views    map[int64]*viewState
}

func newLifecycleTracker() *lifecycleTracker {
	return &lifecycleTracker{
		views: make(map[int64]*viewState),
	}
}

func (t *lifecycleTracker) activate(meta Event) Event {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.nextView++
	meta.ViewSeq = t.nextView
	t.views[meta.ViewSeq] = &viewState{
		meta: baseViewMeta(meta),
		seen: make(map[string]struct{}),
	}

	return t.views[meta.ViewSeq].meta
}

func (t *lifecycleTracker) eventFor(seq int64, marker string, extra Event, once bool) (Event, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state, ok := t.views[seq]
	if !ok {
		return Event{}, false
	}
	if once {
		if _, dup := state.seen[marker]; dup {
			return Event{}, false
		}
		state.seen[marker] = struct{}{}
	}

	ev := state.meta
	ev.Type = EventLifecycleMark
	ev.Marker = marker
	mergeLifecycleEvent(&ev, extra)

	return ev, true
}

func (t *lifecycleTracker) markSeen(seq int64, marker string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if state, ok := t.views[seq]; ok {
		state.seen[marker] = struct{}{}
	}
}

func (t *lifecycleTracker) hasSeen(seq int64, marker string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	state, ok := t.views[seq]
	if !ok {
		return false
	}
	_, ok = state.seen[marker]

	return ok
}

func baseViewMeta(ev Event) Event {
	return Event{
		ViewSeq:   ev.ViewSeq,
		ViewName:  ev.ViewName,
		GVR:       ev.GVR,
		Namespace: ev.Namespace,
		Path:      ev.Path,
	}
}

func mergeLifecycleEvent(dst *Event, src Event) {
	if src.ViewSeq != 0 {
		dst.ViewSeq = src.ViewSeq
	}
	if src.ViewName != "" {
		dst.ViewName = src.ViewName
	}
	if src.GVR != "" {
		dst.GVR = src.GVR
	}
	if src.Namespace != "" {
		dst.Namespace = src.Namespace
	}
	if src.Path != "" {
		dst.Path = src.Path
	}
	if src.SelectedPath != "" {
		dst.SelectedPath = src.SelectedPath
	}
	if src.RowsTotal != 0 {
		dst.RowsTotal = src.RowsTotal
	}
	if src.RowsVisible != 0 {
		dst.RowsVisible = src.RowsVisible
	}
	if src.KeyName != "" {
		dst.KeyName = src.KeyName
	}
	if src.DetailKind != "" {
		dst.DetailKind = src.DetailKind
	}
	if src.CommandLine != "" {
		dst.CommandLine = src.CommandLine
	}
	if src.FilterText != "" {
		dst.FilterText = src.FilterText
	}
	if src.ObjectCount != 0 {
		dst.ObjectCount = src.ObjectCount
	}
}
