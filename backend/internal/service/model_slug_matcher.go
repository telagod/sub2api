package service

import (
	"regexp"
	"strings"
)

// 归一化辅助：去厂商前缀
var knownPrefixes = []string{
	"anthropic/",
	"openai/",
	"google/",
	"meta-llama/",
	"mistralai/",
	"cohere/",
	"deepseek/",
	"qwen/",
	"x-ai/",
	"amazon/",
}

// 尾部日期后缀，例如 -20250101 / -2025-01-01 / :20250101
var reDateSuffix = regexp.MustCompile(`[-:]?\d{6,8}$`)

// 归一化：小写、去厂商前缀、去 -latest、去尾部日期、统一 . → -
func normalizeModelName(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))

	// 去厂商前缀（迭代一次，最长前缀优先）
	for _, pfx := range knownPrefixes {
		if after, ok := strings.CutPrefix(s, pfx); ok {
			s = after
			break
		}
	}

	// 去 -latest 后缀（可能在日期剥离前出现）
	s = strings.TrimSuffix(s, "-latest")
	s = strings.TrimSuffix(s, ":latest")

	// 去尾部日期
	s = reDateSuffix.ReplaceAllString(s, "")

	// 再次去 -latest（日期后面带 -latest 的情况）
	s = strings.TrimSuffix(s, "-latest")

	// 统一分隔符：. → -
	s = strings.ReplaceAll(s, ".", "-")

	return s
}

// platformToProviderPrefix 将 platform 提示映射到 OpenRouter slug 前缀。
var platformToProviderPrefix = map[string]string{
	"anthropic": "anthropic/",
	"openai":    "openai/",
	"google":    "google/",
}

// MatchModelSlug 将本地模型名映射到 OpenRouter catalog slug。
//
// 匹配顺序：
//  1. 直接 slug 命中（catalog 里 ID 完全相同）。
//  2. 归一化后精确相等（去前缀、去日期/latest、. ↔ -）。
//  3. platform 提示 + 归一化前缀模糊匹配；多候选取 slug 最短（最规范）那个。
//
// 匹配不到返回 ("", false)，绝不瞎配。
func MatchModelSlug(localModel string, platform string, catalog []CatalogModel) (slug string, ok bool) {
	if len(catalog) == 0 || strings.TrimSpace(localModel) == "" {
		return "", false
	}

	// ── 第一关：直接精确命中 catalog ID ──
	for i := range catalog {
		if catalog[i].ID == localModel {
			return catalog[i].ID, true
		}
	}

	normLocal := normalizeModelName(localModel)

	// 当 platform 指定时，限定供应商前缀范围（避免跨厂商误匹配）
	pfx, hasPfx := platformToProviderPrefix[strings.ToLower(strings.TrimSpace(platform))]

	// ── 第二关：归一化精确匹配 ──
	for i := range catalog {
		// platform 有值时：仅在该 platform 下查
		if hasPfx && !strings.HasPrefix(catalog[i].ID, pfx) {
			continue
		}
		normCat := normalizeModelName(catalog[i].ID)
		if normCat == normLocal {
			return catalog[i].ID, true
		}
	}

	// ── 第三关：platform 提示 + 前缀模糊匹配 ──
	// 短名（如 "gpt-4"）做前缀模糊会误配多个候选（gpt-4o / gpt-4o-mini），
	// 设最短长度门槛，宁缺毋滥——失配只影响展示价回退，不碰计费。
	if len(normLocal) < 6 {
		return "", false
	}
	if !hasPfx {
		return "", false
	}

	var candidates []CatalogModel
	for i := range catalog {
		if !strings.HasPrefix(catalog[i].ID, pfx) {
			continue
		}
		normCat := normalizeModelName(catalog[i].ID)
		// normLocal 须是 normCat 的前缀（family 前缀模糊匹配）
		if normCat == normLocal || strings.HasPrefix(normCat, normLocal) || strings.HasPrefix(normLocal, normCat) {
			candidates = append(candidates, catalog[i])
		}
	}
	if len(candidates) == 0 {
		return "", false
	}

	// 多候选：取 slug 最短（最规范）那个
	best := candidates[0]
	for _, c := range candidates[1:] {
		if len(c.ID) < len(best.ID) {
			best = c
		}
	}
	return best.ID, true
}
