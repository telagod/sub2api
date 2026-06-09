package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/tidwall/gjson"
)

func ExtractContentModerationText(protocol string, body []byte) string {
	return ExtractContentModerationInput(protocol, body).Text
}

func ExtractContentModerationInput(protocol string, body []byte) ContentModerationInput {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return ContentModerationInput{}
	}
	var textFragments []string
	var imageRefs []string
	switch protocol {
	case ContentModerationProtocolAnthropicMessages:
		gatherLastAnthropicUserTurn(gjson.GetBytes(body, "messages"), &textFragments, &imageRefs)
	case ContentModerationProtocolOpenAIChat:
		gatherLastRoleTurn(gjson.GetBytes(body, "messages"), "user", &textFragments, &imageRefs)
	case ContentModerationProtocolOpenAIResponses:
		gatherLastResponsesEntry(gjson.GetBytes(body, "input"), &textFragments, &imageRefs)
	case ContentModerationProtocolGemini:
		gatherLastGeminiTurn(gjson.GetBytes(body, "contents"), &textFragments, &imageRefs)
	case ContentModerationProtocolOpenAIImages:
		appendTextFragment(&textFragments, gjson.GetBytes(body, "prompt").String())
		walkContentNode(gjson.GetBytes(body, "images"), &textFragments, &imageRefs)
	default:
		gatherLastResponsesEntry(gjson.GetBytes(body, "input"), &textFragments, &imageRefs)
		gatherLastRoleTurn(gjson.GetBytes(body, "messages"), "user", &textFragments, &imageRefs)
		gatherLastGeminiTurn(gjson.GetBytes(body, "contents"), &textFragments, &imageRefs)
	}
	result := ContentModerationInput{
		Text:   collapseWhitespace(strings.Join(textFragments, "\n")),
		Images: deduplicateImages(imageRefs),
	}
	result.Normalize()
	return result
}

// gatherLastRoleTurn extracts text and images from the last message with the specified role.
func gatherLastRoleTurn(messages gjson.Result, targetRole string, textFragments *[]string, imageRefs *[]string) {
	if !messages.IsArray() {
		return
	}
	items := messages.Array()
	if len(items) == 0 {
		return
	}
	finalMsg := items[len(items)-1]
	if strings.ToLower(strings.TrimSpace(finalMsg.Get("role").String())) != targetRole {
		return
	}
	var candidateText []string
	var candidateImages []string
	walkContentNode(finalMsg.Get("content"), &candidateText, &candidateImages)
	if collapseWhitespace(strings.Join(candidateText, "\n")) == "" && len(candidateImages) == 0 {
		return
	}
	*textFragments = append(*textFragments, candidateText...)
	*imageRefs = append(*imageRefs, candidateImages...)
}

// gatherLastAnthropicUserTurn extracts content from the last Anthropic user message.
func gatherLastAnthropicUserTurn(messages gjson.Result, textFragments *[]string, imageRefs *[]string) {
	if !messages.IsArray() {
		return
	}
	items := messages.Array()
	if len(items) == 0 {
		return
	}
	finalMsg := items[len(items)-1]
	if strings.ToLower(strings.TrimSpace(finalMsg.Get("role").String())) != "user" {
		return
	}
	var candidateText []string
	var candidateImages []string
	walkAnthropicUserContent(finalMsg.Get("content"), &candidateText, &candidateImages)
	if collapseWhitespace(strings.Join(candidateText, "\n")) == "" && len(candidateImages) == 0 {
		return
	}
	*textFragments = append(*textFragments, candidateText...)
	*imageRefs = append(*imageRefs, candidateImages...)
}

