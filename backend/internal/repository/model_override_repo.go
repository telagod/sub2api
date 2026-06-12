package repository

import (
	"context"

	dbent "github.com/telagod/subme/ent"
	"github.com/telagod/subme/ent/modelpriceoverride"
	"github.com/telagod/subme/internal/service"
)

type modelOverrideRepository struct {
	client *dbent.Client
}

// NewModelOverrideRepository 创建模型价格覆盖仓储（ent 实现）。
func NewModelOverrideRepository(client *dbent.Client) service.ModelOverrideRepository {
	return &modelOverrideRepository{client: client}
}

func (r *modelOverrideRepository) Get(ctx context.Context, modelID string) (*service.ModelPriceOverride, error) {
	m, err := r.client.ModelPriceOverride.Query().
		Where(modelpriceoverride.ModelIDEQ(modelID)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entToService(m), nil
}

func (r *modelOverrideRepository) Upsert(ctx context.Context, o *service.ModelPriceOverride) error {
	err := r.client.ModelPriceOverride.Create().
		SetModelID(o.ModelID).
		SetNillablePinnedProviderTag(nilIfEmpty(o.PinnedProviderTag)).
		SetNillableManualInput(o.ManualInput).
		SetNillableManualOutput(o.ManualOutput).
		SetNillableManualCacheRead(o.ManualCacheRead).
		SetNillableManualCacheWrite(o.ManualCacheWrite).
		SetNillableNote(nilIfEmpty(o.Note)).
		SetNillableUpdatedBy(nilIfZeroInt64(o.UpdatedBy)).
		OnConflictColumns(modelpriceoverride.FieldModelID).
		UpdateNewValues().
		Exec(ctx)
	return err
}

func (r *modelOverrideRepository) Delete(ctx context.Context, modelID string) error {
	_, err := r.client.ModelPriceOverride.Delete().
		Where(modelpriceoverride.ModelIDEQ(modelID)).
		Exec(ctx)
	return err
}

func (r *modelOverrideRepository) List(ctx context.Context) ([]service.ModelPriceOverride, error) {
	rows, err := r.client.ModelPriceOverride.Query().
		Order(dbent.Asc(modelpriceoverride.FieldModelID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]service.ModelPriceOverride, 0, len(rows))
	for _, m := range rows {
		if s := entToService(m); s != nil {
			out = append(out, *s)
		}
	}
	return out, nil
}

// ── 辅助函数 ──

func entToService(m *dbent.ModelPriceOverride) *service.ModelPriceOverride {
	if m == nil {
		return nil
	}
	return &service.ModelPriceOverride{
		ID:                m.ID,
		ModelID:           m.ModelID,
		PinnedProviderTag: m.PinnedProviderTag,
		ManualInput:       m.ManualInput,
		ManualOutput:      m.ManualOutput,
		ManualCacheRead:   m.ManualCacheRead,
		ManualCacheWrite:  m.ManualCacheWrite,
		Note:              m.Note,
		UpdatedBy:         m.UpdatedBy,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nilIfZeroInt64(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}
