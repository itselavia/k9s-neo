// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"slices"
	"strings"

	"github.com/derailed/k9s/internal/client"
)

// ToLabels converts a string into a map of labels.
func ToLabels(s string) map[string]string {
	var (
		ll   = strings.Split(s, ",")
		lbls = make(map[string]string, len(ll))
	)
	for _, l := range ll {
		if k, v, ok := splitKv(l); ok {
			lbls[k] = v
		} else {
			continue
		}
	}
	if len(lbls) == 0 {
		return nil
	}

	return lbls
}

func splitKv(s string) (k, v string, ok bool) {
	switch {
	case strings.Contains(s, labelFlagNotEq):
		kv := strings.SplitN(s, labelFlagNotEq, 2)
		if len(kv) == 2 && kv[0] != "" && kv[1] != "" {
			return strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]), true
		}
	case strings.Contains(s, labelFlagEqs):
		kv := strings.SplitN(s, labelFlagEqs, 2)
		if len(kv) == 2 && kv[0] != "" && kv[1] != "" {
			return strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]), true
		}
	case strings.Contains(s, labelFlagEq):
		kv := strings.SplitN(s, labelFlagEq, 2)
		if len(kv) == 2 && kv[0] != "" && kv[1] != "" {
			return strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]), true
		}
	}

	return "", "", false
}

// ShouldAddSuggest checks if a suggestion match the given command.
func ShouldAddSuggest(command, suggest string) (string, bool) {
	if command != suggest && strings.HasPrefix(suggest, command) {
		return strings.TrimPrefix(suggest, command), true
	}

	return "", false
}

// ContextSuggestionArg returns the context token currently being completed.
func ContextSuggestionArg(command string) (string, bool) {
	p := NewInterpreter(command)
	return p.ContextArg()
}

// NamespaceSuggestionArg returns the namespace token currently being completed.
func NamespaceSuggestionArg(command string) (string, bool) {
	p := NewInterpreter(command)
	switch {
	case p.IsCowCmd(), p.IsHelpCmd(), p.IsAliasCmd(), p.IsBailCmd(), p.IsDirCmd(), p.IsContextCmd():
		return "", false
	case p.IsXrayCmd():
		_, ns, ok := p.XrayArgs()
		if !ok || ns == "" {
			return "", false
		}
		return ns, true
	case p.HasNS():
		if _, ok := p.HasContext(); ok {
			return "", false
		}
		return p.NSArg()
	default:
		return "", false
	}
}

// ShouldSuggestNamespace reports whether namespace completion should fetch namespace names.
func ShouldSuggestNamespace(command, activeNamespace string) bool {
	ns, ok := NamespaceSuggestionArg(command)
	if !ok {
		return false
	}

	return strings.ToLower(activeNamespace) != ns
}

// SuggestContextSuggestions returns context completions for the current command.
func SuggestContextSuggestions(command string, contexts []string) []string {
	n, ok := ContextSuggestionArg(command)
	if !ok {
		return nil
	}

	return completeCtx(command, n, contexts)
}

// SuggestNamespaceSuggestions returns namespace completions for the current command.
func SuggestNamespaceSuggestions(command string, namespaces client.NamespaceNames) []string {
	n, ok := NamespaceSuggestionArg(command)
	if !ok {
		return nil
	}

	return completeNS(n, namespaces)
}

// SuggestSubCommand suggests namespaces or contexts based on current command.
func SuggestSubCommand(command string, namespaces client.NamespaceNames, contexts []string) []string {
	var suggests []string
	suggests = SuggestContextSuggestions(command, contexts)
	if len(suggests) == 0 {
		suggests = SuggestNamespaceSuggestions(command, namespaces)
	}
	slices.Sort(suggests)

	return suggests
}

func completeNS(s string, nn client.NamespaceNames) []string {
	s = strings.ToLower(s)
	var suggests []string
	if suggest, ok := ShouldAddSuggest(s, client.NamespaceAll); ok {
		suggests = append(suggests, suggest)
	}
	for ns := range nn {
		if suggest, ok := ShouldAddSuggest(s, ns); ok {
			suggests = append(suggests, suggest)
		}
	}

	return suggests
}

func completeCtx(command, s string, contexts []string) []string {
	var suggests []string
	for _, ctxName := range contexts {
		if suggest, ok := ShouldAddSuggest(s, ctxName); ok {
			if s == "" && !strings.HasSuffix(command, " ") {
				suggests = append(suggests, " "+suggest)
				continue
			}
			suggests = append(suggests, suggest)
		}
	}

	return suggests
}
