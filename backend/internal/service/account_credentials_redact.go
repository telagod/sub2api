package service

// SensitiveCredentialKeys enumerates credential map keys that must never be
// exposed to the frontend. Both DTO-layer redaction and service-layer update
// merging reference this list.
var SensitiveCredentialKeys = []string{
	"access_token", "refresh_token", "id_token",
	"api_key", "session_key", "cookie",
	"aws_secret_access_key", "aws_session_token",
	"service_account_json", "service_account", "private_key",
}

var sensitiveCredentialKeySet = func() map[string]struct{} {
	lookup := make(map[string]struct{}, len(SensitiveCredentialKeys))
	for _, name := range SensitiveCredentialKeys {
		lookup[name] = struct{}{}
	}
	return lookup
}()

// IsSensitiveCredentialKey reports whether the given key is a sensitive
// credential sub-key.
func IsSensitiveCredentialKey(key string) bool {
	_, found := sensitiveCredentialKeySet[key]
	return found
}

// MergePreservingSensitiveCreds overlays incoming onto existing while
// preserving sensitive keys that were not explicitly provided in incoming.
// Returns a new map without modifying inputs.
func MergePreservingSensitiveCreds(existing, incoming map[string]any) map[string]any {
	merged := make(map[string]any, len(incoming)+len(SensitiveCredentialKeys))
	for k, v := range incoming {
		merged[k] = v
	}
	for _, sensitiveKey := range SensitiveCredentialKeys {
		if _, provided := incoming[sensitiveKey]; provided {
			continue
		}
		if prev, hasPrev := existing[sensitiveKey]; hasPrev {
			merged[sensitiveKey] = prev
		}
	}
	return merged
}
