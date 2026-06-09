package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// ChannelMonitorRequestTemplateRepository defines the data access interface for monitor templates.
type ChannelMonitorRequestTemplateRepository interface {
	Create(ctx context.Context, t *ChannelMonitorRequestTemplate) error
	GetByID(ctx context.Context, id int64) (*ChannelMonitorRequestTemplate, error)
	Update(ctx context.Context, t *ChannelMonitorRequestTemplate) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params ChannelMonitorRequestTemplateListParams) ([]*ChannelMonitorRequestTemplate, error)
	// ApplyToMonitors batch-copies the template's api_mode / extra_headers / body_override_mode /
	// body_override onto the specified monitorIDs (only if their current template_id matches).
	// monitorIDs must be non-empty; empty list returns 0 without writing.
	// Returns the count of monitors actually updated.
	ApplyToMonitors(ctx context.Context, id int64, monitorIDs []int64) (int64, error)
	// CountAssociatedMonitors returns the number of monitors with template_id = id.
	CountAssociatedMonitors(ctx context.Context, id int64) (int64, error)
	// ListAssociatedMonitors returns brief info for all monitors associated with the template.
	ListAssociatedMonitors(ctx context.Context, id int64) ([]*AssociatedMonitorBrief, error)
}

// AssociatedMonitorBrief contains minimal monitor info for the template picker/list UI.
type AssociatedMonitorBrief struct {
	ID       int64
	Name     string
	Provider string
	APIMode  string
	Enabled  bool
}

// ChannelMonitorRequestTemplateService manages monitor request templates.
type ChannelMonitorRequestTemplateService struct {
	repo ChannelMonitorRequestTemplateRepository
}

// NewChannelMonitorRequestTemplateService creates a template service instance.
func NewChannelMonitorRequestTemplateService(repo ChannelMonitorRequestTemplateRepository) *ChannelMonitorRequestTemplateService {
	return &ChannelMonitorRequestTemplateService{repo: repo}
}

// ---------- CRUD ----------

// List returns templates filtered by provider (empty = all). No pagination (template volume is small).
func (s *ChannelMonitorRequestTemplateService) List(reqCtx context.Context, params ChannelMonitorRequestTemplateListParams) ([]*ChannelMonitorRequestTemplate, error) {
	if params.Provider != "" {
		if validateErr := validateProvider(params.Provider); validateErr != nil {
			return nil, validateErr
		}
	}
	if params.APIMode != "" {
		provForValidation := params.Provider
		if provForValidation == "" {
			provForValidation = MonitorProviderOpenAI
		}
		if validateErr := validateAPIMode(provForValidation, params.APIMode); validateErr != nil {
			return nil, validateErr
		}
	}
	return s.repo.List(reqCtx, params)
}

// Get returns a single template by ID.
func (s *ChannelMonitorRequestTemplateService) Get(reqCtx context.Context, templateID int64) (*ChannelMonitorRequestTemplate, error) {
	return s.repo.GetByID(reqCtx, templateID)
}

// Create creates a new template (validates header blocklist and body mode constraints).
func (s *ChannelMonitorRequestTemplateService) Create(reqCtx context.Context, params ChannelMonitorRequestTemplateCreateParams) (*ChannelMonitorRequestTemplate, error) {
	if validationErr := validateTemplateCreateParams(params); validationErr != nil {
		return nil, validationErr
	}
	tpl := &ChannelMonitorRequestTemplate{
		Name:             strings.TrimSpace(params.Name),
		Provider:         params.Provider,
		APIMode:          defaultAPIMode(params.APIMode),
		Description:      strings.TrimSpace(params.Description),
		ExtraHeaders:     emptyHeadersIfNil(params.ExtraHeaders),
		BodyOverrideMode: defaultBodyMode(params.BodyOverrideMode),
		BodyOverride:     params.BodyOverride,
	}
	if createErr := s.repo.Create(reqCtx, tpl); createErr != nil {
		return nil, fmt.Errorf("create template: %w", createErr)
	}
	return tpl, nil
}

