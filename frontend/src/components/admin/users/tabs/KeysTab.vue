<template>
  <div class="ud-tab-content">
    <div v-if="loading" class="ud-loading">加载中…</div>
    <div v-else-if="error" class="ud-error">{{ error }}</div>
    <div v-else-if="!items.length" class="ud-empty">暂无 API Key</div>
    <div v-else class="ud-list">
      <div v-for="key in items" :key="key.id" class="ud-key-card">
        <div class="ud-key-header">
          <div class="ud-key-name">{{ key.name }}</div>
          <span
            class="ud-badge"
            :class="{
              'ud-badge-ok': key.status === 'active',
              'ud-badge-bad': key.status !== 'active'
            }"
          >{{ key.status === 'active' ? '活跃' : key.status }}</span>
        </div>
        <div class="ud-key-value ud-mono">
          {{ key.key.substring(0, 20) }}…{{ key.key.substring(key.key.length - 6) }}
        </div>
        <div class="ud-key-meta">
          <span class="ud-meta-item" v-if="key.group?.name">组：{{ key.group.name }}</span>
          <span class="ud-meta-item">配额：{{ key.quota === 0 ? '不限' : ('$' + key.quota.toFixed(2)) }}</span>
          <span class="ud-meta-item">已用：${{ key.quota_used.toFixed(4) }}</span>
          <span class="ud-meta-item">创建：{{ fmt(key.created_at) }}</span>
          <span class="ud-meta-item" v-if="key.last_used_at">末用：{{ fmt(key.last_used_at) }}</span>
        </div>
      </div>
    </div>
    <div v-if="total > items.length" class="ud-more">共 {{ total }} 条</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { adminAPI } from '@/api/admin'
import type { AdminUser, ApiKey } from '@/types'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{ user: AdminUser; active: boolean }>()

const loading = ref(false)
const error = ref<string | null>(null)
const items = ref<ApiKey[]>([])
const total = ref(0)
const loaded = ref(false)

function fmt(iso: string | null | undefined) { return iso ? formatDateTime(iso) : '-' }

async function load() {
  if (loaded.value) return
  loading.value = true; error.value = null
  try {
    const res = await adminAPI.users.getUserApiKeys(props.user.id)
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
.ud-key-card {
  padding: 12px 14px;
  background: var(--bg-2);
  border: 1px solid var(--line-0);
  border-radius: 10px;
  display: flex; flex-direction: column; gap: 6px;
}
.ud-key-header { display: flex; align-items: center; justify-content: space-between; }
.ud-key-name { font-size: 13px; font-weight: 600; color: var(--ink-0); }
.ud-key-value { font-size: 11px; color: var(--ink-2); word-break: break-all; }
.ud-key-meta { display: flex; gap: 12px; flex-wrap: wrap; }
.ud-meta-item { font-size: 11.5px; color: var(--ink-2); }
.ud-badge {
  font-size: 10.5px; font-weight: 600; padding: 2px 7px;
  border-radius: 5px; letter-spacing: 0.04em;
}
.ud-badge-ok { background: var(--ok-dim); color: var(--ok); border: 1px solid rgba(70,201,140,.3); }
.ud-badge-bad { background: var(--bad-dim); color: var(--bad); border: 1px solid rgba(242,92,105,.3); }
.ud-mono { font-family: 'IBM Plex Mono', monospace; }
.ud-more { font-size: 11.5px; color: var(--ink-2); text-align: center; }
</style>
