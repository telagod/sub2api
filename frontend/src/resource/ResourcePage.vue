<template>
  <div class="rp-wrap">
    <!-- ── 页头 ── -->
    <div class="rp-header">
      <h1 class="rp-title">{{ resource.title }}</h1>
      <button
        v-if="resource.api.create"
        class="rp-btn rp-btn-primary"
        @click="openCreateDrawer"
      >
        <svg width="13" height="13" viewBox="0 0 13 13" fill="none" aria-hidden="true">
          <path d="M6.5 2v9M2 6.5h9" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/>
        </svg>
        新建
      </button>
    </div>

    <!-- ── SavedViewTabs ── -->
    <SavedViewTabs
      :storage-key="resource.key"
      :current-state="state"
      :total-count="total"
      @apply="onApplyView"
    />

    <!-- ── 筛选条 ── -->
    <div v-if="resource.filters && resource.filters.length > 0" class="rp-filters">
      <!-- 关键词搜索 -->
      <input
        v-model="state.q"
        type="search"
        class="rp-search"
        placeholder="搜索…"
        @input="onSearchInput"
      />

      <!-- 下拉/文本过滤 -->
      <template v-for="filter in resource.filters" :key="filter.key">
        <select
          v-if="filter.type === 'select'"
          v-model="filterValues[filter.key]"
          class="rp-filter-select"
          @change="onFilterChange"
        >
          <option value="">{{ filter.placeholder ?? filter.label }}</option>
          <option
            v-for="opt in (filter.options ?? [])"
            :key="String(opt.value)"
            :value="opt.value"
          >{{ opt.label }}</option>
        </select>
        <input
          v-else
          v-model="filterValues[filter.key]"
          type="text"
          class="rp-search"
          :placeholder="filter.placeholder ?? filter.label"
          @input="onFilterChange"
        />
      </template>

      <!-- 清除筛选 -->
      <button
        v-if="hasActiveFilters"
        class="rp-btn rp-btn-ghost rp-btn-sm"
        @click="clearFilters"
      >清除</button>
    </div>

    <!-- ── DataTableV2 ── -->
    <DataTableV2
      :columns="resource.columns"
      :rows="rows as any[]"
      :total="total"
      :loading="loading"
      :selectable="hasBulkActions"
      :row-key="resource.rowKey ?? 'id'"
      :page="state.page"
      :page-size="state.pageSize"
      :sort="state.sort"
      :order="state.order"
      @update:page="(p) => { state.page = p; loadData() }"
      @update:sort="(s) => { state.sort = s; loadData() }"
      @update:order="(o) => { state.order = o; loadData() }"
      @update:selected="onSelectedChange"
    >
      <!-- 透传行操作插槽 -->
      <template v-if="resource.rowActions && resource.rowActions.length > 0" #[`cell-${rowActionsKey}`]="{ row }">
        <div class="rp-actions">
          <template v-for="action in resource.rowActions" :key="action.key">
            <button
              v-if="!action.hidden || !action.hidden(row as any)"
              class="rp-action-btn"
              :class="{ 'rp-action-danger': action.danger }"
              @click.stop="action.handler(row as any)"
            >{{ action.label }}</button>
          </template>
        </div>
      </template>

      <!-- 透传其他具名单元格插槽（排除已由行操作插槽处理的 key） -->
      <template
        v-for="col in nonActionsColumns"
        :key="col.key"
        #[`cell-${col.key}`]="slotProps"
      >
        <slot :name="`cell-${col.key}`" v-bind="slotProps">
          <span>{{ slotProps.value }}</span>
        </slot>
      </template>

      <template #empty>
        <div class="rp-empty">
          <svg width="36" height="36" viewBox="0 0 36 36" fill="none" aria-hidden="true">
            <rect x="5" y="7" width="26" height="22" rx="4" stroke="currentColor" stroke-width="1.5"/>
            <line x1="11" y1="15" x2="25" y2="15" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            <line x1="11" y1="21" x2="19" y2="21" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
          </svg>
          <span>暂无数据</span>
        </div>
      </template>
    </DataTableV2>

    <!-- ── 错误提示 ── -->
    <div v-if="error" class="rp-error">{{ error }}</div>

    <!-- ── BulkBar ── -->
    <BulkBar
      v-if="hasBulkActions"
      :count="selectedRows.length"
      @clear="selectedRows = []"
    >
      <template v-for="action in resource.bulkActions" :key="action.key">
        <button
          class="rp-bulk-btn"
          :class="{ 'q-btn-danger': action.danger }"
          @click="action.handler(selectedRows as any[])"
        >{{ action.label }}</button>
      </template>
    </BulkBar>

    <!-- ── 新建/编辑抽屉 ── -->
    <ResourceFormDrawer
      v-model="drawerOpen"
      :title="editingRow ? `编辑 ${resource.title}` : `新建 ${resource.title}`"
      :fields="resource.form ?? []"
      :initial-data="editingRow ?? undefined"
      :submitting="submitting"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch, onMounted } from 'vue'
