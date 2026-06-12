package service

import (
	"context"
	"time"
)

// ── 覆盖仓储接口 ──

// ModelOverrideRepository 模型价格覆盖持久化接口（接口定义在 service 包，impl 在 repository 包）。
type ModelOverrideRepository interface {
	// Get 按 model_id 精确查找；未命中返回 (nil, nil)。
	Get(ctx context.Context, modelID string) (*ModelPriceOverride, error)
	// Upsert 创建或全量更新（以 model_id 为冲突键）。
	Upsert(ctx context.Context, o *ModelPriceOverride) error
	// Delete 硬删除，恢复自动定价；model_id 不存在时不报错。
	Delete(ctx context.Context, modelID string) error
	// List 返回全表覆盖记录。
	List(ctx context.Context) ([]ModelPriceOverride, error)
}

// ── 服务端 DTO ──

// ModelPriceOverride 模型价格覆盖（service 层类型）。
type ModelPriceOverride struct {
	ID                int64
	ModelID           string
	PinnedProviderTag string   // 空 = 不指定供应商
	ManualInput       *float64 // nil = 不覆盖
	ManualOutput      *float64
	ManualCacheRead   *float64
	ManualCacheWrite  *float64
	Note              string
	UpdatedBy         int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ── 解析结果 ──

// ResolvedBaseline 解析后的官方价基准（含覆盖应用后的最终值）。
type ResolvedBaseline struct {
	Input      float64
	Output     float64
	CacheRead  float64
	CacheWrite float64
	// Source 价格来源标识：auto-tag / pinned-tag / "manual"
	Source string
	// Overridden 表示数据库中存在 override 记录（即使 manual 字段均为 nil）。
	Overridden bool
	// Override 当前 override 记录快照（nil 表示无记录）。
	Override *ModelPriceOverride
}

// ── Resolver 服务 ──

// OverrideResolver 按优先级规则计算模型官方价基准。
type OverrideResolver struct {
	catalog *OpenRouterCatalogService
	repo    ModelOverrideRepository
}

// NewOverrideResolver 构造函数。
func NewOverrideResolver(catalog *OpenRouterCatalogService, repo ModelOverrideRepository) *OverrideResolver {
	return &OverrideResolver{catalog: catalog, repo: repo}
}

// ResolveBaseline 按优先级规则计算模型最终官方价基准：
//  1. 无 override → auto 最低价
//  2. PinnedProviderTag → 该供应商价作 base
//  3. manual 字段逐字段覆盖 base
//
// localModelOrSlug: 本地模型名或 OpenRouter slug；platform 辅助 MatchModelSlug。
// 返回 ResolvedBaseline；catalog 未命中返回 (nil, nil)（调用方回退 LiteLLM）。
func (r *OverrideResolver) ResolveBaseline(ctx context.Context, localModelOrSlug string, platform string) (*ResolvedBaseline, error) {
	// ── 1. 查 catalog ──
	catalog := r.catalog.List()
	if len(catalog) == 0 {
		return nil, nil
	}
	slug, ok := MatchModelSlug(localModelOrSlug, platform, catalog)
	if !ok {
		return nil, nil
	}
	entry := r.catalog.Get(slug)
	if entry == nil {
		return nil, nil
	}

	// ── 2. 计算 auto baseline ──
	autoInput, autoOutput, autoCacheRead, autoCacheWrite, autoTag := entry.BaselinePrice()

	// ── 3. 查 override 记录 ──
	override, err := r.repo.Get(ctx, localModelOrSlug)
	if err != nil {
		return nil, err
	}
	// 也尝试用 slug 查（若 model_id 存的是 slug）
	if override == nil && slug != localModelOrSlug {
		override, err = r.repo.Get(ctx, slug)
		if err != nil {
			return nil, err
		}
	}

	if override == nil {
		// 无覆盖 → 直接返回 auto
		return &ResolvedBaseline{
			Input:      autoInput,
			Output:     autoOutput,
			CacheRead:  autoCacheRead,
			CacheWrite: autoCacheWrite,
			Source:     autoTag,
			Overridden: false,
			Override:   nil,
		}, nil
	}

	// ── 4. 有 override：先确定 base（pinned 或 auto）──
	baseInput := autoInput
	baseOutput := autoOutput
	baseCacheRead := autoCacheRead
	baseCacheWrite := autoCacheWrite
	baseSource := autoTag

	if override.PinnedProviderTag != "" {
		// 在 providers 里按 tag 找指定供应商
		for _, p := range entry.Providers {
			if p.Tag == override.PinnedProviderTag {
				baseInput = p.Input
				baseOutput = p.Output
				baseCacheRead = p.CacheRead
				baseCacheWrite = p.CacheWrite
				baseSource = p.Tag
				break
			}
		}
		// 若找不到 pinned 供应商，baseSource 保持 autoTag（静默降级）
	}

	// ── 5. 逐字段应用 manual 覆盖 ──
	finalInput := baseInput
	finalOutput := baseOutput
	finalCacheRead := baseCacheRead
	finalCacheWrite := baseCacheWrite

	anyManual := false
	if override.ManualInput != nil {
		finalInput = *override.ManualInput
		anyManual = true
	}
	if override.ManualOutput != nil {
		finalOutput = *override.ManualOutput
		anyManual = true
	}
	if override.ManualCacheRead != nil {
		finalCacheRead = *override.ManualCacheRead
		anyManual = true
	}
	if override.ManualCacheWrite != nil {
		finalCacheWrite = *override.ManualCacheWrite
		anyManual = true
	}

	// ── 6. 确定 source 标注 ──
	finalSource := baseSource
	if anyManual {
		finalSource = "manual"
	}

	return &ResolvedBaseline{
		Input:      finalInput,
		Output:     finalOutput,
		CacheRead:  finalCacheRead,
		CacheWrite: finalCacheWrite,
		Source:     finalSource,
		Overridden: true,
		Override:   override,
	}, nil
}
