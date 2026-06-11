/**
 * DataTable v2 共享类型定义
 * 淬钢 QUENCH · 账房表格底座
 */

/**
 * 列定义 — 泛型 T 代表行数据类型
 * render 通过具名插槽 cell-<key> 覆盖；
 * 金额列约定 cellClass 传入 'q-money'
 */
export interface ColumnDef<T = unknown> {
  /** 对应行数据的字段 key，同时作为插槽名 cell-<key> */
  key: keyof T & string
  /** 表头显示文字 */
  title: string
  /** 对齐方式，默认 left */
  align?: 'left' | 'center' | 'right'
  /** 列宽（CSS 值，如 '120px' / '10%'），不设则自适应 */
  width?: string
  /** 是否可排序 */
  sortable?: boolean
  /** 单元格追加 class（金额列传 'q-money'） */
  cellClass?: string
}

/** 排序方向 */
export type SortOrder = 'asc' | 'desc'

/**
 * 表格查询状态 — 与 URL query 双向同步
 */
export interface TableQueryState {
  page: number
  pageSize: number
  sort: string
  order: SortOrder
  /** 通用关键词搜索 */
  q: string
  /** 任意额外过滤条件，key → 单值或多值 */
  filters: Record<string, string | string[]>
}

/**
 * 保存视图 — 存 localStorage
 */
export interface SavedView {
  id: string
  name: string
  /** 保存时的完整查询状态快照 */
  state: Partial<TableQueryState>
}

/** 行密度模式 */
export type DensityMode = 'comfortable' | 'compact'

/** DataTableV2 emit 事件类型 */
export interface DataTableV2Emits<T> {
  'row-click': [row: T, index: number]
  'update:selected': [rows: T[]]
}