import { DataTableV2, SavedViewTabs, BulkBar, useTableUrlState } from '@/components/datatable'
import type { SavedView } from '@/components/datatable/types'
import ResourceFormDrawer from './ResourceFormDrawer.vue'
import type { ResourceDef } from './types'

// ── Props ──────────────────────────────────────────────────────────────
const props = defineProps<{
  resource: ResourceDef
}>()

// ── URL 状态 ───────────────────────────────────────────────────────────
const { state, reset: resetState } = useTableUrlState(props.resource.key)

// ── 筛选值（filter 字段 → state.filters 同步） ─────────────────────────
const filterValues = reactive<Record<string, string>>({})

// 初始化 filterValues 从 state.filters
for (const filter of props.resource.filters ?? []) {
  const val = state.filters[filter.key]
  filterValues[filter.key] = Array.isArray(val) ? val[0] ?? '' : val ?? ''
}

// filterValues → state.filters 同步
watch(filterValues, () => {
  const f: Record<string, string> = {}
  for (const key of Object.keys(filterValues)) {
    if (filterValues[key]) f[key] = filterValues[key]
  }
  state.filters = f
}, { deep: true })

const hasActiveFilters = computed(() =>
  state.q !== '' || Object.values(filterValues).some(v => v !== '')
)

function clearFilters() {
  state.q = ''
  for (const key of Object.keys(filterValues)) {
    filterValues[key] = ''
  }
  state.page = 1
  loadData()
}

// ── 数据 ───────────────────────────────────────────────────────────────
const rows = ref<Record<string, unknown>[]>([])
const total = ref(0)
const loading = ref(false)
const error = ref<string | null>(null)
const selectedRows = ref<Record<string, unknown>[]>([])

const hasBulkActions = computed(() =>
  (props.resource.bulkActions?.length ?? 0) > 0
)

// 操作列 key（约定：最后一列且 key === 'actions'）
const rowActionsKey = computed(() => {
  const last = props.resource.columns[props.resource.columns.length - 1]
  return last?.key ?? 'actions'
})

// 排除行操作列，避免两个插槽定义同一 key 冲突
const nonActionsColumns = computed(() =>
  props.resource.columns.filter(
    col => !(props.resource.rowActions?.length && col.key === rowActionsKey.value)
  )
)

async function loadData() {
  loading.value = true
  error.value = null
  try {
    const result = await props.resource.api.list({
      page: state.page,
      pageSize: state.pageSize,
      sort: state.sort || undefined,
      order: state.order,
      q: state.q || undefined,
      filters: state.filters
    })
    rows.value = result.items as Record<string, unknown>[]
    total.value = result.total
  } catch (e: unknown) {
    error.value = '数据加载失败'
    console.error('[ResourcePage] load error:', e)
  } finally {
    loading.value = false
  }
}

// 监听 state 变化（page/sort/order/q/filters）自动加载
watch(
  () => ({ ...state, filters: { ...state.filters } }),
  () => loadData(),
  { deep: true }
)

onMounted(() => loadData())

// ── SavedViewTabs 回调 ──────────────────────────────────────────────────
function onApplyView(view: SavedView | null) {
  if (!view) {
    resetState()
  } else {
    const s = view.state
    if (s.page) state.page = s.page
    if (s.pageSize) state.pageSize = s.pageSize
    if (s.sort) state.sort = s.sort
    if (s.order) state.order = s.order
    if (s.q != null) state.q = s.q
    if (s.filters) state.filters = { ...s.filters }
  }
}

// ── 筛选/搜索 ──────────────────────────────────────────────────────────
function onSearchInput() {
  state.page = 1
}

function onFilterChange() {
  state.page = 1
}

// ── 行选择 ────────────────────────────────────────────────────────────
function onSelectedChange(r: Record<string, unknown>[]) {
  selectedRows.value = r
}

// ── 抽屉 ──────────────────────────────────────────────────────────────
const drawerOpen = ref(false)
const editingRow = ref<Record<string, unknown> | null>(null)
const submitting = ref(false)

function openCreateDrawer() {
  editingRow.value = null
  drawerOpen.value = true
}

// 外部通过 expose 打开编辑
function openEditDrawer(row: Record<string, unknown>) {
  editingRow.value = row
  drawerOpen.value = true
}

async function handleFormSubmit(data: Record<string, unknown>) {
  submitting.value = true
  try {
    if (editingRow.value) {
      const id = editingRow.value[props.resource.rowKey ?? 'id'] as number
      await props.resource.api.update?.(id, data)
    } else {
      await props.resource.api.create?.(data)
    }
    drawerOpen.value = false
    await loadData()
  } catch (e) {
    console.error('[ResourcePage] submit error:', e)
  } finally {
    submitting.value = false
  }
}