// walkAnthropicUserContent recursively walks Anthropic content structures.
func walkAnthropicUserContent(node gjson.Result, textFragments *[]string, imageRefs *[]string) {
	switch {
	case !node.Exists():
		return
	case node.Type == gjson.String:
		if !looksLikeSystemReminder(node.String()) {
			appendTextFragment(textFragments, node.String())
		}
	case node.IsArray():
		node.ForEach(func(_, elem gjson.Result) bool {
			walkAnthropicUserContent(elem, textFragments, imageRefs)
			return true
		})
	case node.IsObject():
		contentType := strings.ToLower(strings.TrimSpace(node.Get("type").String()))
		switch contentType {
		case "", "text", "input_text", "message":
			if node.Get("text").Exists() && !looksLikeSystemReminder(node.Get("text").String()) {
				appendTextFragment(textFragments, node.Get("text").String())
			}
			if node.Get("content").Exists() {
				walkAnthropicUserContent(node.Get("content"), textFragments, imageRefs)
			}
		case "image_url", "input_image", "image":
			walkContentNode(node, textFragments, imageRefs)
		}
	}
}

// looksLikeSystemReminder checks if text starts with the system-reminder XML tag.
func looksLikeSystemReminder(text string) bool {
	return strings.HasPrefix(strings.TrimSpace(text), "<system-reminder>")
}

// gatherLastResponsesEntry extracts content from the last Responses API input item.
func gatherLastResponsesEntry(inputNode gjson.Result, textFragments *[]string, imageRefs *[]string) {
	switch {
	case !inputNode.Exists():
		return
	case inputNode.Type == gjson.String:
		appendTextFragment(textFragments, inputNode.String())
	case inputNode.IsArray():
		elements := inputNode.Array()
		if len(elements) == 0 {
			return
		}
		lastElem := elements[len(elements)-1]
		if !isResponsesUserItem(lastElem) {
			return
		}
		walkContentNode(lastElem.Get("content"), textFragments, imageRefs)
		if lastElem.Get("type").String() == "input_text" || lastElem.Get("text").Exists() {
			walkContentNode(lastElem, textFragments, imageRefs)
		}
	case inputNode.IsObject():
		if isResponsesUserItem(inputNode) {
			walkContentNode(inputNode.Get("content"), textFragments, imageRefs)
			if inputNode.Get("type").String() == "input_text" || inputNode.Get("text").Exists() {
				walkContentNode(inputNode, textFragments, imageRefs)
			}
		}
	}
}

// isResponsesUserItem determines if a Responses input item is from the user role.
func isResponsesUserItem(elem gjson.Result) bool {
	role := strings.ToLower(strings.TrimSpace(elem.Get("role").String()))
	if role == "user" {
		return responsesItemContainsModerationContent(elem)
	}
	if role != "" {
		return false
	}
	return responsesItemContainsModerationContent(elem)
}

// responsesItemContainsModerationContent checks if an item has extractable text or images.
func responsesItemContainsModerationContent(elem gjson.Result) bool {
	var probeText []string
	var probeImages []string
	walkContentNode(elem.Get("content"), &probeText, &probeImages)
	if elem.Get("type").String() == "input_text" || elem.Get("text").Exists() {
		walkContentNode(elem, &probeText, &probeImages)
	}
	return collapseWhitespace(strings.Join(probeText, "\n")) != "" || len(probeImages) > 0
}

// gatherLastGeminiTurn extracts content from the last Gemini user turn.
func gatherLastGeminiTurn(contents gjson.Result, textFragments *[]string, imageRefs *[]string) {
	if !contents.IsArray() {
		return
	}
	elements := contents.Array()
	if len(elements) == 0 {
		return
	}
	lastElem := elements[len(elements)-1]
	role := strings.ToLower(strings.TrimSpace(lastElem.Get("role").String()))
	if role != "" && role != "user" {
		return
	}
	var candidateText []string
	var candidateImages []string
	if partsNode := lastElem.Get("parts"); partsNode.IsArray() {
		partsNode.ForEach(func(_, segment gjson.Result) bool {
			appendTextFragment(&candidateText, segment.Get("text").String())
			extractGeminiPartImage(&candidateImages, segment)
			return true
		})
	}
	if collapseWhitespace(strings.Join(candidateText, "\n")) == "" && len(candidateImages) == 0 {
		return
	}
	*textFragments = append(*textFragments, candidateText...)
	*imageRefs = append(*imageRefs, candidateImages...)
}

