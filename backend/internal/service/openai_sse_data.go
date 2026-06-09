package service

import (
	"strings"

	"github.com/tidwall/gjson"
)

type openAISSEDataAccumulator struct {
	buf []string
}

func (acc *openAISSEDataAccumulator) AddLine(raw string, emit func([]byte)) {
	if emit == nil {
		return
	}
	stripped := strings.TrimRight(raw, "\r\n")
	if payload, ok := extractOpenAISSEDataLine(stripped); ok {
		acc.buf = append(acc.buf, payload)
		return
	}
	if strings.TrimSpace(stripped) == "" {
		acc.Flush(emit)
	}
}

func (acc *openAISSEDataAccumulator) Flush(emit func([]byte)) {
	if emit == nil || len(acc.buf) == 0 {
		return
	}
	dispatchSSEPayloads(acc.buf, emit)
	acc.buf = acc.buf[:0]
}

func forEachOpenAISSEDataPayload(body string, emit func([]byte)) {
	if emit == nil || strings.TrimSpace(body) == "" {
		return
	}
	var collector openAISSEDataAccumulator
	for _, ln := range strings.Split(body, "\n") {
		collector.AddLine(ln, emit)
	}
	collector.Flush(emit)
}

func dispatchSSEPayloads(parts []string, emit func([]byte)) {
	if emit == nil || len(parts) == 0 {
		return
	}
	if len(parts) == 1 {
		sendSSEPayload(parts[0], emit)
		return
	}
	combined := strings.Join(parts, "\n")
	if gjson.Valid(combined) {
		sendSSEPayload(combined, emit)
		return
	}
	for _, p := range parts {
		sendSSEPayload(p, emit)
	}
}

func sendSSEPayload(data string, emit func([]byte)) {
	trimmed := strings.TrimSpace(data)
	if trimmed == "" || trimmed == "[DONE]" {
		return
	}
	emit([]byte(trimmed))
}
