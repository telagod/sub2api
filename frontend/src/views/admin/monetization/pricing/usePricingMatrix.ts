/**
 * usePricingMatrix — PayGo 计价台 数据装配层
 *
 * 数据流：
 *   adminAPI.channels.list (all pages) → 每渠道 model_pricing
 *   adminAPI.groups.getAll           → groups.rate_multiplier
 *   关联：Channel.group_ids → 哪些 group 用哪个渠道定价
 *
 * 矩阵单元格 = 渠道价 × 分组倍率（input/output 两维度）
 * 官方价通过 getModelDefaultPricing 懒加载（hover/展开时按需拉取，避免全表 N 请求）
 */

import { ref, computed } from 'vue'
import channelsAPI from '@/api/admin/channels'
import groupsAPI from '@/api/admin/groups'
import modelCatalogAPI from '@/api/admin/modelCatalog'
import type { Channel, ChannelModelPricing } from '@/api/admin/channels'
import type { AdminGroup } from '@/types'

export interface MatrixCell {
  inputPrice: number | null    // 渠道价 × 分组倍率
  outputPrice: number | null
  cacheWritePrice: number | null
  cacheReadPrice: number | null
  billingMode: string
  hasIntervals: boolean
  intervals: ChannelModelPricing['intervals']
  channelId: number
  channelName: string
}

export interface MatrixRow {
  model: string
  platform: string
  /** groupId → MatrixCell */
  cells: Record<number, MatrixCell>
}

export interface OfficialPricing {
  found: boolean
  inputPrice?: number
  outputPrice?: number
  cacheWritePrice?: number
  cacheReadPrice?: number
  /** Price source tag, e.g. "openrouter", "litellm" */
  source?: string
  /** OpenRouter model slug, e.g. "openai/gpt-4o" */
  slug?: string
  /** Short model description */
  description?: string
  /** Number of providers available (from providers array length) */
  providerCount?: number
}

