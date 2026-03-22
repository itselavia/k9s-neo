// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package perftrace

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Session represents a runtime perf trace session.
type Session struct {
	opts     Options
	file     *os.File
	writer   *bufio.Writer
	mu       sync.Mutex
	enabled  bool
	closed   bool
	failOnce sync.Once
	seq      atomic.Int64
	reqSeq   atomic.Int64
}

// NewSession returns a new perf trace session.
func NewSession(opts Options) (*Session, error) {
	s := &Session{opts: opts}
	if opts.File == "" {
		return s, nil
	}

	dir := filepath.Dir(opts.File)
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(opts.File, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}
	s.file = f
	s.writer = bufio.NewWriter(f)
	s.enabled = true

	return s, nil
}

// Enabled returns true if the session should emit events.
func (s *Session) Enabled() bool {
	if s == nil {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.enabled && !s.closed
}

// Emit appends one perf trace event to the output stream.
func (s *Session) Emit(ev Event) {
	if s == nil || !s.Enabled() {
		return
	}

	ev.SchemaVersion = schemaVersion
	ev.Seq = s.seq.Add(1)
	ev.TS = time.Now().UTC()
	if ev.RunID == "" {
		ev.RunID = s.opts.RunID
	}
	if ev.Scenario == "" {
		ev.Scenario = s.opts.Scenario
	}

	payload, err := json.Marshal(ev)
	if err != nil {
		s.disableOnFailure(fmt.Errorf("marshal perf trace event: %w", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.enabled || s.closed || s.writer == nil {
		return
	}
	if _, err := s.writer.Write(payload); err != nil {
		s.disableLocked(err)
		return
	}
	if err := s.writer.WriteByte('\n'); err != nil {
		s.disableLocked(err)
		return
	}
	if err := s.writer.Flush(); err != nil {
		s.disableLocked(err)
	}
}

// WrapTransport wraps an HTTP transport with perf tracing.
func (s *Session) WrapTransport(role string) func(http.RoundTripper) http.RoundTripper {
	return func(rt http.RoundTripper) http.RoundTripper {
		if rt == nil {
			rt = http.DefaultTransport
		}
		if s == nil || !s.Enabled() {
			return rt
		}

		return &traceTransport{
			base:    rt,
			session: s,
			role:    role,
		}
	}
}

// Close emits the final session event and closes the underlying file.
func (s *Session) Close(exitErr error) error {
	if s == nil {
		return nil
	}
	if s.Enabled() {
		ev := Event{Type: EventSessionEnd}
		if exitErr != nil {
			ev.Error = exitErr.Error()
		}
		s.Emit(ev)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true

	var err error
	if s.writer != nil {
		err = s.writer.Flush()
	}
	if s.file != nil {
		if closeErr := s.file.Close(); err == nil {
			err = closeErr
		}
	}

	return err
}

func (s *Session) nextRequestID() string {
	return strconv.FormatInt(s.reqSeq.Add(1), 10)
}

func (s *Session) disableOnFailure(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disableLocked(err)
}

func (s *Session) disableLocked(err error) {
	s.enabled = false
	s.failOnce.Do(func() {
		slog.Warn("Disabling perf trace after write failure", "error", err)
	})
}
