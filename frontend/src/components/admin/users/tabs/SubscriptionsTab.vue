<template>
  <div class="ud-tab-content">
    <div v-if="loading" class="ud-loading">加载中…</div>
    <div v-else-if="error" class="ud-error">{{ error }}</div>
    <div v-else-if="!items.length" class="ud-empty">暂无订阅记录</div>
    <div v-else class="ud-list">
      <div v-for="sub in items" :key="sub.id" class="ud-sub-card">
        <div class="ud-sub-header">
          <div class="ud-sub-title">
            <span class="ud-mono">#{{ sub.id }}</span>
            <span v-if="sub.group?.name" class="ud-sub-group">{{ sub.group.name }}</span>
          </div>
          <span
            class="ud-badge"
            :class="{
              'ud-badge-ok': sub.status === 'active',
              'ud-badge-warn': sub.status === 'expired',
              'ud-badge-bad': sub.status === 'revoked'
            }"
          >{{ statusLabel(sub.status) }}</span>
        </div>
        <div class="ud-sub-meta">
          <span class="ud-meta-item">开始：{{ fmt(sub.starts_at) }}</span>
          <span class="ud-meta-item" v-if="sub.expires_at">到期：{{ fmt(sub.expires_at) }}</span>
          <span class="ud-meta-item" v-else>永久有效</span>
        </div>
        <div class="ud-sub-usage">
          <span class="ud-meta-item">日消耗 ${{ fmtCost(sub.daily_usage_usd) }}</span>
          <span class="ud-meta-item">月消耗 ${{ fmtCost(sub.monthly_usage_usd) }}</span>
        </div>
      </div>
    </div>
    <div v-if="total > items.length" class="ud-more">共 {{ total }} 条，仅展示前 {{ items.length }} 条</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { adminAPI } from '@/api/admin'
import type { AdminUser, UserSubscription } from '@/types'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{ user: AdminUser; active: boolean }>()

const loading = ref(false)
const error = ref<string | null>(null)
const items = ref<UserSubscription[]>([])
const total = ref(0)
const loaded = ref(false)

function fmt(iso: string | null | undefined) { return iso ? formatDateTime(iso) : '-' }
function fmtCost(v: number) { return v.toFixed(4) }
function statusLabel(s: string) {
  return s === 'active' ? '活跃' : s === 'expired' ? '已过期' : '已撤销'
}

async function load() {
  if (loaded.value) return
  loading.value = true; error.value = null
  try {
    const res = await adminAPI.subscriptions.listByUser(props.user.id, 1, 20)
    items.value = res.items; total.value = res.total; loaded.value = true
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
.ud-sub-card {
  padding: 12px 14px;
  background: var(--bg-2);
  border: 1px solid var(--line-0);
  border-radius: 10px;
  display: flex; flex-direction: column; gap: 6px;
}
.ud-sub-header { display: flex; align-items: center; justify-content: space-between; }
.ud-sub-title { display: flex; align-items: center; gap: 8px; font-size: 12.5px; }
.ud-sub-group { color: var(--ink-1); }
.ud-badge {
  font-size: 10.5px; font-weight: 600; padding: 2px 7px;
  border-radius: 5px; letter-spacing: 0.04em;
}
.ud-badge-ok { background: var(--ok-dim); color: var(--ok); border: 1px solid rgba(70,201,140,.3); }
.ud-badge-warn { background: var(--warn-dim); color: var(--warn); border: 1px solid rgba(224,179,78,.3); }
.ud-badge-bad { background: var(--bad-dim); color: var(--bad); border: 1px solid rgba(242,92,105,.3); }
.ud-sub-meta, .ud-sub-usage { display: flex; gap: 16px; flex-wrap: wrap; }
.ud-meta-item { font-size: 11.5px; color: var(--ink-2); }
.ud-mono { font-family: 'IBM Plex Mono', monospace; font-size: 11.5px; color: var(--ink-1); }
.ud-more { font-size: 11.5px; color: var(--ink-2); text-align: center; }
</style>