export function usePricingMatrix() {
  const channels = ref<Channel[]>([])
  const groups = ref<AdminGroup[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 官方价懒缓存 model → OfficialPricing
  const officialPricingCache = ref<Record<string, OfficialPricing | 'loading'>>({})

  async function fetchAll() {
    loading.value = true
    error.value = null
    try {
      // 拉全量渠道（最多 500 条，分页翻完）
      const allChannels: Channel[] = []
      let page = 1
      const pageSize = 100
      while (true) {
        const resp = await channelsAPI.list(page, pageSize)
        allChannels.push(...resp.items)
        if (allChannels.length >= resp.total) break
        page++
      }
      channels.value = allChannels.filter(c => c.model_pricing && c.model_pricing.length > 0)

      // 拉全量 groups
      groups.value = await groupsAPI.getAll()
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e)
    } finally {
      loading.value = false
    }

    // 矩阵加载后立即并发拉取所有 model 官方价（去重，并发上限 5）
    // 这是折扣染色的数据基础，不能懒加载等 hover
    void prefetchAllOfficialPricing()
  }

  /**
   * 进入视口后并发预拉取所有 model 的官方价
   * 并发上限 CONCURRENCY，去重（已缓存的跳过）
   */
  async function prefetchAllOfficialPricing() {
    const CONCURRENCY = 5
    // 收集矩阵中所有不重复 model
    const models = [...new Set(channels.value.flatMap(c => c.model_pricing.flatMap(mp => mp.models)))]
    // 过滤掉已缓存的
    const todo = models.filter(m => !officialPricingCache.value[m])
    // 分批并发
    for (let i = 0; i < todo.length; i += CONCURRENCY) {
      const batch = todo.slice(i, i + CONCURRENCY)
      await Promise.allSettled(batch.map(m => ensureOfficialPricing(m)))
    }
  }

  /**
   * 按需懒加载官方价（hover/展开触发）
   */
  async function ensureOfficialPricing(model: string): Promise<OfficialPricing> {
    const cached = officialPricingCache.value[model]
    if (cached && cached !== 'loading') return cached
    if (cached === 'loading') {
      // 等待已有请求完成（简单轮询 10 次）
      for (let i = 0; i < 20; i++) {
        await new Promise(r => setTimeout(r, 100))
        const c = officialPricingCache.value[model]
        if (c && c !== 'loading') return c
      }
      return { found: false }
    }
    officialPricingCache.value[model] = 'loading'
    try {
      const resp = await channelsAPI.getModelDefaultPricing(model)
      const result: OfficialPricing = {
        found: resp.found,
        inputPrice: resp.input_price,
        outputPrice: resp.output_price,
        cacheWritePrice: resp.cache_write_price,
        cacheReadPrice: resp.cache_read_price,
        source: resp.source,
        slug: resp.slug,
        description: resp.description,
        providerCount: resp.providers?.length ?? 0
      }
      officialPricingCache.value[model] = result
      return result
    } catch {
      const result: OfficialPricing = { found: false }
      officialPricingCache.value[model] = result
      return result
    }
  }

  /**
   * 同步全目录多供应商价格（POST /admin/model-catalog/sync）
   * 成功后失效当前官方价缓存，触发重新拉取
   */
  async function syncCatalog(): Promise<{ synced: number }> {
    const result = await modelCatalogAPI.syncCatalog()
    // 失效缓存：强制下次 hover/prefetch 时重新拉取
    officialPricingCache.value = {}
    void prefetchAllOfficialPricing()
    return result
  }

  /**
   * 矩阵行列计算：行=model，列=group
   * 一个 model 可能被多个渠道覆盖；取最后一个匹配渠道（与渠道列表顺序一致）
   */
  const matrix = computed<MatrixRow[]>(() => {
    // groupId → AdminGroup
    const groupMap = new Map<number, AdminGroup>(groups.value.map(g => [g.id, g]))

    // 收集所有 model 及关联的渠道定价 (model → [{channel, pricing}])
    type PricingEntry = { channel: Channel; pricing: ChannelModelPricing }
    const modelMap = new Map<string, PricingEntry[]>()

    for (const channel of channels.value) {
      for (const mp of channel.model_pricing) {
        for (const model of mp.models) {
          if (!modelMap.has(model)) modelMap.set(model, [])
          modelMap.get(model)!.push({ channel, pricing: mp })
        }
      }
    }

    const rows: MatrixRow[] = []

    for (const [model, entries] of modelMap.entries()) {
      // 每个 entry 的 channel.group_ids 说明这个渠道定价适用的分组
      const cells: Record<number, MatrixCell> = {}

      for (const { channel, pricing } of entries) {
        for (const groupId of channel.group_ids) {
          const group = groupMap.get(groupId)
          if (!group) continue
          const multiplier = group.rate_multiplier ?? 1

          cells[groupId] = {
            inputPrice: pricing.input_price != null ? pricing.input_price * multiplier : null,
            outputPrice: pricing.output_price != null ? pricing.output_price * multiplier : null,
            cacheWritePrice: pricing.cache_write_price != null ? pricing.cache_write_price * multiplier : null,
            cacheReadPrice: pricing.cache_read_price != null ? pricing.cache_read_price * multiplier : null,
            billingMode: pricing.billing_mode,
            hasIntervals: pricing.intervals.length > 0,
            // Blocker fix: apply rate_multiplier to interval tier prices to stay
            // consistent with the top-level cell prices shown in the matrix.
            intervals: pricing.intervals.map(iv => ({
              ...iv,
              input_price: iv.input_price != null ? iv.input_price * multiplier : null,
              output_price: iv.output_price != null ? iv.output_price * multiplier : null,
              cache_write_price: iv.cache_write_price != null ? iv.cache_write_price * multiplier : null,
              cache_read_price: iv.cache_read_price != null ? iv.cache_read_price * multiplier : null
            })),
            channelId: channel.id,
            channelName: channel.name
          }
        }
      }

      if (Object.keys(cells).length === 0) continue

      // 获取 platform（从 entries 的第一个 pricing 拿）
      const platform = entries[0]?.pricing.platform ?? 'unknown'

      rows.push({ model, platform, cells })
    }

    // 按 platform 排序，platform 内按 model 字母排序
    rows.sort((a, b) => {
      const pc = a.platform.localeCompare(b.platform)
      return pc !== 0 ? pc : a.model.localeCompare(b.model)
    })

    return rows
  })

  /**
   * 所有 platform 分组（用于折叠行组）
   */
  const platforms = computed<string[]>(() => {
    const set = new Set(matrix.value.map(r => r.platform))
    return [...set].sort()
  })

  /**
   * 所有 groups（只取有定价关联的分组）
   */
  const activeGroups = computed<AdminGroup[]>(() => {
    const usedIds = new Set<number>()
    for (const row of matrix.value) {
      for (const id of Object.keys(row.cells)) usedIds.add(Number(id))
    }
    return groups.value.filter(g => usedIds.has(g.id))
  })

  /**
   * Optimistic 更新分组倍率（就地编辑列头）
   */
  async function updateGroupMultiplier(groupId: number, newMultiplier: number) {
    // Optimistic: 先更新本地 groups
    const idx = groups.value.findIndex(g => g.id === groupId)
    if (idx === -1) return
    const old = groups.value[idx].rate_multiplier
    groups.value[idx] = { ...groups.value[idx], rate_multiplier: newMultiplier }
    try {
      await groupsAPI.update(groupId, { rate_multiplier: newMultiplier })
    } catch (e) {
      // rollback
      groups.value[idx] = { ...groups.value[idx], rate_multiplier: old }
      throw e
    }
  }

  return {
    channels,
    groups,
    loading,
    error,
    matrix,
    platforms,
    activeGroups,
    officialPricingCache,
    fetchAll,
    ensureOfficialPricing,
    updateGroupMultiplier,
    syncCatalog
  }
}
