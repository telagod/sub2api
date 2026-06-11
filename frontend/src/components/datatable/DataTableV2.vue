<template>
  <div class="q-table-wrap" :style="cssVars">
    <!-- ── 骨架态 ── -->
    <template v-if="loading">
      <table class="q-dt">
        <thead>
          <tr>
            <th v-if="selectable" class="q-th q-th-cb"></th>
            <th
              v-for="col in columns"
              :key="col.key"
              class="q-th"
              :class="[`q-align-${col.align ?? 'left'}`]"
              :style="col.width ? { width: col.width } : {}"
            >
              {{ col.title }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="i in 5" :key="i" class="q-tr-skeleton">
            <td v-if="selectable" class="q-td q-td-cb">
              <div class="q-skel q-skel-cb"></div>
            </td>
            <td v-for="col in columns" :key="col.key" class="q-td">
              <div class="q-skel" :style="{ width: col.align === 'right' ? '60%' : '75%', marginLeft: col.align === 'right' ? 'auto' : undefined }"></div>
            </td>
          </tr>
        </tbody>
      </table>
    </template>

    <!-- ── 空态 ── -->
    <template v-else-if="!rows || rows.length === 0">
      <slot name="empty">
        <div class="q-empty">
          <svg width="40" height="40" viewBox="0 0 40 40" fill="none" aria-hidden="true">
            <rect x="6" y="8" width="28" height="24" rx="4" stroke="currentColor" stroke-width="1.5"/>
            <line x1="12" y1="16" x2="28" y2="16" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            <line x1="12" y1="22" x2="22" y2="22" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
          </svg>
          <span>暂无数据</span>
        </div>
      </slot>
    </template>

    <!-- ── 正文表格 ── -->
    <template v-else>
      <table class="q-dt">
        <thead>
          <tr>
            <!-- 全选 checkbox -->
            <th v-if="selectable" class="q-th q-th-cb">
              <input
                type="checkbox"
                class="q-cbox"
                :checked="isAllSelected"
                :indeterminate="isIndeterminate"
                aria-label="全选"
                @change="onToggleAll"
              />
            </th>
            <!-- 列头 -->
            <th
              v-for="col in columns"
              :key="col.key"
              class="q-th"
              :class="[
                `q-align-${col.align ?? 'left'}`,
                col.sortable ? 'q-th-sort' : ''
              ]"
              :style="col.width ? { width: col.width } : {}"
              :tabindex="col.sortable ? 0 : undefined"
              :role="col.sortable ? 'button' : undefined"
              :aria-sort="sortAriaAttr(col)"
              @click="col.sortable && onSort(col.key)"
              @keydown.enter="col.sortable && onSort(col.key)"
              @keydown.space.prevent="col.sortable && onSort(col.key)"
            >
              <span class="q-th-inner">
                {{ col.title }}
                <span v-if="col.sortable" class="q-sort-icon" aria-hidden="true">
                  <svg v-if="currentSort === col.key && currentOrder === 'asc'" width="10" height="10" viewBox="0 0 10 10"><path d="M5 2 L8 7 L2 7 Z" fill="currentColor"/></svg>
                  <svg v-else-if="currentSort === col.key && currentOrder === 'desc'" width="10" height="10" viewBox="0 0 10 10"><path d="M5 8 L8 3 L2 3 Z" fill="currentColor"/></svg>
                  <svg v-else width="10" height="10" viewBox="0 0 10 10" opacity=".35"><path d="M5 2 L7 5 L3 5 Z" fill="currentColor"/><path d="M5 8 L7 5 L3 5 Z" fill="currentColor"/></svg>
                </span>
              </span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(row, idx) in rows"
            :key="resolveRowKey(row, idx)"
            class="q-tr"
            :class="{ 'q-tr-sel': isSelected(row) }"
            tabindex="0"
            @click="onRowClick(row, idx)"
            @keydown.enter="onRowClick(row, idx)"
          >
            <!-- 行选 checkbox -->
            <td v-if="selectable" class="q-td q-td-cb" @click.stop>
              <input
                type="checkbox"
                class="q-cbox"
                :checked="isSelected(row)"
                :aria-label="`选择行 ${idx + 1}`"
                @change="onToggleRow(row)"
              />
            </td>
            <!-- 单元格 -->
            <td
              v-for="col in columns"
              :key="col.key"
              class="q-td"
              :class="[`q-align-${col.align ?? 'left'}`, col.cellClass ?? '']"
            >
              <!-- 具名插槽优先，fallback 默认渲染 -->
              <slot :name="`cell-${col.key}`" :row="row" :value="getCellValue(row, col.key)" :index="idx">
                <span :class="col.cellClass ?? ''">{{ getCellValue(row, col.key) }}</span>
              </slot>
            </td>
          </tr>
        </tbody>
      </table>
    </template>

    <!-- ── 底部分页栏 ── -->
    <div v-if="!loading" class="q-tfoot">
      <span class="q-tfoot-info">
        共 <b class="q-mono">{{ total.toLocaleString() }}</b> 条
        <template v-if="selectable && selectedRows.length > 0">
          · 已选 <b class="q-mono q-azure">{{ selectedRows.length }}</b>
        </template>
      </span>
      <!-- 页码 -->
      <nav class="q-pages" aria-label="分页">
        <button
          class="q-pg"
          :disabled="currentPage <= 1"
          aria-label="上一页"
          @click="onPageChange(currentPage - 1)"
        >‹</button>
        <button
          v-for="p in pageNumbers"
          :key="p"
          class="q-pg"
          :class="{ 'q-pg-on': p === currentPage, 'q-pg-ellipsis': p === -1 }"
          :disabled="p === -1"
          :aria-current="p === currentPage ? 'page' : undefined"
          @click="p !== -1 && onPageChange(p)"
        >
          {{ p === -1 ? '…' : p }}
        </button>
        <button
          class="q-pg"
          :disabled="currentPage >= totalPages"
          aria-label="下一页"
          @click="onPageChange(currentPage + 1)"
        >›</button>
      </nav>
    </div>
  </div>
