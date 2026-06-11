<template>
  <div class="q-vtabs" role="tablist" :aria-label="ariaLabel">
    <!-- 固定「全部」页签 -->
    <button
      class="q-vtab"
      :class="{ 'q-vtab-on': activeId === '__all__' }"
      role="tab"
      :aria-selected="activeId === '__all__'"
      @click="applyAll"
    >
      {{ t('datatable.savedViews.all') }}
      <span v-if="totalCount != null" class="q-vtab-n">{{ totalCount.toLocaleString() }}</span>
    </button>

    <!-- 用户保存的视图页签 -->
    <button
      v-for="view in savedViews"
      :key="view.id"
      class="q-vtab"
      :class="{ 'q-vtab-on': activeId === view.id }"
      role="tab"
      :aria-selected="activeId === view.id"
      @click="applyView(view)"
    >
      {{ view.name }}
      <span
        class="q-vtab-del"
        role="button"
        :aria-label="t('datatable.savedViews.delete', { name: view.name })"
        tabindex="0"
        @click.stop="deleteView(view.id)"
        @keydown.enter.stop="deleteView(view.id)"
      >✕</span>
    </button>

    <!-- 保存当前视图按钮 -->
    <button
      class="q-vtab q-vtab-add"
      :aria-label="t('datatable.savedViews.save')"
      @click="saveCurrentView"
    >
      + {{ t('datatable.savedViews.save') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SavedView, TableQueryState } from './types'

const { t } = useI18n()

// ── Props ──────────────────────────────────────────────────────────────
const props = withDefaults(defineProps<{
  /** localStorage key 前缀，用于持久化 */
  storageKey: string
  /** 当前查询状态（用于保存快照） */
  currentState?: Partial<TableQueryState>
  /** 「全部」页签显示的数量（可选） */
  totalCount?: number
  /** aria label */
  ariaLabel?: string
}>(), {
  currentState: undefined,
  totalCount: undefined,
  ariaLabel: '视图页签'
})

// ── Emits ──────────────────────────────────────────────────────────────
const emit = defineEmits<{
  'apply': [view: SavedView | null]
}>()

// ── 持久化 ────────────────────────────────────────────────────────────
const STORAGE_KEY = computed(() => `q_saved_views_${props.storageKey}`)

function loadViews(): SavedView[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY.value)
    return raw ? (JSON.parse(raw) as SavedView[]) : []
  } catch {
    return []
  }
}

function persistViews(views: SavedView[]) {
  try {
    localStorage.setItem(STORAGE_KEY.value, JSON.stringify(views))
  } catch {
    // localStorage 满或者 SSR，忽略
  }
}

const savedViews = ref<SavedView[]>(loadViews())

// 当 storageKey 变化时重新加载
watch(STORAGE_KEY, () => {
  savedViews.value = loadViews()
})

// ── 当前激活 id ──────────────────────────────────────────────────────
const activeId = ref<string>('__all__')

// ── 操作 ───────────────────────────────────────────────────────────────
function applyAll() {
  activeId.value = '__all__'
  emit('apply', null)
}

function applyView(view: SavedView) {
  activeId.value = view.id
  emit('apply', view)
}

function deleteView(id: string) {
  savedViews.value = savedViews.value.filter(v => v.id !== id)
  persistViews(savedViews.value)
  // 如果删除的是当前激活视图，回到全部
  if (activeId.value === id) {
    applyAll()
  }
}

function saveCurrentView() {
  const name = window.prompt(t('datatable.savedViews.namePrompt'))
  if (!name || !name.trim()) return
  const newView: SavedView = {
    id: `view_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`,
    name: name.trim(),
    state: props.currentState ? { ...props.currentState } : {}
  }
  savedViews.value = [...savedViews.value, newView]
  persistViews(savedViews.value)
  applyView(newView)
}


</script>

<style scoped>
/* ── 淬钢 QUENCH · SavedViewTabs 样式 ── */
.q-vtabs {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-bottom: 14px;
  align-items: center;
}

.q-vtab {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 6px 13px;
  border-radius: 9px;
  border: 1px solid var(--line-0, #20242C);
  background: var(--bg-1, #101216);
  color: var(--ink-1, #97A0AF);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  font-family: var(--font-ui, "Archivo", "PingFang SC", sans-serif);
  transition: border-color 0.15s, color 0.15s, background 0.15s, box-shadow 0.15s;
  line-height: 1.4;
}

.q-vtab:hover {
  border-color: var(--line-1, #2F3540);
  color: var(--ink-0, #E8EBF0);
}

.q-vtab-on {
  background: var(--azure-dim, rgba(92, 168, 255, 0.12));
  border-color: rgba(92, 168, 255, 0.4);
  color: var(--azure-hi, #8CC4FF) !important;
  box-shadow: 0 0 12px rgba(92, 168, 255, 0.12);
}

.q-vtab-add {
  border-style: dashed;
}

.q-vtab-n {
  font-family: var(--font-mono, "IBM Plex Mono", monospace);
  font-size: 10.5px;
  opacity: 0.75;
}

/* 删除小按钮 */
.q-vtab-del {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border-radius: 3px;
  font-size: 9px;
  opacity: 0;
  color: var(--ink-2, #5C6470);
  transition: opacity 0.15s, color 0.15s, background 0.15s;
  cursor: pointer;
  margin-left: -2px;
}

.q-vtab:hover .q-vtab-del {
  opacity: 1;
}

.q-vtab-del:hover {
  background: var(--bad-dim, rgba(242, 92, 105, 0.12));
  color: var(--bad, #F25C69);
}

/* 键盘焦点：页签按钮 */
.q-vtab:focus-visible {
  outline: none;
  box-shadow: 0 0 0 1.5px rgba(92, 168, 255, 0.65), 0 0 14px rgba(92, 168, 255, 0.2);
}

/* 删除小按钮焦点 */
.q-vtab-del:focus-visible {
  outline: none;
  box-shadow: 0 0 0 1.5px rgba(92, 168, 255, 0.65);
  border-radius: 3px;
  opacity: 1;
}

@media (prefers-reduced-motion: reduce) {
  .q-vtab { transition: none; }
  .q-vtab-del { transition: none; }
}
</style>
