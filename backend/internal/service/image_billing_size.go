package service

import (
	"sort"
	"strconv"
	"strings"
)

const (
	ImageBillingSize1K = "1K"
	ImageBillingSize2K = "2K"
	ImageBillingSize4K = "4K"

	ImageSizeSourceOutput  = "output"
	ImageSizeSourceInput   = "input"
	ImageSizeSourceDefault = "default"
	ImageSizeSourceLegacy  = "legacy"
)

type ImageBillingSizeResolution struct {
	BillingSize string
	InputSize   string
	OutputSize  string
	Source      string
	Breakdown   map[string]int
}

func ClassifyImageBillingTier(dim string) (string, bool) {
	lower := strings.ToLower(strings.TrimSpace(dim))
	switch lower {
	case "", "auto":
		return "", false
	case "1k":
		return ImageBillingSize1K, true
	case "2k":
		return ImageBillingSize2K, true
	case "4k":
		return ImageBillingSize4K, true
	case "2048x2048", "2048x1152":
		return ImageBillingSize2K, true
	case "3840x2160", "2160x3840":
		return ImageBillingSize4K, true
	}

	w, h, ok := parseImageBillingDimensions(strings.TrimSpace(dim))
	if !ok {
		return "", false
	}
	longest := w
	if h > longest {
		longest = h
	}
	if longest <= 1024 {
		return ImageBillingSize1K, true
	}
	if longest <= 2048 {
		return ImageBillingSize2K, true
	}
	return ImageBillingSize4K, true
}

func NormalizeImageBillingTierOrDefault(dim string) string {
	if t, ok := ClassifyImageBillingTier(dim); ok {
		return t
	}
	return ImageBillingSize2K
}

func ResolveImageBillingSize(inSize string, outSizes []string) ImageBillingSizeResolution {
	inSize = strings.TrimSpace(inSize)
	outSizes = compactTrimmedStrings(outSizes)

	tierCounts := map[string]int{}
	displayOut := firstDisplayImageOutputSize(outSizes)
	highestTier := ""

	for _, sz := range outSizes {
		t, ok := ClassifyImageBillingTier(sz)
		if !ok {
			continue
		}
		tierCounts[t]++
		if imageTierRank(t) > imageTierRank(highestTier) {
			highestTier = t
		}
	}

	if highestTier != "" {
		return ImageBillingSizeResolution{
			BillingSize: highestTier,
			InputSize:   inSize,
			OutputSize:  displayOut,
			Source:      ImageSizeSourceOutput,
			Breakdown:   normalizeImageSizeBreakdown(tierCounts),
		}
	}

	if t, ok := ClassifyImageBillingTier(inSize); ok {
		return ImageBillingSizeResolution{
			BillingSize: t,
			InputSize:   inSize,
			OutputSize:  displayOut,
			Source:      ImageSizeSourceInput,
		}
	}

	return ImageBillingSizeResolution{
		BillingSize: ImageBillingSize2K,
		InputSize:   inSize,
		OutputSize:  displayOut,
		Source:      ImageSizeSourceDefault,
	}
}

func ApplyOpenAIImageBillingResolution(fwd *OpenAIForwardResult) {
	if fwd == nil || fwd.ImageCount <= 0 {
		return
	}
	inDim := strings.TrimSpace(fwd.ImageInputSize)
	if inDim == "" && strings.TrimSpace(fwd.ImageSize) != ImageBillingSize2K {
		inDim = strings.TrimSpace(fwd.ImageSize)
	}
	outDims := fwd.ImageOutputSizes
	if len(outDims) == 0 && strings.TrimSpace(fwd.ImageOutputSize) != "" {
		outDims = []string{fwd.ImageOutputSize}
	}
	res := ResolveImageBillingSize(inDim, outDims)
	applyImageBillingResolution(
		&fwd.ImageSize,
		&fwd.ImageInputSize,
		&fwd.ImageOutputSize,
		&fwd.ImageSizeSource,
		&fwd.ImageSizeBreakdown,
		res,
	)
}

func ApplyForwardImageBillingResolution(fwd *ForwardResult) {
	if fwd == nil || fwd.ImageCount <= 0 {
		return
	}
	inDim := strings.TrimSpace(fwd.ImageInputSize)
	if inDim == "" && strings.TrimSpace(fwd.ImageSize) != ImageBillingSize2K {
		inDim = strings.TrimSpace(fwd.ImageSize)
	}
	outDims := fwd.ImageOutputSizes
	if len(outDims) == 0 && strings.TrimSpace(fwd.ImageOutputSize) != "" {
		outDims = []string{fwd.ImageOutputSize}
	}
	res := ResolveImageBillingSize(inDim, outDims)
	applyImageBillingResolution(
		&fwd.ImageSize,
		&fwd.ImageInputSize,
		&fwd.ImageOutputSize,
		&fwd.ImageSizeSource,
		&fwd.ImageSizeBreakdown,
		res,
	)
}

func applyImageBillingResolution(
	billingSizePtr *string,
	inputSizePtr *string,
	outputSizePtr *string,
	sourcePtr *string,
	breakdownPtr *map[string]int,
	resolved ImageBillingSizeResolution,
) {
	*billingSizePtr = resolved.BillingSize
	*inputSizePtr = resolved.InputSize
	*outputSizePtr = resolved.OutputSize
	*sourcePtr = resolved.Source
	*breakdownPtr = resolved.Breakdown
}

func parseImageBillingDimensions(raw string) (int, int, bool) {
	segments := strings.Split(strings.ToLower(strings.TrimSpace(raw)), "x")
	if len(segments) != 2 {
		return 0, 0, false
	}
	w, wErr := strconv.Atoi(strings.TrimSpace(segments[0]))
	if wErr != nil {
		return 0, 0, false
	}
	h, hErr := strconv.Atoi(strings.TrimSpace(segments[1]))
	if hErr != nil {
		return 0, 0, false
	}
	if w <= 0 || h <= 0 {
		return 0, 0, false
	}
	return w, h, true
}

func compactTrimmedStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	filtered := make([]string, 0, len(items))
	for _, s := range items {
		cleaned := strings.TrimSpace(s)
		if cleaned != "" {
			filtered = append(filtered, cleaned)
		}
	}
	return filtered
}

func firstDisplayImageOutputSize(sizes []string) string {
	for _, s := range sizes {
		cleaned := strings.TrimSpace(s)
		if cleaned != "" {
			return cleaned
		}
	}
	return ""
}

func imageTierRank(tier string) int {
	switch strings.ToUpper(strings.TrimSpace(tier)) {
	case ImageBillingSize1K:
		return 1
	case ImageBillingSize2K:
		return 2
	case ImageBillingSize4K:
		return 3
	default:
		return 0
	}
}

func normalizeImageSizeBreakdown(counts map[string]int) map[string]int {
	if len(counts) == 0 {
		return nil
	}
	normalized := make(map[string]int, len(counts))
	for _, t := range []string{ImageBillingSize1K, ImageBillingSize2K, ImageBillingSize4K} {
		if n := counts[t]; n > 0 {
			normalized[t] = n
		}
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func SortedImageBillingBreakdownKeys(bkdn map[string]int) []string {
	labels := make([]string, 0, len(bkdn))
	for lbl := range bkdn {
		labels = append(labels, lbl)
	}
	sort.Slice(labels, func(a, b int) bool {
		ra, rb := imageTierRank(labels[a]), imageTierRank(labels[b])
		if ra != rb {
			return ra < rb
		}
		return labels[a] < labels[b]
	})
	return labels
}
