package handler

import "github.com/tidwall/gjson"

const invalidStreamFieldTypeMessage = "invalid stream field type"

// parseOpenAICompatibleStream extracts the "stream" boolean from a request body.
// Returns (streamValue, valid). When the field exists but is not a boolean,
// valid is false.
func parseOpenAICompatibleStream(body []byte) (bool, bool) {
	res := gjson.GetBytes(body, "stream")
	if !res.Exists() {
		return false, true
	}
	if res.Type != gjson.True && res.Type != gjson.False {
		return false, false
	}
	return res.Bool(), true
}