</template>

<script setup lang="ts" generic="T extends Record<string, unknown>">
import { computed, shallowRef } from 'vue'
import type { ColumnDef, SortOrder } from './types'

// ── Props ──────────────────────────────────────────────────────────────
const props = withDefaults(defineProps<{
  columns: ColumnDef<T>[]
  rows: T[]
  total: number
  loading?: boolean
  selectable?: boolean
  rowKey?: keyof T & string
  density?: 'comfortable' | 'compact'
  /** 当前页（受控），不传则组件内部管理 */
  page?: number
  /** 每页条数，默认 20 */
  pageSize?: number
  /** 当前排序字段 */
  sort?: string
  /** 当前排序方向 */
  order?: SortOrder
}>(), {
  loading: false,
  selectable: false,
  density: 'compact',
  page: 1,
  pageSize: 20,
  sort: '',
  order: 'asc'
})

// ── Emits ──────────────────────────────────────────────────────────────
const emit = defineEmits<{
  'row-click': [row: T, index: number]
  'update:selected': [rows: T[]]
  'update:page': [page: number]
  'update:sort': [sort: string]
  'update:order': [order: SortOrder]
}>()

// ── 内部状态 ────────────────────────────────────────────────────────────
const selectedRows = shallowRef<T[]>([])

const currentPage = computed(() => props.page)
const currentSort = computed(() => props.sort)
const currentOrder = computed(() => props.order)
const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))

// ── CSS 变量（密度控行高） ─────────────────────────────────────────────
const cssVars = computed(() => ({
  '--q-row-h': props.density === 'comfortable' ? '44px' : '32px'
}))

// ── 工具 ───────────────────────────────────────────────────────────────
function resolveRowKey(row: T, idx: number): string {
  if (props.rowKey && row[props.rowKey] != null) {
    return String(row[props.rowKey])
  }
  return String(idx)
}

function getCellValue(row: T, key: string): unknown {
  return row[key]
}

function isSelected(row: T): boolean {
  const key = props.rowKey
  if (key) return selectedRows.value.some(r => r[key] === row[key])
  return selectedRows.value.includes(row)
}

const isAllSelected = computed(() =>
  props.rows.length > 0 && props.rows.every(r => isSelected(r))
)

