package service

import (
	"strings"

	"github.com/tidwall/gjson"
)

const (
	openAIResponsesEndpoint          = "/v1/responses"
	openAIResponsesCompactEndpoint   = "/v1/responses/compact"
	imageGenerationPermissionMessage = "Image generation is not enabled for this group"
)

// ImageGenerationPermissionMessage returns the stable end-user error text for disabled groups.
func ImageGenerationPermissionMessage() string {
	return imageGenerationPermissionMessage
}

// GroupAllowsImageGeneration preserves ungrouped-key behavior and enforces the flag when a group is present.
func GroupAllowsImageGeneration(grp *Group) bool {
	return grp == nil || grp.AllowImageGeneration
}

// IsImageGenerationIntent classifies requests that can produce generated images.
func IsImageGenerationIntent(ep string, requestedModel string, body []byte) bool {
	if IsImageGenerationEndpoint(ep) {
		return true
	}
	if isOpenAIImageGenerationModel(requestedModel) {
		return true
	}
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return false
	}
	bodyModel := strings.TrimSpace(gjson.GetBytes(body, "model").String())
	if isOpenAIImageGenerationModel(bodyModel) {
		return true
	}
	if openAIJSONToolsContainImageGeneration(gjson.GetBytes(body, "tools")) {
		return true
	}
	return openAIJSONToolChoiceSelectsImageGeneration(gjson.GetBytes(body, "tool_choice"))
}

// IsImageGenerationIntentMap is the map-backed variant used after service-side request mutation.
func IsImageGenerationIntentMap(ep string, requestedModel string, reqMap map[string]any) bool {
	if IsImageGenerationEndpoint(ep) {
		return true
	}
	if isOpenAIImageGenerationModel(requestedModel) {
		return true
	}
	if reqMap == nil {
		return false
	}
	if isOpenAIImageGenerationModel(firstNonEmptyString(reqMap["model"])) {
		return true
	}
	if hasOpenAIImageGenerationToolV2(reqMap) {
		return true
	}
	return openAIAnyToolChoiceSelectsImageGeneration(reqMap["tool_choice"])
}

// IsImageGenerationEndpoint identifies dedicated generated-image endpoints.
func IsImageGenerationEndpoint(ep string) bool {
	normalized := normalizeImageGenerationEndpoint(ep)
	return normalized == "/v1/images/generations" ||
		normalized == "/v1/images/edits" ||
		normalized == "/images/generations" ||
		normalized == "/images/edits"
}

func normalizeImageGenerationEndpoint(ep string) string {
	ep = strings.TrimSpace(strings.ToLower(ep))
	if ep == "" {
		return ""
	}
	ep = strings.TrimPrefix(ep, "https://api.openai.com")
	if qmark := strings.IndexByte(ep, '?'); qmark >= 0 {
		ep = ep[:qmark]
	}
	return strings.TrimRight(ep, "/")
}

func openAIJSONToolsContainImageGeneration(toolsArr gjson.Result) bool {
	if !toolsArr.IsArray() {
		return false
	}
	matched := false
	toolsArr.ForEach(func(_, entry gjson.Result) bool {
		if openAIJSONString(entry.Get("type")) == "image_generation" {
			matched = true
			return false
		}
		return true
	})
	return matched
}

func openAIRequestBodyHasImageGenerationTool(payload []byte) bool {
	if len(payload) == 0 || !gjson.ValidBytes(payload) {
		return false
	}
	return openAIJSONToolsContainImageGeneration(gjson.GetBytes(payload, "tools"))
}

func openAIRequestBodyImageGenerationToolNeedsNormalization(payload []byte) bool {
	if len(payload) == 0 || !gjson.ValidBytes(payload) {
		return false
	}
	toolsArr := gjson.GetBytes(payload, "tools")
	if !toolsArr.IsArray() {
		return false
	}
	needsFix := false
	toolsArr.ForEach(func(_, entry gjson.Result) bool {
		if openAIJSONString(entry.Get("type")) != "image_generation" {
			return true
		}
		// Only enter map-based modification when legacy fields need migration.
		if entry.Get("format").Exists() || entry.Get("compression").Exists() {
			needsFix = true
			return false
		}
		return true
	})
	return needsFix
}

func openAIJSONToolChoiceSelectsImageGeneration(choiceResult gjson.Result) bool {
	if !choiceResult.Exists() {
		return false
	}
	if choiceResult.Type == gjson.String {
		return strings.TrimSpace(choiceResult.String()) == "image_generation"
	}
	if !choiceResult.IsObject() {
		return false
	}
	if strings.TrimSpace(choiceResult.Get("type").String()) == "image_generation" {
		return true
	}
	if strings.TrimSpace(choiceResult.Get("tool.type").String()) == "image_generation" {
		return true
	}
	return strings.TrimSpace(choiceResult.Get("function.name").String()) == "image_generation"
}

