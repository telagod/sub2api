/**
 * useTableUrlState — 把 TableQueryState 双向同步到 route.query
 * replace 模式防历史爆炸；数字/字符串/数组序列化处理干净
 * 淬钢 QUENCH · 账房表格底座
 */
import { reactive, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { TableQueryState, SortOrder } from './types'

const DEFAULT_STATE: TableQueryState = {
  page: 1,
  pageSize: 20,
  sort: '',
  order: 'asc',
  q: '',
  filters: {}
}

/**
 * 从 route.query 反序列化成 TableQueryState
 */
function parseQuery(query: Record<string, string | string[] | null | undefined>): TableQueryState {
  const page = query.page ? parseInt(String(query.page), 10) : DEFAULT_STATE.page
  const pageSize = query.pageSize ? parseInt(String(query.pageSize), 10) : DEFAULT_STATE.pageSize
  const sort = query.sort ? String(query.sort) : DEFAULT_STATE.sort
  const order: SortOrder = (query.order === 'desc' ? 'desc' : 'asc') as SortOrder
  const q = query.q ? String(query.q) : DEFAULT_STATE.q

  // 解析 filters：以 f_ 为前缀的 key
  const filters: Record<string, string | string[]> = {}
  for (const [k, v] of Object.entries(query)) {
    if (k.startsWith('f_') && v != null) {
      const field = k.slice(2)
      filters[field] = Array.isArray(v) ? v : String(v)
    }
  }

  return {
    page: isNaN(page) ? DEFAULT_STATE.page : page,
    pageSize: isNaN(pageSize) ? DEFAULT_STATE.pageSize : pageSize,
    sort,
    order,
    q,
    filters
  }
}

/**
 * 把 TableQueryState 序列化成 route.query 字典
 * 默认值不写入 URL，保持 URL 干净
 */
function serializeState(s: TableQueryState): Record<string, string | string[]> {
  const q: Record<string, string | string[]> = {}

  if (s.page !== DEFAULT_STATE.page) q.page = String(s.page)
  if (s.pageSize !== DEFAULT_STATE.pageSize) q.pageSize = String(s.pageSize)
  if (s.sort) q.sort = s.sort
  if (s.order !== DEFAULT_STATE.order) q.order = s.order
  if (s.q) q.q = s.q

  for (const [k, v] of Object.entries(s.filters)) {
    if (v != null && v !== '' && !(Array.isArray(v) && v.length === 0)) {
      q[`f_${k}`] = Array.isArray(v) ? v : String(v)
    }
  }

  return q
}

export function useTableUrlState(namespace?: string) {
  const route = useRoute()
  const router = useRouter()

  // 支持命名空间：多表格同页时隔离 key
  function nsKey(key: string) {
    return namespace ? `${namespace}_${key}` : key
  }

  function buildNamespacedQuery(raw: Record<string, string | string[]>): Record<string, string | string[]> {
    if (!namespace) return raw
    const out: Record<string, string | string[]> = {}
    for (const [k, v] of Object.entries(raw)) {
      out[nsKey(k)] = v
    }
    return out
  }

  function parseNamespacedQuery(query: Record<string, string | string[] | null | undefined>): Record<string, string | string[] | null | undefined> {
    if (!namespace) return query
    const out: Record<string, string | string[] | null | undefined> = {}
    for (const [k, v] of Object.entries(query)) {
      if (k.startsWith(`${namespace}_`)) {
        out[k.slice(namespace.length + 1)] = v
      }
    }
    return out
  }

  // 初始化从 URL 读取状态
  const state = reactive<TableQueryState>(parseQuery(parseNamespacedQuery(route.query as Record<string, string | string[] | null | undefined>)))

  // 监听 state 变化，replace 写入 URL
  watch(
    () => ({ ...state, filters: { ...state.filters } }),
    (val) => {
      const serialized = buildNamespacedQuery(serializeState(val))
      // 保留非本组件管理的 query 参数
      const existing: Record<string, string | string[]> = {}
      for (const [k, v] of Object.entries(route.query)) {
        if (v != null) {
          existing[k] = Array.isArray(v) ? v as string[] : String(v)
        }
      }
      // 清除旧的本组件 key（先清再写，防残留）
      if (namespace) {
        for (const k of Object.keys(existing)) {
          if (k.startsWith(`${namespace}_`)) delete existing[k]
        }
      } else {
        // 清掉本组件管理的固定 key
        for (const k of ['page', 'pageSize', 'sort', 'order', 'q'] as const) {
          delete existing[k]
        }
        // 清掉旧 filter keys
        for (const k of Object.keys(existing)) {
          if (k.startsWith('f_')) delete existing[k]
        }
      }
      const merged = { ...existing, ...serialized }
      router.replace({ query: merged })
    },
    { deep: true }
  )

  // 监听 route.query 变化（浏览器前进/后退）
  watch(
    () => route.query,
    (q) => {
      const parsed = parseQuery(parseNamespacedQuery(q as Record<string, string | string[] | null | undefined>))
      // 仅当值真正不同时才更新，避免循环触发
      if (parsed.page !== state.page) state.page = parsed.page
      if (parsed.pageSize !== state.pageSize) state.pageSize = parsed.pageSize
      if (parsed.sort !== state.sort) state.sort = parsed.sort
      if (parsed.order !== state.order) state.order = parsed.order
      if (parsed.q !== state.q) state.q = parsed.q
      // filters 做 JSON 简单对比
      if (JSON.stringify(parsed.filters) !== JSON.stringify(state.filters)) {
        state.filters = parsed.filters
      }
    }
  )

  function reset() {
    state.page = DEFAULT_STATE.page
    state.pageSize = DEFAULT_STATE.pageSize
    state.sort = DEFAULT_STATE.sort
    state.order = DEFAULT_STATE.order
    state.q = DEFAULT_STATE.q
    state.filters = {}
  }

  return { state, reset }
}
