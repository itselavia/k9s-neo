// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package perftrace

import (
	"net/http"
	"net/url"
	"strings"
)

type requestInfo struct {
	httpMethod  string
	kubeVerb    string
	host        string
	path        string
	query       string
	apiGroup    string
	apiVersion  string
	resource    string
	subresource string
	namespace   string
	name        string
	watch       bool
	follow      bool
	streaming   bool
}

func classifyRequest(method string, u *url.URL) requestInfo {
	info := requestInfo{
		httpMethod: method,
	}
	if u == nil {
		info.kubeVerb = classifyVerb(method, false, false)
		return info
	}

	info.host = u.Host
	info.path = u.Path
	info.query = u.RawQuery
	info.watch = truthy(u.Query().Get("watch"))
	info.follow = truthy(u.Query().Get("follow"))

	path := strings.Trim(u.Path, "/")
	if path == "" {
		info.kubeVerb = classifyVerb(method, false, false)
		return info
	}
	segments := strings.Split(path, "/")
	if len(segments) == 0 {
		info.kubeVerb = classifyVerb(method, false, false)
		return info
	}

	switch segments[0] {
	case "version":
		info.kubeVerb = classifyVerb(method, false, false)
		return info
	case "openapi":
		info.kubeVerb = classifyVerb(method, false, false)
		return info
	case "api":
		if len(segments) == 1 {
			info.kubeVerb = classifyVerb(method, false, false)
			return info
		}
		info.apiVersion = segments[1]
		parseResourcePath(&info, segments[2:])
	case "apis":
		switch len(segments) {
		case 1, 2:
			info.kubeVerb = classifyVerb(method, false, false)
			if len(segments) == 2 {
				info.apiGroup = segments[1]
			}
			return info
		default:
			info.apiGroup = segments[1]
			info.apiVersion = segments[2]
			parseResourcePath(&info, segments[3:])
		}
	default:
		info.kubeVerb = classifyVerb(method, false, false)
		return info
	}

	info.streaming = info.watch || (info.resource == "pods" && info.subresource == "log" && info.follow)
	info.kubeVerb = classifyVerb(method, info.watch, info.name == "")
	return info
}

func parseResourcePath(info *requestInfo, segments []string) {
	if len(segments) == 0 {
		return
	}

	if len(segments) >= 2 && segments[0] == "namespaces" {
		info.namespace = segments[1]
		segments = segments[2:]
	}
	if len(segments) == 0 {
		return
	}

	info.resource = segments[0]
	if len(segments) > 1 {
		info.name = segments[1]
	}
	if len(segments) > 2 {
		info.subresource = segments[2]
	}
}

func classifyVerb(method string, watch, collection bool) string {
	switch method {
	case http.MethodGet:
		switch {
		case watch:
			return "watch"
		case collection:
			return "list"
		default:
			return "get"
		}
	case http.MethodPost:
		return "create"
	case http.MethodPatch:
		return "patch"
	case http.MethodPut:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return strings.ToLower(method)
	}
}

func truthy(v string) bool {
	switch strings.ToLower(v) {
	case "1", "t", "true", "y", "yes":
		return true
	default:
		return false
	}
}