// Update updates an existing template (provider is immutable).
func (s *ChannelMonitorRequestTemplateService) Update(reqCtx context.Context, templateID int64, params ChannelMonitorRequestTemplateUpdateParams) (*ChannelMonitorRequestTemplate, error) {
	current, fetchErr := s.repo.GetByID(reqCtx, templateID)
	if fetchErr != nil {
		return nil, fetchErr
	}
	if mergeErr := applyTemplateUpdate(current, params); mergeErr != nil {
		return nil, mergeErr
	}
	if saveErr := s.repo.Update(reqCtx, current); saveErr != nil {
		return nil, fmt.Errorf("update template: %w", saveErr)
	}
	return current, nil
}

// Delete removes a template. Associated monitors have their template_id SET NULL
// and retain their snapshot configuration.
func (s *ChannelMonitorRequestTemplateService) Delete(reqCtx context.Context, templateID int64) error {
	if delErr := s.repo.Delete(reqCtx, templateID); delErr != nil {
		return fmt.Errorf("delete template: %w", delErr)
	}
	return nil
}

// ApplyToMonitors copies the template's current configuration to the specified associated monitors.
// monitorIDs must be non-empty and each must have template_id = id; non-matching IDs are filtered by SQL.
// Returns the count of monitors actually updated.
func (s *ChannelMonitorRequestTemplateService) ApplyToMonitors(reqCtx context.Context, templateID int64, monitorIDs []int64) (int64, error) {
	if _, fetchErr := s.repo.GetByID(reqCtx, templateID); fetchErr != nil {
		return 0, fetchErr
	}
	if len(monitorIDs) == 0 {
		return 0, ErrChannelMonitorTemplateApplyEmpty
	}
	affected, applyErr := s.repo.ApplyToMonitors(reqCtx, templateID, monitorIDs)
	if applyErr != nil {
		return 0, fmt.Errorf("apply template to monitors: %w", applyErr)
	}
	return affected, nil
}

// CountAssociatedMonitors returns the count of monitors linked to the template.
func (s *ChannelMonitorRequestTemplateService) CountAssociatedMonitors(reqCtx context.Context, templateID int64) (int64, error) {
	return s.repo.CountAssociatedMonitors(reqCtx, templateID)
}

// ListAssociatedMonitors returns brief info for all monitors linked to the template.
// Used by the frontend apply picker to avoid an extra list+filter round trip.
func (s *ChannelMonitorRequestTemplateService) ListAssociatedMonitors(reqCtx context.Context, templateID int64) ([]*AssociatedMonitorBrief, error) {
	if _, fetchErr := s.repo.GetByID(reqCtx, templateID); fetchErr != nil {
		return nil, fetchErr
	}
	return s.repo.ListAssociatedMonitors(reqCtx, templateID)
}

// ---------- Validation and utilities ----------

// validateTemplateCreateParams aggregates all create-parameter validations.
func validateTemplateCreateParams(params ChannelMonitorRequestTemplateCreateParams) error {
	if strings.TrimSpace(params.Name) == "" {
		return ErrChannelMonitorTemplateMissingName
	}
	if provErr := validateProvider(params.Provider); provErr != nil {
		return ErrChannelMonitorTemplateInvalidProvider
	}
	if modeErr := validateAPIMode(params.Provider, params.APIMode); modeErr != nil {
		return ErrChannelMonitorTemplateInvalidAPIMode
	}
	if bodyErr := validateBodyModeForProtocol(params.Provider, params.APIMode, params.BodyOverrideMode, params.BodyOverride); bodyErr != nil {
		return bodyErr
	}
	if headerErr := validateExtraHeaders(params.ExtraHeaders); headerErr != nil {
		return headerErr
	}
	return nil
}

