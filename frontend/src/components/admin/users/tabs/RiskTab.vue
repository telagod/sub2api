<template>
  <div class="ud-tab-content">
    <div v-if="loading" class="ud-loading">加载中…</div>
    <div v-else-if="error" class="ud-error">{{ error }}</div>
    <div v-else-if="!items.length" class="ud-empty">暂无风控日志</div>
    <div v-else class="ud-list">
      <div v-for="log in items" :key="log.id" class="ud-log-card" :class="{ 'ud-log-flagged': log.flagged }">
        <div class="ud-log-header">
          <div class="ud-log-time ud-muted ud-xs">{{ fmt(log.created_at) }}</div>
          <div class="ud-log-badges">
            <span v-if="log.flagged" class="ud-badge ud-badge-bad">命中</span>
            <span v-if="log.auto_banned" class="ud-badge ud-badge-bad">自动封禁</span>
            <span v-if="!log.flagged" class="ud-badge ud-badge-ok">通过</span>
          </div>
        </div>
        <div class="ud-log-meta">
          <span class="ud-meta-item">模式：{{ log.mode }}</span>
          <span class="ud-meta-item" v-if="log.highest_category">类别：{{ log.highest_category }}</span>
          <span class="ud-meta-item" v-if="log.highest_score">分值：{{ (log.highest_score * 100).toFixed(1) }}%</span>
          <span class="ud-meta-item">模型：{{ log.model || '-' }}</span>
        </div>
        <div v-if="log.input_excerpt" class="ud-log-excerpt">{{ log.input_excerpt }}</div>
      </div>
    </div>
    <div v-if="total > items.length" class="ud-more">共 {{ total }} 条，仅展示前 {{ items.length }} 条</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { adminAPI } from '@/api/admin'
import type { AdminUser } from '@/types'
import type { ContentModerationLog } from '@/api/admin/riskControl'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{ user: AdminUser; active: boolean }>()

const loading = ref(false)
const error = ref<string | null>(null)
const items = ref<ContentModerationLog[]>([])
const total = ref(0)
const loaded = ref(false)

function fmt(iso: string | null | undefined) { return iso ? formatDateTime(iso) : '-' }

async function load() {
  if (loaded.value) return
  loading.value = true; error.value = null
  try {
    const res = await adminAPI.riskControl.listLogs({ search: String(props.user.id), page: 1, page_size: 20 })
    // filter to this user since the API searches by email/id text
    items.value = res.items.filter(l => l.user_id === props.user.id)
    total.value = items.value.length
    loaded.value = true
  } catch { error.value = '加载失败' } finally { loading.value = false }
}

watch(() => props.active, (v) => { if (v) load() })
onMounted(() => { if (props.active) load() })
</script>

<style scoped>
.ud-tab-content { display: flex; flex-direction: column; gap: 10px; }
.ud-loading, .ud-empty { color: var(--ink-2); font-size: 12.5px; padding: 20px 0; text-align: center; }
.ud-error { color: var(--bad); font-size: 12.5px; }
.ud-list { display: flex; flex-direction: column; gap: 8px; }
.ud-log-card {
  padding: 11px 14px;
  background: var(--bg-2);
  border: 1px solid var(--line-0);
  border-radius: 10px;
  display: flex; flex-direction: column; gap: 6px;
}
.ud-log-flagged { border-color: rgba(242,92,105,.35); background: rgba(242,92,105,.04); }
.ud-log-header { display: flex; align-items: center; justify-content: space-between; }
.ud-log-badges { display: flex; gap: 6px; }
.ud-log-meta { display: flex; gap: 12px; flex-wrap: wrap; }
.ud-log-excerpt {
  font-size: 11px; color: var(--ink-2); font-family: 'IBM Plex Mono', monospace;
  white-space: pre-wrap; word-break: break-all; max-height: 48px; overflow: hidden;
  line-height: 1.5;
}
.ud-badge {
  font-size: 10.5px; font-weight: 600; padding: 2px 7px;
  border-radius: 5px; letter-spacing: 0.04em;
}
.ud-badge-ok { background: var(--ok-dim); color: var(--ok); border: 1px solid rgba(70,201,140,.3); }
.ud-badge-bad { background: var(--bad-dim); color: var(--bad); border: 1px solid rgba(242,92,105,.3); }
.ud-meta-item { font-size: 11.5px; color: var(--ink-2); }
.ud-muted { color: var(--ink-2); }
.ud-xs { font-size: 11.5px; }
.ud-more { font-size: 11.5px; color: var(--ink-2); text-align: center; }
</style>