func openAIAnyToolChoiceSelectsImageGeneration(raw any) bool {
	switch typed := raw.(type) {
	case string:
		return strings.TrimSpace(typed) == "image_generation"
	case map[string]any:
		if strings.TrimSpace(firstNonEmptyString(typed["type"])) == "image_generation" {
			return true
		}
		if toolMap, ok := typed["tool"].(map[string]any); ok {
			if strings.TrimSpace(firstNonEmptyString(toolMap["type"])) == "image_generation" {
				return true
			}
		}
		if fnMap, ok := typed["function"].(map[string]any); ok {
			if strings.TrimSpace(firstNonEmptyString(fnMap["name"])) == "image_generation" {
				return true
			}
		}
	}
	return false
}

func getAPIKeyFromContext(ctx interface{ Get(string) (any, bool) }) *APIKey {
	if ctx == nil {
		return nil
	}
	raw, found := ctx.Get("api_key")
	if !found {
		return nil
	}
	typed, _ := raw.(*APIKey)
	return typed
}

func apiKeyGroup(ak *APIKey) *Group {
	if ak == nil {
		return nil
	}
	return ak.Group
}

type OpenAIResponsesImageBillingConfig struct {
	Model     string
	SizeTier  string
	InputSize string
}

func resolveOpenAIResponsesImageBillingConfigDetailed(reqMap map[string]any, fallbackModel string) (OpenAIResponsesImageBillingConfig, error) {
	imgModel := ""
	imgSize := ""
	foundImageTool := false

	if reqMap != nil {
		toolsList, _ := reqMap["tools"].([]any)
		for _, rawEntry := range toolsList {
			entry, ok := rawEntry.(map[string]any)
			if !ok {
				continue
			}
			if strings.TrimSpace(firstNonEmptyString(entry["type"])) != "image_generation" {
				continue
			}
			foundImageTool = true
			imgModel = strings.TrimSpace(firstNonEmptyString(entry["model"]))
			imgSize = strings.TrimSpace(firstNonEmptyString(entry["size"]))
			break
		}
		if imgSize == "" {
			imgSize = strings.TrimSpace(firstNonEmptyString(reqMap["size"]))
		}
	}

	if imgModel == "" && reqMap != nil {
		topModel := strings.TrimSpace(firstNonEmptyString(reqMap["model"]))
		if isOpenAIImageBillingModelAlias(topModel) || !foundImageTool {
			imgModel = topModel
		}
	}
	if imgModel == "" && foundImageTool {
		imgModel = "gpt-image-2"
	}
	if imgModel == "" {
		imgModel = strings.TrimSpace(fallbackModel)
	}

	return OpenAIResponsesImageBillingConfig{
		Model:     imgModel,
		SizeTier:  normalizeOpenAIImageSizeTier(imgSize),
		InputSize: imgSize,
	}, nil
}

func resolveOpenAIResponsesImageBillingConfigFromBody(payload []byte, fallbackModel string) (string, string, error) {
	cfg, cfgErr := resolveOpenAIResponsesImageBillingConfigDetailedFromBody(payload, fallbackModel)
	if cfgErr != nil {
		return "", "", cfgErr
	}
	return cfg.Model, cfg.SizeTier, nil
}

func resolveOpenAIResponsesImageBillingConfigDetailedFromBody(payload []byte, fallbackModel string) (OpenAIResponsesImageBillingConfig, error) {
	imgModel := ""
	imgSize := ""
	foundImageTool := false

	if len(payload) > 0 && gjson.ValidBytes(payload) {
		toolsArr := gjson.GetBytes(payload, "tools")
		if toolsArr.IsArray() {
			toolsArr.ForEach(func(_, entry gjson.Result) bool {
				if openAIJSONString(entry.Get("type")) != "image_generation" {
					return true
				}
				foundImageTool = true
				imgModel = openAIJSONString(entry.Get("model"))
				imgSize = openAIJSONString(entry.Get("size"))
				return false
			})
		}
		if imgSize == "" {
			imgSize = openAIJSONString(gjson.GetBytes(payload, "size"))
		}
		if imgModel == "" {
			topModel := openAIJSONString(gjson.GetBytes(payload, "model"))
			if isOpenAIImageBillingModelAlias(topModel) || !foundImageTool {
				imgModel = topModel
			}
		}
	}

	if imgModel == "" && foundImageTool {
		imgModel = "gpt-image-2"
	}
	if imgModel == "" {
		imgModel = strings.TrimSpace(fallbackModel)
	}

	return OpenAIResponsesImageBillingConfig{
		Model:     imgModel,
		SizeTier:  normalizeOpenAIImageSizeTier(imgSize),
		InputSize: imgSize,
	}, nil
}

func isOpenAIImageBillingModelAlias(mdl string) bool {
	lower := strings.ToLower(strings.TrimSpace(mdl))
	if lower == "" {
		return false
	}
	return isOpenAIImageGenerationModel(lower) || strings.Contains(lower, "image")
}

func openAIJSONString(val gjson.Result) string {
	if val.Type != gjson.String {
		return ""
	}
	return strings.TrimSpace(val.String())
}
