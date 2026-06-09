// Package dto provides data transfer objects for HTTP handlers.
package dto

import "github.com/telagod/subme/internal/service"

// RedactCredentials returns a copy of in with all sensitive credential keys
// removed, plus a status map indicating which sensitive keys were present
// with non-zero values. Returns (nil, nil) when in is nil.
func RedactCredentials(in map[string]any) (out map[string]any, status map[string]bool) {
	if in == nil {
		return nil, nil
	}

	out = make(map[string]any, len(in))
	for field, val := range in {
		if service.IsSensitiveCredentialKey(field) {
			if credentialHasValue(val) {
				if status == nil {
					status = make(map[string]bool, 4)
				}
				status["has_"+field] = true
			}
			continue
		}
		out[field] = val
	}
	return out, status
}

// credentialHasValue reports whether v is considered a non-empty credential.
// nil, empty string, and false are treated as absent.
func credentialHasValue(v any) bool {
	if v == nil {
		return false
	}
	switch typed := v.(type) {
	case string:
		return len(typed) > 0
	case bool:
		return typed
	default:
		return true
	}
}