// walkContentNode recursively walks a generic content node, extracting text and images.
func walkContentNode(node gjson.Result, textFragments *[]string, imageRefs *[]string) {
	switch {
	case !node.Exists():
		return
	case node.Type == gjson.String:
		appendTextFragment(textFragments, node.String())
	case node.IsArray():
		node.ForEach(func(_, child gjson.Result) bool {
			walkContentNode(child, textFragments, imageRefs)
			return true
		})
	case node.IsObject():
		contentType := strings.ToLower(strings.TrimSpace(node.Get("type").String()))
		appendImageRef(imageRefs, node.Get("image_url.url").String())
		appendImageRef(imageRefs, node.Get("image_url").String())
		appendImageRef(imageRefs, node.Get("url").String())
		appendImageFromParts(imageRefs, node.Get("source.media_type").String(), node.Get("source.data").String())
		appendImageFromParts(imageRefs, node.Get("source.mediaType").String(), node.Get("source.data").String())
		appendImageFromParts(imageRefs, node.Get("media_type").String(), node.Get("data").String())
		appendImageFromParts(imageRefs, node.Get("mime_type").String(), node.Get("data").String())
		appendImageFromParts(imageRefs, node.Get("mimeType").String(), node.Get("data").String())
		appendImageRef(imageRefs, node.Get("source.data").String())
		appendImageRef(imageRefs, node.Get("data").String())
		appendImageRef(imageRefs, node.Get("base64").String())
		switch contentType {
		case "", "text", "input_text", "message":
			if node.Get("text").Exists() {
				appendTextFragment(textFragments, node.Get("text").String())
			}
			if node.Get("content").Exists() {
				walkContentNode(node.Get("content"), textFragments, imageRefs)
			}
		case "image_url", "input_image", "image":
		}
	}
}

// extractGeminiPartImage extracts image references from a Gemini parts segment.
func extractGeminiPartImage(imageRefs *[]string, segment gjson.Result) {
	if inlineNode := segment.Get("inline_data"); inlineNode.IsObject() {
		mtype := strings.TrimSpace(inlineNode.Get("mime_type").String())
		payload := strings.TrimSpace(inlineNode.Get("data").String())
		if mtype != "" && payload != "" {
			appendImageRef(imageRefs, fmt.Sprintf("data:%s;base64,%s", mtype, payload))
		}
	}
	if inlineNode := segment.Get("inlineData"); inlineNode.IsObject() {
		mtype := strings.TrimSpace(inlineNode.Get("mimeType").String())
		payload := strings.TrimSpace(inlineNode.Get("data").String())
		if mtype != "" && payload != "" {
			appendImageRef(imageRefs, fmt.Sprintf("data:%s;base64,%s", mtype, payload))
		}
	}
	appendImageRef(imageRefs, segment.Get("file_data.file_uri").String())
	appendImageRef(imageRefs, segment.Get("fileData.fileUri").String())
}

// appendImageFromParts constructs a data URL from MIME type and base64 data, then appends it.
func appendImageFromParts(imageRefs *[]string, mimeType string, data string) {
	mimeType = strings.TrimSpace(mimeType)
	data = strings.TrimSpace(data)
	if mimeType == "" || data == "" {
		return
	}
	appendImageRef(imageRefs, fmt.Sprintf("data:%s;base64,%s", mimeType, data))
}

// appendImageRef appends a valid image URL to the list.
func appendImageRef(imageRefs *[]string, ref string) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return
	}
	if strings.HasPrefix(ref, "data:") || strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		*imageRefs = append(*imageRefs, ref)
	}
}

// appendTextFragment appends non-empty, non-system-reminder text to the list.
func appendTextFragment(fragments *[]string, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	if strings.Contains(text, "<system-reminder>") {
		return
	}
	*fragments = append(*fragments, text)
}

// normalizeContentModerationText collapses whitespace and trims a text string.
func normalizeContentModerationText(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

// pickRandomIndex returns a cryptographically random index in [0, count).
func pickRandomIndex(count int) (int, error) {
	idx, err := rand.Int(rand.Reader, big.NewInt(int64(count)))
	if err != nil {
		return 0, err
	}
	return int(idx.Int64()), nil
}