// ── expose 给外部使用（可选） ─────────────────────────────────────────
defineExpose({ openEditDrawer, reload: loadData })
</script>

<style scoped>
/* ── 淬钢 QUENCH · ResourcePage ── */
.rp-wrap {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 22px 24px;
  gap: 16px;
  font-family: var(--font-ui, "Archivo", "PingFang SC", sans-serif);
  font-size: 13px;
  color: var(--ink-0, #E8EBF0);
  box-sizing: border-box;
}

/* 页头 */
.rp-header {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-shrink: 0;
}

.rp-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--ink-0, #E8EBF0);
  letter-spacing: -0.01em;
  margin: 0;
  flex: 1;
}

/* 筛选条 */
.rp-filters {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  flex-shrink: 0;
}

.rp-search {
  background: var(--bg-2, #171A20);
  border: 1px solid var(--line-1, #2F3540);
  border-radius: 8px;
  padding: 7px 12px;
  font-size: 12.5px;
  color: var(--ink-0, #E8EBF0);
  outline: none;
  width: 200px;
  transition: border-color 0.15s, box-shadow 0.15s;
  font-family: inherit;
}

.rp-search:focus {
  border-color: rgba(92, 168, 255, 0.55);
  box-shadow: 0 0 0 2px rgba(92, 168, 255, 0.12);
}

.rp-search::placeholder {
  color: var(--ink-2, #5C6470);
}

.rp-filter-select {
  background: var(--bg-2, #171A20);
  border: 1px solid var(--line-1, #2F3540);
  border-radius: 8px;
  padding: 7px 30px 7px 12px;
  font-size: 12.5px;
  color: var(--ink-0, #E8EBF0);
  outline: none;
  cursor: pointer;
  font-family: inherit;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' viewBox='0 0 10 10'%3E%3Cpath d='M2 3.5L5 6.5L8 3.5' stroke='%235C6470' stroke-width='1.5' stroke-linecap='round' fill='none'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 10px center;
  transition: border-color 0.15s;
}

.rp-filter-select:focus {
  border-color: rgba(92, 168, 255, 0.55);
}

/* 按钮 */
.rp-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 12.5px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
  border: 1px solid transparent;
  white-space: nowrap;
}

.rp-btn-sm {
  padding: 5px 10px;
  font-size: 11.5px;
}

.rp-btn-primary {
  background: var(--azure-dim, rgba(92, 168, 255, 0.15));
  border-color: rgba(92, 168, 255, 0.4);
  color: var(--azure-hi, #8CC4FF);
}

.rp-btn-primary:hover {
  background: rgba(92, 168, 255, 0.25);
  border-color: rgba(92, 168, 255, 0.6);
  box-shadow: 0 0 14px rgba(92, 168, 255, 0.18);
}

.rp-btn-ghost {
  background: transparent;
  border-color: var(--line-1, #2F3540);
  color: var(--ink-1, #97A0AF);
}

.rp-btn-ghost:hover {
  background: var(--bg-2, #171A20);
  color: var(--ink-0, #E8EBF0);
}

/* 行操作 */
.rp-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.rp-action-btn {
  padding: 3px 9px;
  border-radius: 6px;
  font-size: 11.5px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid var(--line-1, #2F3540);
  background: transparent;
  color: var(--ink-1, #97A0AF);
  font-family: inherit;
  transition: all 0.13s;
  white-space: nowrap;
}

.rp-action-btn:hover {
  background: var(--bg-2, #171A20);
  color: var(--ink-0, #E8EBF0);
  border-color: #3D4554;
}

.rp-action-danger {
  color: var(--bad, #F25C69);
  border-color: rgba(242, 92, 105, 0.3);
}

.rp-action-danger:hover {
  background: var(--bad-dim, rgba(242, 92, 105, 0.12));
  border-color: rgba(242, 92, 105, 0.5);
  color: var(--bad, #F25C69) !important;
}

/* 批量操作按钮 */
.rp-bulk-btn {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  font-size: 11.5px;
  font-weight: 600;
  border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540);
  background: var(--bg-2, #171A20);
  color: var(--ink-0, #E8EBF0);
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
}

.rp-bulk-btn:hover {
  background: var(--bg-3, #1F232B);
  border-color: #3D4554;
}

/* 错误提示 */
.rp-error {
  font-size: 12.5px;
  color: var(--bad, #F25C69);
  padding: 8px 12px;
  background: var(--bad-dim, rgba(242, 92, 105, 0.1));
  border: 1px solid rgba(242, 92, 105, 0.3);
  border-radius: 8px;
}

/* 空态 */
.rp-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 56px 24px;
  color: var(--ink-2, #5C6470);
  font-size: 13px;
}
</style>
