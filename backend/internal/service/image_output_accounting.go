package service

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/tidwall/gjson"
)

type openAIImageOutputCounter struct {
	seen         map[string]struct{}
	seenSizes    map[string]string
	seenOrder    []string
	dataSizes    []string
	count        int
	maxDataCount int
}

func newOpenAIImageOutputCounter() *openAIImageOutputCounter {
	return &openAIImageOutputCounter{
		seen:      make(map[string]struct{}),
		seenSizes: make(map[string]string),
	}
}

func (ctr *openAIImageOutputCounter) Count() int {
	if ctr == nil {
		return 0
	}
	if ctr.maxDataCount > ctr.count {
		return ctr.maxDataCount
	}
	return ctr.count
}

func (ctr *openAIImageOutputCounter) Sizes() []string {
	if ctr == nil {
		return nil
	}
	collected := make([]string, 0, len(ctr.seenOrder)+len(ctr.dataSizes))
	for _, k := range ctr.seenOrder {
		if sz := strings.TrimSpace(ctr.seenSizes[k]); sz != "" {
			collected = append(collected, sz)
		}
	}
	if len(collected) == 0 && len(ctr.dataSizes) > 0 {
		collected = append(collected, ctr.dataSizes...)
	}
	if len(collected) == 0 {
		return nil
	}
	return collected
}

func (ctr *openAIImageOutputCounter) AddJSONResponse(payload []byte) {
	if ctr == nil || len(payload) == 0 || !gjson.ValidBytes(payload) {
		return
	}
	ctr.addDataArray(gjson.GetBytes(payload, "data"))
	ctr.addOutputArray(gjson.GetBytes(payload, "output"))
	ctr.addOutputArray(gjson.GetBytes(payload, "response.output"))
}

func (ctr *openAIImageOutputCounter) AddSSEData(event []byte) {
	if ctr == nil || len(event) == 0 {
		return
	}
	if strings.TrimSpace(string(event)) == "[DONE]" || !gjson.ValidBytes(event) {
		return
	}
	parsed := gjson.ParseBytes(event)
	ctr.addDataArray(parsed.Get("data"))

	evtType := strings.TrimSpace(parsed.Get("type").String())
	switch evtType {
	case "response.output_item.done":
		ctr.addImageOutputItem(parsed.Get("item"))
	case "response.completed", "response.done":
		ctr.addOutputArray(parsed.Get("response.output"))
	case "image_generation.completed":
		if itemResult := parsed.Get("item"); itemResult.Exists() {
			ctr.addImageOutputItem(itemResult)
			return
		}
		if outResult := parsed.Get("output"); outResult.Exists() {
			ctr.addImageOutputItem(outResult)
			return
		}
		ctr.addImageOutputItem(parsed)
	}
}

func (ctr *openAIImageOutputCounter) AddSSEBody(raw string) {
	if ctr == nil || strings.TrimSpace(raw) == "" {
		return
	}
	forEachOpenAISSEDataPayload(raw, ctr.AddSSEData)
}

func (ctr *openAIImageOutputCounter) addDataArray(dataResult gjson.Result) {
	if !dataResult.IsArray() {
		return
	}
	elements := dataResult.Array()
	n := len(elements)
	if n > ctr.maxDataCount {
		ctr.maxDataCount = n
	}
	dims := make([]string, 0, n)
	for _, el := range elements {
		if sz := strings.TrimSpace(el.Get("size").String()); sz != "" {
			dims = append(dims, sz)
		}
	}
	if len(dims) > 0 {
		ctr.dataSizes = dims
	}
}

func (ctr *openAIImageOutputCounter) addOutputArray(outputResult gjson.Result) {
	if !outputResult.IsArray() {
		return
	}
	outputResult.ForEach(func(_, el gjson.Result) bool {
		ctr.addImageOutputItem(el)
		return true
	})
}

func (ctr *openAIImageOutputCounter) addImageOutputItem(entry gjson.Result) {
	if !entry.Exists() || !entry.IsObject() {
		return
	}
	kind := strings.TrimSpace(entry.Get("type").String())
	if kind != "" && kind != "image_generation_call" && kind != "image_generation.completed" {
		return
	}
	if strings.Contains(strings.ToLower(entry.Raw), "partial_image") {
		return
	}

	content := strings.TrimSpace(entry.Get("result").String())
	if content == "" {
		content = strings.TrimSpace(entry.Get("b64_json").String())
	}
	if content == "" {
		content = strings.TrimSpace(entry.Get("url").String())
	}
	if content == "" && kind != "image_generation.completed" {
		return
	}

	dedupKey := strings.TrimSpace(entry.Get("id").String())
	if dedupKey == "" {
		dedupKey = strings.TrimSpace(entry.Get("call_id").String())
	}
	if dedupKey == "" {
		dedupKey = hashOpenAIImageOutputResult(content)
	}
	if dedupKey == "" {
		return
	}

	dim := strings.TrimSpace(entry.Get("size").String())
	if _, already := ctr.seen[dedupKey]; already {
		if dim != "" && strings.TrimSpace(ctr.seenSizes[dedupKey]) == "" {
			ctr.seenSizes[dedupKey] = dim
		}
		return
	}
	ctr.seen[dedupKey] = struct{}{}
	ctr.seenOrder = append(ctr.seenOrder, dedupKey)
	if dim != "" {
		ctr.seenSizes[dedupKey] = dim
	}
	ctr.count++
}

func hashOpenAIImageOutputResult(data string) string {
	data = strings.TrimSpace(data)
	if data == "" {
		return ""
	}
	digest := sha256.Sum256([]byte(data))
	return hex.EncodeToString(digest[:])
}

func countOpenAIResponseImageOutputsFromJSONBytes(payload []byte) int {
	ctr := newOpenAIImageOutputCounter()
	ctr.AddJSONResponse(payload)
	return ctr.Count()
}

func collectOpenAIResponseImageOutputSizesFromJSONBytes(payload []byte) []string {
	ctr := newOpenAIImageOutputCounter()
	ctr.AddJSONResponse(payload)
	return ctr.Sizes()
}

func countOpenAIImageOutputsFromSSEBody(raw string) int {
	ctr := newOpenAIImageOutputCounter()
	ctr.AddSSEBody(raw)
	return ctr.Count()
}

func collectOpenAIImageOutputSizesFromSSEBody(raw string) []string {
	ctr := newOpenAIImageOutputCounter()
	ctr.AddSSEBody(raw)
	return ctr.Sizes()
}
