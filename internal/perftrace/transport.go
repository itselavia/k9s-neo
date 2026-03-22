// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package perftrace

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const maxInspectBytes = 8 << 20

type traceTransport struct {
	base    http.RoundTripper
	session *Session
	role    string
}

func (t *traceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	info := classifyRequest(http.MethodGet, nil)
	if req != nil {
		info = classifyRequest(req.Method, req.URL)
	}
	reqID := t.session.nextRequestID()

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		t.session.Emit(requestEvent(EventKubeRequestComplete, reqID, t.role, info, start, 0, 0, 0, "", err, "", 0, false))
		return nil, err
	}
	if resp == nil {
		t.session.Emit(requestEvent(EventKubeRequestComplete, reqID, t.role, info, start, 0, 0, 0, "", io.ErrUnexpectedEOF, "", 0, false))
		return resp, nil
	}

	contentType := resp.Header.Get("Content-Type")
	if info.streaming {
		t.session.Emit(requestEvent(EventKubeStreamOpen, reqID, t.role, info, start, resp.StatusCode, 0, resp.ContentLength, contentType, nil, "", 0, false))
		resp.Body = newTracingBody(resp.Body, t.session, reqID, t.role, info, start, resp.StatusCode, resp.ContentLength, contentType, true)
		return resp, nil
	}

	if resp.Body == nil {
		t.session.Emit(requestEvent(EventKubeRequestComplete, reqID, t.role, info, start, resp.StatusCode, 0, resp.ContentLength, contentType, nil, "", 0, false))
		return resp, nil
	}

	resp.Body = newTracingBody(resp.Body, t.session, reqID, t.role, info, start, resp.StatusCode, resp.ContentLength, contentType, false)

	return resp, nil
}

type tracingBody struct {
	base          io.ReadCloser
	session       *Session
	requestID     string
	role          string
	info          requestInfo
	start         time.Time
	statusCode    int
	contentLength int64
	contentType   string
	streaming     bool

	once        sync.Once
	bytes       int64
	buf         bytes.Buffer
	collectBody bool
	overflow    bool
	sawEOF      bool
}

func newTracingBody(
	base io.ReadCloser,
	session *Session,
	requestID, role string,
	info requestInfo,
	start time.Time,
	statusCode int,
	contentLength int64,
	contentType string,
	streaming bool,
) io.ReadCloser {
	return &tracingBody{
		base:          base,
		session:       session,
		requestID:     requestID,
		role:          role,
		info:          info,
		start:         start,
		statusCode:    statusCode,
		contentLength: contentLength,
		contentType:   contentType,
		streaming:     streaming,
		collectBody:   !streaming && strings.Contains(strings.ToLower(contentType), "json"),
	}
}

func (b *tracingBody) Read(p []byte) (int, error) {
	n, err := b.base.Read(p)
	if n > 0 {
		b.bytes += int64(n)
		if b.collectBody && !b.overflow {
			remaining := maxInspectBytes - b.buf.Len()
			if remaining > 0 {
				if n > remaining {
					_, _ = b.buf.Write(p[:remaining])
				} else {
					_, _ = b.buf.Write(p[:n])
				}
			}
			if b.bytes > maxInspectBytes {
				b.overflow = true
			}
		}
	}
	if err == io.EOF {
		b.sawEOF = true
		b.finalize(nil)
	}

	return n, err
}

func (b *tracingBody) Close() error {
	err := b.base.Close()
	b.finalize(err)
	return err
}

func (b *tracingBody) finalize(err error) {
	b.once.Do(func() {
		kind, itemCount, inspected := b.inspectBody()
		eventType := EventKubeRequestComplete
		if b.streaming {
			eventType = EventKubeStreamClose
		}
		b.session.Emit(requestEvent(
			eventType,
			b.requestID,
			b.role,
			b.info,
			b.start,
			b.statusCode,
			b.bytes,
			b.contentLength,
			b.contentType,
			err,
			kind,
			itemCount,
			inspected,
		))
	})
}

func (b *tracingBody) inspectBody() (string, int, bool) {
	if !b.collectBody || b.overflow || !b.sawEOF {
		return "", 0, false
	}

	var doc struct {
		Kind  string            `json:"kind"`
		Items []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(b.buf.Bytes(), &doc); err != nil {
		return "", 0, false
	}

	return doc.Kind, len(doc.Items), true
}

func requestEvent(
	eventType, requestID, role string,
	info requestInfo,
	start time.Time,
	statusCode int,
	responseBytes int64,
	contentLength int64,
	contentType string,
	err error,
	responseKind string,
	itemCount int,
	bodyInspected bool,
) Event {
	ev := Event{
		Type:          eventType,
		RequestID:     requestID,
		ClientRole:    role,
		HTTPMethod:    info.httpMethod,
		KubeVerb:      info.kubeVerb,
		Host:          info.host,
		Path:          info.path,
		Query:         info.query,
		APIGroup:      info.apiGroup,
		APIVersion:    info.apiVersion,
		Resource:      info.resource,
		Subresource:   info.subresource,
		Namespace:     info.namespace,
		Name:          info.name,
		StatusCode:    statusCode,
		DurationMS:    durationMS(time.Since(start)),
		ResponseBytes: responseBytes,
		ContentLength: contentLength,
		ContentType:   contentType,
		Streaming:     info.streaming,
		Watch:         info.watch,
		Follow:        info.follow,
	}
	if responseKind != "" {
		ev.ResponseKind = responseKind
	}
	if bodyInspected {
		ev.BodyInspected = true
		ev.ItemCount = itemCount
	}
	if err != nil && err != io.EOF {
		ev.Error = err.Error()
	}

	return ev
}
