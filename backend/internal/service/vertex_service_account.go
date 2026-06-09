package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/telagod/subme/internal/pkg/proxyurl"
	"github.com/telagod/subme/internal/pkg/proxyutil"
)

const (
	vertexDefaultLocation         = "us-central1"
	vertexDefaultTokenURL         = "https://oauth2.googleapis.com/token"
	vertexCloudPlatformScope      = "https://www.googleapis.com/auth/cloud-platform"
	vertexServiceAccountCacheSkew = 5 * time.Minute
	vertexLockWaitTime            = 200 * time.Millisecond
	vertexAnthropicVersion        = "vertex-2023-10-16"
)

var (
	vertexLocationPattern                = regexp.MustCompile(`^[a-z0-9-]+$`)
	vertexAnthropicDatedModelIDPattern   = regexp.MustCompile(`^(.+)-([0-9]{8})$`)
	vertexAnthropicAlreadyDatedIDPattern = regexp.MustCompile(`^.+@[0-9]{8}$`)
)

// vertexServiceAccountKey holds the parsed fields from a GCP service account JSON.
type vertexServiceAccountKey struct {
	Type         string `json:"type"`
	ProjectID    string `json:"project_id"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ClientEmail  string `json:"client_email"`
	TokenURI     string `json:"token_uri"`
}

// vertexTokenResponse is the JSON structure returned by the Google OAuth2 token endpoint.
type vertexTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

func (a *Account) IsVertexServiceAccount() bool {
	return a != nil && a.Type == AccountTypeServiceAccount
}

func (a *Account) VertexProjectID() string {
	if a == nil {
		return ""
	}
	if proj := strings.TrimSpace(a.GetCredential("project_id")); proj != "" {
		return proj
	}
	parsed, parseErr := decodeServiceAccountKey(a)
	if parseErr == nil {
		return strings.TrimSpace(parsed.ProjectID)
	}
	return ""
}

func (a *Account) VertexLocation(model string) string {
	if a == nil {
		return vertexDefaultLocation
	}
	if model != "" && a.Credentials != nil {
		if locMap, ok := a.Credentials["vertex_model_locations"].(map[string]any); ok {
			if regionStr, ok := locMap[model].(string); ok && strings.TrimSpace(regionStr) != "" {
				return strings.TrimSpace(regionStr)
			}
		}
	}
	if loc := strings.TrimSpace(a.GetCredential("location")); loc != "" {
		return loc
	}
	if loc := strings.TrimSpace(a.GetCredential("vertex_location")); loc != "" {
		return loc
	}
	return vertexDefaultLocation
}

// decodeServiceAccountKey extracts and parses the service account JSON from an account's credentials.
func decodeServiceAccountKey(acct *Account) (*vertexServiceAccountKey, error) {
	if acct == nil || acct.Credentials == nil {
		return nil, errors.New("service account credentials are not configured")
	}

	if blob := strings.TrimSpace(acct.GetCredential("service_account_json")); blob != "" {
		return unmarshalServiceAccountJSON([]byte(blob))
	}
	if blob := strings.TrimSpace(acct.GetCredential("service_account")); blob != "" {
		return unmarshalServiceAccountJSON([]byte(blob))
	}
	if nested, ok := acct.Credentials["service_account_json"].(map[string]any); ok {
		encoded, _ := json.Marshal(nested)
		return unmarshalServiceAccountJSON(encoded)
	}
	if nested, ok := acct.Credentials["service_account"].(map[string]any); ok {
		encoded, _ := json.Marshal(nested)
		return unmarshalServiceAccountJSON(encoded)
	}
	return nil, errors.New("service_account_json field not found in credentials")
}

// unmarshalServiceAccountJSON parses and validates a raw service account JSON blob.
func unmarshalServiceAccountJSON(blob []byte) (*vertexServiceAccountKey, error) {
	var saKey vertexServiceAccountKey
	if decodeErr := json.Unmarshal(blob, &saKey); decodeErr != nil {
		return nil, fmt.Errorf("malformed service account JSON: %w", decodeErr)
	}
	if strings.TrimSpace(saKey.ClientEmail) == "" {
		return nil, errors.New("service account JSON is missing client_email")
	}
	if strings.TrimSpace(saKey.PrivateKey) == "" {
		return nil, errors.New("service account JSON is missing private_key")
	}
	if strings.TrimSpace(saKey.ProjectID) == "" {
		return nil, errors.New("service account JSON is missing project_id")
	}
	// Always use the well-known Google token endpoint to prevent SSRF via crafted token_uri.
	saKey.TokenURI = vertexDefaultTokenURL
	return &saKey, nil
}

// computeServiceAccountCacheKey builds a cache key for a service account's access token.
func computeServiceAccountCacheKey(acct *Account, saKey *vertexServiceAccountKey) string {
	fingerprint := ""
	if saKey != nil {
		digest := sha256.Sum256([]byte(saKey.ClientEmail + "\x00" + saKey.PrivateKeyID))
		fingerprint = hex.EncodeToString(digest[:8])
	}
	if fingerprint == "" && acct != nil {
		fingerprint = fmt.Sprintf("account:%d", acct.ID)
	}
	return "vertex:service_account:" + fingerprint
}

// getVertexServiceAccountAccessToken obtains an access token for a Vertex service account,
// using the shared cache and distributed lock to avoid redundant exchanges.
func getVertexServiceAccountAccessToken(ctx context.Context, cache GeminiTokenCache, acct *Account) (string, error) {
	saKey, parseErr := decodeServiceAccountKey(acct)
	if parseErr != nil {
		return "", parseErr
	}
	cacheID := computeServiceAccountCacheKey(acct, saKey)

	if cache != nil {
		if cachedToken, lookupErr := cache.GetAccessToken(ctx, cacheID); lookupErr == nil && strings.TrimSpace(cachedToken) != "" {
			return cachedToken, nil
		}
	}

	acquired := false
	if cache != nil {
		var lockErr error
		acquired, lockErr = cache.AcquireRefreshLock(ctx, cacheID, 30*time.Second)
		if lockErr == nil && acquired {
			defer func() { _ = cache.ReleaseRefreshLock(ctx, cacheID) }()
		} else if lockErr != nil {
			slog.Warn("vertex_service_account_token_lock_failed", "account_id", acct.ID, "error", lockErr)
		} else {
			time.Sleep(vertexLockWaitTime)
			if cachedToken, lookupErr := cache.GetAccessToken(ctx, cacheID); lookupErr == nil && strings.TrimSpace(cachedToken) != "" {
				return cachedToken, nil
			}
		}
	}

	freshToken, ttl, exchangeErr := performTokenExchange(ctx, saKey, resolveServiceAccountProxy(acct))
	if exchangeErr != nil {
		return "", exchangeErr
	}
	if cache != nil {
		_ = cache.SetAccessToken(ctx, cacheID, freshToken, ttl)
	}
	return freshToken, nil
}

// resolveServiceAccountProxy returns the proxy URL for a service account, if any.
func resolveServiceAccountProxy(acct *Account) string {
	if acct == nil || acct.ProxyID == nil || acct.Proxy == nil {
		return ""
	}
	return acct.Proxy.URL()
}

// buildProxiedHTTPClient creates an HTTP client optionally configured with a proxy.
func buildProxiedHTTPClient(proxyAddr string) (*http.Client, error) {
	proxyAddr = strings.TrimSpace(proxyAddr)
	if proxyAddr == "" {
		return &http.Client{Timeout: 15 * time.Second}, nil
	}

	_, resolvedProxy, parseErr := proxyurl.Parse(proxyAddr)
	if parseErr != nil {
		return nil, parseErr
	}
	baseTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("unexpected default transport type: %T", http.DefaultTransport)
	}
	clonedTransport := baseTransport.Clone()
	clonedTransport.Proxy = nil
	if cfgErr := proxyutil.ConfigureTransportProxy(clonedTransport, resolvedProxy); cfgErr != nil {
		return nil, cfgErr
	}
	return &http.Client{Timeout: 15 * time.Second, Transport: clonedTransport}, nil
}

// performTokenExchange creates a signed JWT and exchanges it for an access token.
func performTokenExchange(ctx context.Context, saKey *vertexServiceAccountKey, proxyAddr string) (string, time.Duration, error) {
	currentTime := time.Now()
	jwtClaims := jwt.MapClaims{
		"iss":   saKey.ClientEmail,
		"scope": vertexCloudPlatformScope,
		"aud":   saKey.TokenURI,
		"iat":   currentTime.Unix(),
		"exp":   currentTime.Add(time.Hour).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims)
	if strings.TrimSpace(saKey.PrivateKeyID) != "" {
		jwtToken.Header["kid"] = saKey.PrivateKeyID
	}
	rsaKey, keyErr := jwt.ParseRSAPrivateKeyFromPEM([]byte(saKey.PrivateKey))
	if keyErr != nil {
		return "", 0, fmt.Errorf("failed to parse service account private key: %w", keyErr)
	}
	signedAssertion, signErr := jwtToken.SignedString(rsaKey)
	if signErr != nil {
		return "", 0, fmt.Errorf("failed to sign service account JWT assertion: %w", signErr)
	}

	formData := url.Values{}
	formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	formData.Set("assertion", signedAssertion)

	httpReq, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, saKey.TokenURI, strings.NewReader(formData.Encode()))
	if reqErr != nil {
		return "", 0, reqErr
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient, clientErr := buildProxiedHTTPClient(proxyAddr)
	if clientErr != nil {
		return "", 0, fmt.Errorf("failed to configure token exchange proxy: %w", clientErr)
	}
	httpResp, doErr := httpClient.Do(httpReq)
	if doErr != nil {
		return "", 0, fmt.Errorf("token exchange request failed: %w", doErr)
	}
	defer func() { _ = httpResp.Body.Close() }()

	respBytes, _ := io.ReadAll(io.LimitReader(httpResp.Body, 1<<20))
	var tokenResp vertexTokenResponse
	_ = json.Unmarshal(respBytes, &tokenResp)
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		detail := strings.TrimSpace(tokenResp.ErrorDesc)
		if detail == "" {
			detail = strings.TrimSpace(tokenResp.Error)
		}
		if detail == "" {
			detail = string(bytes.TrimSpace(respBytes))
		}
		return "", 0, fmt.Errorf("token exchange returned HTTP %d: %s", httpResp.StatusCode, detail)
	}
	if strings.TrimSpace(tokenResp.AccessToken) == "" {
		return "", 0, errors.New("token exchange response is missing the access_token field")
	}
	tokenTTL := time.Duration(tokenResp.ExpiresIn) * time.Second
	if tokenTTL <= 0 {
		tokenTTL = time.Hour
	}
	if tokenTTL > vertexServiceAccountCacheSkew {
		tokenTTL -= vertexServiceAccountCacheSkew
	}
	return tokenResp.AccessToken, tokenTTL, nil
}

func buildVertexGeminiURL(projectID, location, model, action string, stream bool) (string, error) {
	projectID = strings.TrimSpace(projectID)
	location = strings.TrimSpace(location)
	model = strings.TrimSpace(model)
	action = strings.TrimSpace(action)
	if projectID == "" {
		return "", errors.New("vertex project_id is required")
	}
	if location == "" {
		location = vertexDefaultLocation
	}
	if !vertexLocationPattern.MatchString(location) {
		return "", fmt.Errorf("invalid vertex location: %s", location)
	}
	if model == "" {
		return "", errors.New("vertex model is required")
	}
	switch action {
	case "generateContent", "streamGenerateContent", "countTokens":
	default:
		return "", fmt.Errorf("unsupported vertex gemini action: %s", action)
	}
	hostname := fmt.Sprintf("%s-aiplatform.googleapis.com", location)
	if location == "global" {
		hostname = "aiplatform.googleapis.com"
	}
	endpoint := fmt.Sprintf(
		"https://%s/v1/projects/%s/locations/%s/publishers/google/models/%s:%s",
		hostname,
		url.PathEscape(projectID),
		url.PathEscape(location),
		url.PathEscape(model),
		action,
	)
	if stream {
		endpoint += "?alt=sse"
	}
	return endpoint, nil
}

func buildVertexAnthropicURL(projectID, location, model string, stream bool) (string, error) {
	projectID = strings.TrimSpace(projectID)
	location = strings.TrimSpace(location)
	model = strings.TrimSpace(model)
	if projectID == "" {
		return "", errors.New("vertex project_id is required")
	}
	if location == "" {
		location = vertexDefaultLocation
	}
	if !vertexLocationPattern.MatchString(location) {
		return "", fmt.Errorf("invalid vertex location: %s", location)
	}
	if model == "" {
		return "", errors.New("vertex model is required")
	}
	verb := "rawPredict"
	if stream {
		verb = "streamRawPredict"
	}
	hostname := fmt.Sprintf("%s-aiplatform.googleapis.com", location)
	if location == "global" {
		hostname = "aiplatform.googleapis.com"
	}
	escapedModel := strings.ReplaceAll(url.PathEscape(model), "%40", "@")
	return fmt.Sprintf(
		"https://%s/v1/projects/%s/locations/%s/publishers/anthropic/models/%s:%s",
		hostname,
		url.PathEscape(projectID),
		url.PathEscape(location),
		escapedModel,
		verb,
	), nil
}

func normalizeVertexAnthropicModelID(model string) string {
	model = strings.TrimSpace(model)
	if model == "" || vertexAnthropicAlreadyDatedIDPattern.MatchString(model) {
		return model
	}
	if groups := vertexAnthropicDatedModelIDPattern.FindStringSubmatch(model); len(groups) == 3 {
		return groups[1] + "@" + groups[2]
	}
	return model
}

// parseVertexServiceAccountKey is a package-internal alias retained for callers
// outside this file (e.g. gemini_token_provider.go).
func parseVertexServiceAccountKey(acct *Account) (*vertexServiceAccountKey, error) {
	return decodeServiceAccountKey(acct)
}

// vertexServiceAccountCacheKey is a package-internal alias retained for callers
// outside this file (e.g. gemini_token_provider.go).
func vertexServiceAccountCacheKey(acct *Account, saKey *vertexServiceAccountKey) string {
	return computeServiceAccountCacheKey(acct, saKey)
}

func buildVertexAnthropicRequestBody(body []byte) ([]byte, error) {
	var fields map[string]any
	if decErr := json.Unmarshal(body, &fields); decErr != nil {
		return nil, fmt.Errorf("failed to parse anthropic vertex request body: %w", decErr)
	}
	delete(fields, "model")
	fields["anthropic_version"] = vertexAnthropicVersion
	return json.Marshal(fields)
}