const isIndeterminate = computed(() =>
  !isAllSelected.value && props.rows.some(r => isSelected(r))
)

// ── 事件处理 ───────────────────────────────────────────────────────────
function onRowClick(row: T, idx: number) {
  emit('row-click', row, idx)
}

function onToggleRow(row: T) {
  const key = props.rowKey
  const current = selectedRows.value
  const idx = key
    ? current.findIndex(r => r[key] === row[key])
    : current.indexOf(row)
  if (idx >= 0) {
    const next = [...current]
    next.splice(idx, 1)
    selectedRows.value = next
  } else {
    selectedRows.value = [...current, row]
  }
  emit('update:selected', [...selectedRows.value])
}

function onToggleAll(e: Event) {
  const checked = (e.target as HTMLInputElement).checked
  selectedRows.value = checked ? [...props.rows] : []
  emit('update:selected', [...selectedRows.value])
}

function onSort(key: string) {
  if (currentSort.value === key) {
    emit('update:order', currentOrder.value === 'asc' ? 'desc' : 'asc')
  } else {
    emit('update:sort', key)
    emit('update:order', 'asc')
  }
  emit('update:page', 1)
}

function onPageChange(page: number) {
  if (page < 1 || page > totalPages.value) return
  emit('update:page', page)
}

function sortAriaAttr(col: ColumnDef<T>): 'ascending' | 'descending' | 'none' | undefined {
  if (!col.sortable) return undefined
  if (currentSort.value !== col.key) return 'none'
  return currentOrder.value === 'asc' ? 'ascending' : 'descending'
}

// ── 页码序列（含省略号 -1） ────────────────────────────────────────────
const pageNumbers = computed<number[]>(() => {
  const total = totalPages.value
  const cur = currentPage.value
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }
  const pages: number[] = [1]
  if (cur > 3) pages.push(-1)
  const start = Math.max(2, cur - 1)
  const end = Math.min(total - 1, cur + 1)
  for (let i = start; i <= end; i++) pages.push(i)
  if (cur < total - 2) pages.push(-1)
  pages.push(total)
  return pages
})
</script>

