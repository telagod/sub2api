package service

import "time"

// UserErrorRequest is the sanitized, end-user-facing view of a failed request (allowlist only).
// Fields like client_ip, user_agent, account, api_key_prefix, upstream_endpoint,
// and user_email are intentionally excluded. The message field (standardized gateway
// error description) and key_name (user-owned API Key name, already visible in KeysView)
// are exposed by product decision; error_body is only returned through
// GetUserErrorRequestDetail after ownership verification.
type UserErrorRequest struct {
	ID              int64     `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	Model           string    `json:"model"`
	InboundEndpoint string    `json:"inbound_endpoint"`
	StatusCode      int       `json:"status_code"`
	Category        string    `json:"category"`
	Platform        string    `json:"platform"`
	Message         string    `json:"message"`
	KeyName         string    `json:"key_name"`
	KeyDeleted      bool      `json:"key_deleted"`
}

// UserErrorRequestList holds a paginated set of user error requests.
type UserErrorRequestList struct {
	Items    []*UserErrorRequest `json:"items"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// MapUserErrorCategory maps backend error_phase + error_type into a stable user-facing
// category code (for frontend i18n), not a display label.
func MapUserErrorCategory(phase, errKind string) string {
	if phase == "auth" {
		return "auth"
	}
	if phase == "routing" {
		return "service_unavailable"
	}
	if phase == "upstream" || phase == "network" {
		return "upstream"
	}
	if phase == "internal" {
		return "internal"
	}
	if phase == "request" {
		switch errKind {
		case "rate_limit_error":
			return "rate_limit"
		case "billing_error", "subscription_error":
			return "quota"
		case "invalid_request_error":
			return "invalid_request"
		}
	}
	return "other"
}

// CategoryToFilter reverse-maps a user-facing category code to backend filter criteria
// (plain ANY). Unknown categories return empty slices (no category filter applied).
// Note: "other" also returns empty slices because it has no precise phase/type combination.
func CategoryToFilter(cat string) (phases []string, errorTypes []string) {
	switch cat {
	case "auth":
		return []string{"auth"}, nil
	case "service_unavailable":
		return []string{"routing"}, nil
	case "upstream":
		return []string{"upstream", "network"}, nil
	case "internal":
		return []string{"internal"}, nil
	case "rate_limit":
		return nil, []string{"rate_limit_error"}
	case "quota":
		return nil, []string{"billing_error", "subscription_error"}
	case "invalid_request":
		return nil, []string{"invalid_request_error"}
	default:
		return nil, nil
	}
}

// ToUserErrorRequest converts an internal OpsErrorLog into the user-safe view.
func ToUserErrorRequest(src *OpsErrorLog) *UserErrorRequest {
	if src == nil {
		return nil
	}
	displayModel := src.RequestedModel
	if displayModel == "" {
		displayModel = src.Model
	}
	return &UserErrorRequest{
		ID:              src.ID,
		CreatedAt:       src.CreatedAt,
		Model:           displayModel,
		InboundEndpoint: src.InboundEndpoint,
		StatusCode:      src.StatusCode,
		Category:        MapUserErrorCategory(src.Phase, src.Type),
		Platform:        src.Platform,
		Message:         src.Message,
		KeyName:         src.APIKeyName,
		KeyDeleted:      src.APIKeyDeleted,
	}
}

// UserErrorRequestDetail extends UserErrorRequest with the upstream error body and
// status code, shown when a user clicks a single row. Internal/sensitive fields
// remain excluded.
type UserErrorRequestDetail struct {
	UserErrorRequest
	ErrorBody          string `json:"error_body"`
	UpstreamStatusCode *int   `json:"upstream_status_code,omitempty"`
}

// ToUserErrorRequestDetail converts an internal OpsErrorLogDetail into the
// user-safe detail view.
func ToUserErrorRequestDetail(src *OpsErrorLogDetail) *UserErrorRequestDetail {
	if src == nil {
		return nil
	}
	baseView := ToUserErrorRequest(&src.OpsErrorLog)
	return &UserErrorRequestDetail{
		UserErrorRequest:   *baseView,
		ErrorBody:          src.ErrorBody,
		UpstreamStatusCode: src.UpstreamStatusCode,
	}
}