// applyTemplateUpdate merges non-nil update fields onto the existing template.
func applyTemplateUpdate(current *ChannelMonitorRequestTemplate, params ChannelMonitorRequestTemplateUpdateParams) error {
	if params.Name != nil {
		trimmedName := strings.TrimSpace(*params.Name)
		if trimmedName == "" {
			return ErrChannelMonitorTemplateMissingName
		}
		current.Name = trimmedName
	}
	if params.Description != nil {
		current.Description = strings.TrimSpace(*params.Description)
	}
	effectiveAPIMode := defaultAPIMode(current.APIMode)
	if params.APIMode != nil {
		effectiveAPIMode = defaultAPIMode(*params.APIMode)
	}
	if modeErr := validateAPIMode(current.Provider, effectiveAPIMode); modeErr != nil {
		return ErrChannelMonitorTemplateInvalidAPIMode
	}
	if params.ExtraHeaders != nil {
		if headerErr := validateExtraHeaders(*params.ExtraHeaders); headerErr != nil {
			return headerErr
		}
		current.ExtraHeaders = emptyHeadersIfNil(*params.ExtraHeaders)
	}
	// Body override mode/body are validated together: either may change, so always use effective values.
	effectiveBodyMode := current.BodyOverrideMode
	effectiveBody := current.BodyOverride
	if params.BodyOverrideMode != nil {
		effectiveBodyMode = *params.BodyOverrideMode
	}
	if params.BodyOverride != nil {
		effectiveBody = *params.BodyOverride
	}
	if bodyErr := validateBodyModeForProtocol(current.Provider, effectiveAPIMode, effectiveBodyMode, effectiveBody); bodyErr != nil {
		return bodyErr
	}
	current.APIMode = effectiveAPIMode
	current.BodyOverrideMode = defaultBodyMode(effectiveBodyMode)
	current.BodyOverride = effectiveBody
	return nil
}

// validateBodyModeForProtocol checks body_override_mode against provider/api_mode protocol requirements.
func validateBodyModeForProtocol(prov, apiMode, mode string, body map[string]any) error {
	if modeErr := validateBodyModeParams(mode, body); modeErr != nil {
		return modeErr
	}
	if defaultBodyMode(mode) != MonitorBodyOverrideModeReplace {
		return nil
	}
	if replaceErr := validateReplaceRequestBody(prov, defaultAPIMode(apiMode), body); replaceErr != nil {
		return ErrChannelMonitorInvalidRequestBody
	}
	return nil
}

// validateBodyModeParams checks that the body_override_mode is valid and that
// merge/replace modes have a non-empty body_override.
func validateBodyModeParams(mode string, body map[string]any) error {
	switch mode {
	case "", MonitorBodyOverrideModeOff:
		return nil
	case MonitorBodyOverrideModeMerge, MonitorBodyOverrideModeReplace:
		if len(body) == 0 {
			return ErrChannelMonitorTemplateBodyRequired
		}
		return nil
	default:
		return ErrChannelMonitorTemplateInvalidBodyMode
	}
}

// headerNameRegex matches valid HTTP header names per RFC 7230 token production.
var headerNameRegex = regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+\-.^_` + "`" + `|~]+$`)

// forbiddenHeaderNames lists hop-by-hop and client-managed headers that must not be overridden.
var forbiddenHeaderNames = map[string]bool{
	"host":              true,
	"content-length":    true,
	"content-encoding":  true,
	"transfer-encoding": true,
	"connection":        true,
}

// IsForbiddenHeaderName checks whether the header name is in the deny list.
// Exported so the checker can also filter at runtime as a safety net.
func IsForbiddenHeaderName(headerName string) bool {
	return forbiddenHeaderNames[strings.ToLower(strings.TrimSpace(headerName))]
}

// validateExtraHeaders checks header name format and deny list. Early rejection at save time.
func validateExtraHeaders(headers map[string]string) error {
	for headerKey := range headers {
		if !headerNameRegex.MatchString(headerKey) {
			return ErrChannelMonitorTemplateHeaderInvalidName
		}
		if IsForbiddenHeaderName(headerKey) {
			return ErrChannelMonitorTemplateHeaderForbidden
		}
	}
	return nil
}

// emptyHeadersIfNil normalizes nil map to empty map (repo layer needs non-nil for JSONB).
func emptyHeadersIfNil(headers map[string]string) map[string]string {
	if headers == nil {
		return map[string]string{}
	}
	return headers
}

// defaultBodyMode normalizes empty string to "off".
func defaultBodyMode(mode string) string {
	if mode == "" {
		return MonitorBodyOverrideModeOff
	}
	return mode
}
