package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ModelPriceOverride 管理员对模型「官方价基准」的覆盖。
//
// 取舍优先级（高 → 低）：
//  1. 手动价（manual_*）—— 任一字段非空即覆盖对应项
//  2. 指定供应商（pinned_provider_tag）—— 锁定某 OpenRouter 供应商的价
//  3. 自动最低价 —— OpenRouterCatalogService.BaselinePrice() 默认逻辑
//
// model_id 为匹配键（本地模型名或 OpenRouter slug，硬删除即恢复自动）。
type ModelPriceOverride struct {
	ent.Schema
}

func (ModelPriceOverride) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "model_price_overrides"},
	}
}

func (ModelPriceOverride) Fields() []ent.Field {
	pg := func(t string) map[string]string { return map[string]string{dialect.Postgres: t} }
	return []ent.Field{
		field.String("model_id").
			MaxLen(200).
			NotEmpty().
			Unique(),
		// 指定供应商（②）：锁定某 OpenRouter 供应商 tag 的价；空表示不指定。
		field.String("pinned_provider_tag").
			MaxLen(120).
			Optional(),
		// 手动覆盖价（①，每 token，USD）：Nillable → *float64，区分「未设置」与 0。
		field.Float("manual_input").Optional().Nillable(),
		field.Float("manual_output").Optional().Nillable(),
		field.Float("manual_cache_read").Optional().Nillable(),
		field.Float("manual_cache_write").Optional().Nillable(),
		field.String("note").
			MaxLen(500).
			Optional(),
		field.Int64("updated_by").
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			SchemaType(pg("timestamptz")),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(pg("timestamptz")),
	}
}

func (ModelPriceOverride) Indexes() []ent.Index {
	// model_id 已 Unique，无需额外索引
	return nil
}
