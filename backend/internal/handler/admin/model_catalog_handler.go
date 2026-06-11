package admin

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/telagod/subme/internal/pkg/response"
	"github.com/telagod/subme/internal/service"
)

// catalogProvider 目录服务最小接口（由 *service.OpenRouterCatalogService 满足）。
// C 注入时直接传 *service.OpenRouterCatalogService，无需额外适配。
type catalogProvider interface {
	List() []service.CatalogModel
	Get(slug string) *service.CatalogModel
	SyncModels(ctx context.Context) (int, error)
	SyncModelEndpoints(ctx context.Context, slug string) error
	LastUpdated() time.Time
}

// ModelCatalogHandler 处理模型目录的只读查询与同步触发。
type ModelCatalogHandler struct {
	catalog catalogProvider
}

// NewModelCatalogHandler 构造函数；catalog 参数为 *service.OpenRouterCatalogService。
func NewModelCatalogHandler(catalog *service.OpenRouterCatalogService) *ModelCatalogHandler {
	return &ModelCatalogHandler{catalog: catalog}
}

// ── DTO ──

// catalogBaselineDTO 官方价基准（per-token，USD）。
type catalogBaselineDTO struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cache_read"`
	CacheWrite float64 `json:"cache_write"`
	Source     string  `json:"source"` // 来源供应商 tag
}

// catalogListItemDTO 列表条目（精简版）。
type catalogListItemDTO struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Description   string             `json:"description"` // 最多 200 字符截断
	ContextLen    int                `json:"context_len"`
	Capabilities  []string           `json:"capabilities"`
	Baseline      catalogBaselineDTO `json:"baseline"`
	ProviderCount int                `json:"provider_count"`
	HasOverride   bool               `json:"has_override"` // 占位，当前恒 false
}

// catalogProviderDTO 供应商价格（详情页）。
type catalogProviderDTO struct {
	Provider   string  `json:"provider"`
	Tag        string  `json:"tag"`
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cache_read"`
	CacheWrite float64 `json:"cache_write"`
	Uptime1d   float64 `json:"uptime_1d"`
	Quant      string  `json:"quant,omitempty"`
}

// catalogDetailDTO 单模型详情（全量）。
type catalogDetailDTO struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	Description  string               `json:"description"` // 完整 description
	ContextLen   int                  `json:"context_len"`
	Modalities   []string             `json:"modalities"`
	Capabilities []string             `json:"capabilities"`
	Baseline     catalogBaselineDTO   `json:"baseline"`
	Providers    []catalogProviderDTO `json:"providers"`
}

// catalogListResponse GET /admin/model-catalog 响应体。
type catalogListResponse struct {
	Models      []catalogListItemDTO `json:"models"`
	LastUpdated string               `json:"last_updated"`
}

// catalogSyncResponse POST /admin/model-catalog/sync 响应体。
type catalogSyncResponse struct {
	Synced int `json:"synced"`
}

// catalogSyncRequest POST /admin/model-catalog/sync 请求体（可选）。
type catalogSyncRequest struct {
	Models []string `json:"models"` // 额外触发 SyncModelEndpoints 的 slug 列表
}

// ── 工具函数 ──

const maxDescriptionLen = 200

func truncateDescription(s string) string {
	runes := []rune(s)
	if len(runes) > maxDescriptionLen {
		return string(runes[:maxDescriptionLen]) + "..."
	}
	return s
}

func toBaselineDTO(m *service.CatalogModel) catalogBaselineDTO {
	inp, out, cr, cw, tag := m.BaselinePrice()
	return catalogBaselineDTO{
		Input:      inp,
		Output:     out,
		CacheRead:  cr,
		CacheWrite: cw,
		Source:     tag,
	}
}

func toProviderDTO(p service.CatalogProviderPrice) catalogProviderDTO {
	return catalogProviderDTO{
		Provider:   p.Provider,
		Tag:        p.Tag,
		Input:      p.Input,
		Output:     p.Output,
		CacheRead:  p.CacheRead,
		CacheWrite: p.CacheWrite,
		Uptime1d:   p.Uptime1d,
		Quant:      p.Quant,
	}
}

// ── 端点处理器 ──

// ListModels GET /admin/model-catalog
// 返回全量目录摘要列表 + last_updated。
func (h *ModelCatalogHandler) ListModels(c *gin.Context) {
	models := h.catalog.List()
	items := make([]catalogListItemDTO, 0, len(models))
	for i := range models {
		m := &models[i]
		items = append(items, catalogListItemDTO{
			ID:            m.ID,
			Name:          m.Name,
			Description:   truncateDescription(m.Description),
			ContextLen:    m.ContextLen,
			Capabilities:  m.Capabilities,
			Baseline:      toBaselineDTO(m),
			ProviderCount: len(m.Providers),
			HasOverride:   false,
		})
	}

	lastUpdated := h.catalog.LastUpdated()
	var lastUpdatedStr string
	if !lastUpdated.IsZero() {
		lastUpdatedStr = lastUpdated.Format(time.RFC3339)
	}

	response.Success(c, catalogListResponse{
		Models:      items,
		LastUpdated: lastUpdatedStr,
	})
}

// GetModel GET /admin/model-catalog/detail?model=<slug>
// 返回单模型全量详情；slug 走 query 参数（slug 含斜杠，gin 路径段不能承载 %2F）。
func (h *ModelCatalogHandler) GetModel(c *gin.Context) {
	slug := strings.TrimSpace(c.Query("model"))
	if slug == "" {
		response.Error(c, http.StatusBadRequest, "model query parameter is required")
		return
	}

	m := h.catalog.Get(slug)
	if m == nil {
		response.Error(c, http.StatusNotFound, "model not found in catalog")
		return
	}

	providers := make([]catalogProviderDTO, 0, len(m.Providers))
	for _, p := range m.Providers {
		providers = append(providers, toProviderDTO(p))
	}

	response.Success(c, catalogDetailDTO{
		ID:           m.ID,
		Name:         m.Name,
		Description:  m.Description,
		ContextLen:   m.ContextLen,
		Modalities:   m.Modalities,
		Capabilities: m.Capabilities,
		Baseline:     toBaselineDTO(m),
		Providers:    providers,
	})
}

// SyncCatalog POST /admin/model-catalog/sync
// 触发全量 SyncModels；body 中 models[] 非空时对各 slug 额外触发 SyncModelEndpoints。
// 返回 {synced: count}（count 为同步后目录总条目数）。
func (h *ModelCatalogHandler) SyncCatalog(c *gin.Context) {
	var req catalogSyncRequest
	// 请求体可选；忽略解析失败（空 body 合法）。
	_ = c.ShouldBindJSON(&req)

	count, err := h.catalog.SyncModels(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "sync models failed: "+err.Error())
		return
	}

	// 对指定 slug 追加拉取 endpoints（逐个尝试，忽略单个失败避免中断）。
	for _, slug := range req.Models {
		if slug == "" {
			continue
		}
		if endpointErr := h.catalog.SyncModelEndpoints(c.Request.Context(), slug); endpointErr != nil {
			// 记录但不中断；前端可重试单个 slug。
			_ = endpointErr
		}
	}

	response.Success(c, catalogSyncResponse{Synced: count})
}
