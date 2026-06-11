<template>
  <div class="ud-tab-content">
    <!-- 汇总卡 -->
    <div class="ud-stats-row" v-if="!statsLoading && !statsError">
      <div class="ud-stat-card">
        <span class="ud-stat-label">总请求</span>
        <span class="ud-stat-val">{{ stats.total_requests.toLocaleString() }}</span>
      </div>
      <div class="ud-stat-card">
        <span class="ud-stat-label">总费用</span>
        <span class="ud-stat-val q-money">${{ stats.total_cost.toFixed(4) }}</span>
      </div>
      <div class="ud-stat-card">
        <span class="ud-stat-label">总 Token</span>
        <span class="ud-stat-val">{{ stats.total_tokens.toLocaleString() }}</span>
      </div>
    </div>

    <div v-if="loading" class="ud-loading">加载中…</div>
    <div v-else-if="error" class="ud-error">{{ error }}</div>
    <div v-else-if="!items.length" class="ud-empty">暂无用量记录</div>
    <div v-else class="ud-table-wrap">
      <table class="ud-table">
        <thead>
          <tr>
            <th>时间</th>
            <th>模型</th>
            <th>费用</th>
            <th>Token</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in items" :key="row.id">
            <td class="ud-muted ud-xs">{{ fmt(row.created_at) }}</td>
            <td class="ud-mono ud-xs">{{ row.model }}</td>
            <td class="q-money ud-xs">${{ row.total_cost?.toFixed(6) ?? '-' }}</td>
            <td class="ud-xs">{{ ((row.input_tokens || 0) + (row.output_tokens || 0)).toLocaleString() }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-if="total > items.length" class="ud-more">共 {{ total }} 条，仅展示前 {{ items.length }} 条</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { adminAPI } from '@/api/admin'
import type { AdminUser, AdminUsageLog } from '@/types'
import type { AdminUsageStatsResponse } from '@/api/admin/usage'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{ user: AdminUser; active: boolean }>()

const loading = ref(false)
const error = ref<string | null>(null)
const statsLoading = ref(false)
const statsError = ref<string | null>(null)
const items = ref<AdminUsageLog[]>([])
const total = ref(0)
const loaded = ref(false)
const stats = ref<AdminUsageStatsResponse>({
  total_requests: 0, total_input_tokens: 0, total_output_tokens: 0,
  total_cache_tokens: 0, total_tokens: 0, total_cost: 0,
  total_actual_cost: 0, total_account_cost: 0, average_duration_ms: 0
})

function fmt(iso: string | null | undefined) { return iso ? formatDateTime(iso) : '-' }

async function load() {
  if (loaded.value) return
  loading.value = true; error.value = null
  statsLoading.value = true; statsError.value = null
  try {
    const [listRes, statsRes] = await Promise.all([
      adminAPI.usage.list({ user_id: props.user.id, page: 1, page_size: 20 }),
      adminAPI.usage.getStats({ user_id: props.user.id })
    ])
    items.value = listRes.items; total.value = listRes.total
    stats.value = statsRes
    loaded.value = true
  } catch { error.value = '加载失败'; statsError.value = '统计失败' } finally {
    loading.value = false; statsLoading.value = false
  }
}

watch(() => props.active, (v) => { if (v) load() })
onMounted(() => { if (props.active) load() })
</script>

<style scoped>
.ud-tab-content { display: flex; flex-direction: column; gap: 14px; }
.ud-loading, .ud-empty { color: var(--ink-2); font-size: 12.5px; padding: 20px 0; text-align: center; }
.ud-error { color: var(--bad); font-size: 12.5px; }

.ud-stats-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }
.ud-stat-card {
  display: flex; flex-direction: column; gap: 3px;
  padding: 11px 14px;
  background: var(--bg-2); border: 1px solid var(--line-0); border-radius: 10px;
}
.ud-stat-label { font-size: 10.5px; color: var(--ink-2); }
.ud-stat-val { font-size: 14px; font-weight: 700; color: var(--ink-0); }

.ud-table-wrap { overflow-x: auto; }
.ud-table {
  width: 100%; border-collapse: collapse; font-size: 12px;
}
.ud-table th {
  text-align: left; padding: 7px 10px; font-size: 10.5px; font-weight: 600;
  color: var(--ink-2); border-bottom: 1px solid var(--line-0); white-space: nowrap;
}
.ud-table td {
  padding: 7px 10px; border-bottom: 1px solid var(--line-0); vertical-align: middle;
}
.ud-table tbody tr:last-child td { border-bottom: none; }
.ud-table tbody tr:hover td { background: var(--bg-2); }
.ud-mono { font-family: 'IBM Plex Mono', monospace; }
.ud-muted { color: var(--ink-2); }
.ud-xs { font-size: 11.5px; }
.ud-more { font-size: 11.5px; color: var(--ink-2); text-align: center; }
</style>
