//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// 取舍核心：一模型多供应商时，官方价基准取「各供应商最低价」，并返回来源 tag。
func TestCatalogModel_BaselinePrice_LowestAcrossProviders(t *testing.T) {
	m := &CatalogModel{
		Repr: CatalogProviderPrice{Input: 9e-6, Output: 9e-6, Tag: "repr"},
		Providers: []CatalogProviderPrice{
			{Provider: "Together", Tag: "together", Input: 5e-6, Output: 8e-6},
			{Provider: "DeepInfra", Tag: "deepinfra", Input: 3e-6, Output: 5e-6}, // 最低
			{Provider: "Fireworks", Tag: "fireworks", Input: 4e-6, Output: 6e-6},
		},
	}
	in, out, _, _, src := m.BaselinePrice()
	require.InDelta(t, 3e-6, in, 1e-12)
	require.InDelta(t, 5e-6, out, 1e-12)
	require.Equal(t, "deepinfra", src, "应取最低价供应商作为来源")
}

// 首方模型各供应商同价：仍返回其中之一，价正确。
func TestCatalogModel_BaselinePrice_FirstPartyEqualProviders(t *testing.T) {
	m := &CatalogModel{
		Providers: []CatalogProviderPrice{
			{Provider: "Anthropic", Tag: "anthropic", Input: 3e-6, Output: 15e-6, CacheRead: 0.3e-6},
			{Provider: "Amazon Bedrock", Tag: "amazon-bedrock", Input: 3e-6, Output: 15e-6, CacheRead: 0.3e-6},
		},
	}
	in, out, cr, _, src := m.BaselinePrice()
	require.InDelta(t, 3e-6, in, 1e-12)
	require.InDelta(t, 15e-6, out, 1e-12)
	require.InDelta(t, 0.3e-6, cr, 1e-12)
	require.Contains(t, []string{"anthropic", "amazon-bedrock"}, src)
}

// providers 为空时回退代表价（/models 的单一价）。
func TestCatalogModel_BaselinePrice_FallbackToRepr(t *testing.T) {
	m := &CatalogModel{Repr: CatalogProviderPrice{Input: 2e-6, Output: 4e-6, Tag: "repr"}}
	in, out, _, _, src := m.BaselinePrice()
	require.InDelta(t, 2e-6, in, 1e-12)
	require.InDelta(t, 4e-6, out, 1e-12)
	require.Equal(t, "repr", src)
}

// 缺价供应商（0 价）被跳过，不会误选为最低。
func TestCatalogModel_BaselinePrice_SkipsZeroPriced(t *testing.T) {
	m := &CatalogModel{
		Providers: []CatalogProviderPrice{
			{Provider: "Broken", Tag: "broken", Input: 0, Output: 0},
			{Provider: "Real", Tag: "real", Input: 6e-6, Output: 12e-6},
		},
	}
	in, _, _, _, src := m.BaselinePrice()
	require.InDelta(t, 6e-6, in, 1e-12)
	require.Equal(t, "real", src)
}

func TestCapabilitiesFromParams(t *testing.T) {
	caps := capabilitiesFromParams([]string{"tools", "tool_choice", "reasoning", "response_format", "temperature"})
	require.Contains(t, caps, "tools")
	require.Contains(t, caps, "reasoning")
	require.Contains(t, caps, "structured_outputs")
	require.NotContains(t, caps, "temperature")
	// 去重：tools 与 tool_choice 只产生一个 "tools"
	count := 0
	for _, c := range caps {
		if c == "tools" {
			count++
		}
	}
	require.Equal(t, 1, count)
}
