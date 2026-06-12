/**
 * Admin Model Catalog API endpoints
 * Handles OpenRouter model catalog data including per-provider pricing details
 */

import { apiClient } from '../client'

// ==================== Types ====================

export interface CatalogProvider {
  /** Provider identifier (e.g. "openai", "deepseek") */
  provider: string
  /** Provider tag used in OpenRouter slugs (e.g. "openai", "deepseek") */
  tag: string
  /** Input price per token */
  input: number
  /** Output price per token */
  output: number
  /** Cache read price per token (optional, provider may not support) */
  cache_read?: number
  /** Cache write price per token (optional, provider may not support) */
  cache_write?: number
  /** 24-hour uptime ratio 0–1 (optional) */
  uptime_1d?: number
  /** Quantization label e.g. "fp16", "int8" (optional) */
  quant?: string
}

export interface CatalogModelBaseline {
  /** Baseline input price per token */
  input: number
  /** Baseline output price per token */
  output: number
  /** Baseline cache read price per token */
  cache_read?: number
  /** Baseline cache write price per token */
  cache_write?: number
  /** Source tag of the baseline price (e.g. "openrouter", "litellm") */
  source?: string
}

/** Current override record for a model (returned by GET /detail and PUT /override) */
export interface ModelPriceOverride {
  model_id: string
  pinned_provider_tag?: string
  manual_input?: number | null
  manual_output?: number | null
  manual_cache_read?: number | null
  manual_cache_write?: number | null
  note?: string
  updated_by?: number
  updated_at: string
}

export interface CatalogModelDetail {
  /** OpenRouter model slug (e.g. "openai/gpt-4o") */
  id: string
  /** Human-readable model name */
  name: string
  /** Model description */
  description?: string
  /** Maximum context window in tokens */
  context_len?: number
  /** Supported input/output modalities (e.g. ["text", "image"]) */
  modalities?: string[]
  /** Model capability tags (e.g. ["tools", "vision", "streaming"]) */
  capabilities?: string[]
  /** All available providers for this model */
  providers: CatalogProvider[]
  /** Baseline (official) pricing — override-applied final value */
  baseline?: CatalogModelBaseline
  /** Whether an override record exists in the DB */
  overridden?: boolean
  /** Current override record (null/absent = no override) */
  override?: ModelPriceOverride | null
}

export interface SyncModelCatalogResult {
  /** Number of models synced */
  synced: number
}

export interface CatalogModelListItem {
  id: string
  name: string
  description?: string
  context_len?: number
  capabilities?: string[]
}

// ==================== API Functions ====================

/**
 * Get full model detail including all provider pricing.
 * If providers array is empty, call syncModelEndpoints first.
 * @param slug - OpenRouter model slug, e.g. "openai/gpt-4o"
 */
export async function getModelCatalogDetail(slug: string): Promise<CatalogModelDetail> {
  const { data } = await apiClient.get<CatalogModelDetail>('/admin/model-catalog/detail', {
    params: { model: slug }
  })
  return data
}

/**
 * Trigger a lazy sync for a specific model's provider endpoints.
 * Call when getModelCatalogDetail returns an empty providers array.
 * @param slug - OpenRouter model slug
 */
export async function syncModelEndpoints(slug: string): Promise<void> {
  await apiClient.post('/admin/model-catalog/sync', { models: [slug] })
}

/**
 * Sync the entire model catalog (all models).
 * Returns the number of models synced.
 */
export async function syncCatalog(): Promise<SyncModelCatalogResult> {
  const { data } = await apiClient.post<SyncModelCatalogResult>('/admin/model-catalog/sync', {})
  return data
}

/**
 * List all models in the catalog (summary view).
 */
export async function listModelCatalog(): Promise<CatalogModelListItem[]> {
  const { data } = await apiClient.get<CatalogModelListItem[]>('/admin/model-catalog')
  return data
}

/** PUT /admin/model-catalog/override — create or update a price override */
export interface UpsertOverridePayload {
  model_id: string
  pinned_provider_tag?: string
  /** per-token (USD) — UI collects per-MTok, divide by 1e6 before sending */
  manual_input?: number | null
  manual_output?: number | null
  manual_cache_read?: number | null
  manual_cache_write?: number | null
  note?: string
}

/**
 * Create or update a model price override.
 * Returns the saved override record.
 */
export async function putModelOverride(payload: UpsertOverridePayload): Promise<ModelPriceOverride> {
  const { data } = await apiClient.put<ModelPriceOverride>('/admin/model-catalog/override', payload)
  return data
}

/**
 * Delete a model price override, restoring auto pricing.
 * @param model_id - OpenRouter model slug / model_id
 */
export async function deleteModelOverride(model_id: string): Promise<void> {
  await apiClient.delete('/admin/model-catalog/override', { params: { model: model_id } })
}

const modelCatalogAPI = {
  getModelCatalogDetail,
  syncModelEndpoints,
  syncCatalog,
  listModelCatalog,
  putModelOverride,
  deleteModelOverride
}

export default modelCatalogAPI