<style scoped>
/* ── 淬钢 QUENCH · DataTableV2 样式 ── */
.q-table-wrap {
  --q-row-h: 32px; /* compact default，被 density prop 覆盖 */
  width: 100%;
  font-family: var(--font-ui, "Archivo", "PingFang SC", sans-serif);
  font-size: 12.5px;
  color: var(--ink-0, #E8EBF0);
}

/* 表格结构 */
.q-dt {
  width: 100%;
  border-collapse: collapse;
}

/* 表头 */
.q-th {
  text-align: left;
  font-size: 10.5px;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--ink-2, #5C6470);
  padding: 9px 14px;
  border-bottom: 1px solid var(--line-0, #20242C);
  background: transparent;
  position: sticky;
  top: 0;
  white-space: nowrap;
  user-select: none;
  z-index: 1;
}

.q-th-cb {
  width: 34px;
  padding: 9px 10px;
}

.q-th-sort {
  cursor: pointer;
  transition: color 0.15s;
}

.q-th-sort:hover {
  color: var(--ink-0, #E8EBF0);
}

.q-th-sort:focus-visible {
  outline: none;
  box-shadow: var(--glow-focus, 0 0 0 1.5px rgba(92, 168, 255, 0.65), 0 0 20px rgba(92, 168, 255, 0.28));
  border-radius: 4px;
}

.q-th-inner {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.q-sort-icon {
  display: inline-flex;
  align-items: center;
  color: var(--azure, #5CA8FF);
  flex-shrink: 0;
}

/* 行 */
.q-tr {
  cursor: pointer;
  transition: background 0.12s;
  position: relative;
}

.q-tr:hover {
  background: var(--bg-2, #171A20);
}

/* 首列淬火蓝左缘线（行 hover 时） */
.q-tr:hover .q-td:first-child,
.q-tr:hover .q-td-cb + .q-td {
  box-shadow: inset 2px 0 0 var(--azure, #5CA8FF);
}

/* selectable 时：首列是 cb，第二列出左缘线 */
.q-tr:hover .q-td:nth-child(2) {
  box-shadow: none; /* 重置，由下面具体选择器处理 */
}

/* 当 selectable 时 hover 在 cb 之后的第一列出左缘线 */
:where(.q-table-wrap[data-selectable]) .q-tr:hover .q-td:nth-child(2) {
  box-shadow: inset 2px 0 0 var(--azure, #5CA8FF);
}

.q-tr-sel {
  background: var(--azure-dim, rgba(92, 168, 255, 0.12));
}

.q-tr:focus-visible {
  outline: none;
  box-shadow: inset 0 0 0 2px var(--azure, #5CA8FF);
}

/* 单元格 */
.q-td {
  padding: 0 14px;
  height: var(--q-row-h);
  border-bottom: 1px solid var(--line-0, #20242C);
  vertical-align: middle;
  font-size: 12.5px;
}

.q-td-cb {
  width: 34px;
  padding: 0 10px;
}

/* 对齐 */
.q-align-left { text-align: left; }
.q-align-center { text-align: center; }
.q-align-right { text-align: right; }

th.q-align-right .q-th-inner { justify-content: flex-end; }

/* checkbox 样式 */
.q-cbox {
  width: 14px;
  height: 14px;
  accent-color: var(--azure, #5CA8FF);
  cursor: pointer;
  display: block;
}

/* 骨架行 */
.q-tr-skeleton .q-td {
  padding: 0 14px;
  height: var(--q-row-h);
  border-bottom: 1px solid var(--line-0, #20242C);
}

.q-skel {
  height: 12px;
  border-radius: 4px;
  background: var(--bg-2, #171A20);
  animation: q-skel-pulse 1.6s ease-in-out infinite;
}

.q-skel-cb {
  width: 14px;
  height: 14px;
  border-radius: 3px;
}

@keyframes q-skel-pulse {
  0%, 100% { opacity: 0.5; }
  50% { opacity: 1; }
}

/* 空态 */
.q-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 64px 24px;
  color: var(--ink-2, #5C6470);
  font-size: 13px;
}

/* 底部分页栏 */
.q-tfoot {
  display: flex;
  align-items: center;
  padding: 11px 16px;
  color: var(--ink-2, #5C6470);
  font-size: 11.5px;
  gap: 14px;
  border-top: 1px solid var(--line-0, #20242C);
}

.q-tfoot-info {
  color: var(--ink-2, #5C6470);
}

.q-mono {
  font-family: var(--font-mono, "IBM Plex Mono", monospace);
  font-variant-numeric: tabular-nums;
  color: var(--ink-0, #E8EBF0);
}

.q-azure {
  color: var(--azure, #5CA8FF) !important;
}

.q-pages {
  margin-left: auto;
  display: flex;
  gap: 4px;
  align-items: center;
}

.q-pg {
  min-width: 26px;
  height: 26px;
  border-radius: 7px;
  display: grid;
  place-items: center;
  cursor: pointer;
  font-family: var(--font-mono, "IBM Plex Mono", monospace);
  font-size: 11px;
  color: var(--ink-1, #97A0AF);
  background: transparent;
  border: none;
  padding: 0 4px;
  transition: background 0.12s, color 0.12s;
}

.q-pg:hover:not(:disabled):not(.q-pg-ellipsis) {
  background: var(--bg-2, #171A20);
  color: var(--ink-0, #E8EBF0);
}

.q-pg-on {
  background: var(--azure-dim, rgba(92, 168, 255, 0.12)) !important;
  color: var(--azure, #5CA8FF) !important;
  font-weight: 600;
}

.q-pg:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.q-pg-ellipsis {
  cursor: default;
}

/* 金额约定 class（工作包 A 提供全局变量，此处兜底） */
:global(.q-money) {
  font-family: var(--font-mono, "IBM Plex Mono", monospace);
  font-variant-numeric: tabular-nums;
  color: var(--money, #F2F5FA);
  text-shadow: var(--money-glow, 0 0 18px rgba(214, 232, 255, 0.22));
}

@media (prefers-reduced-motion: reduce) {
  .q-skel { animation: none; }
  .q-tr, .q-pg, .q-th-sort { transition: none; }
}
</style>
