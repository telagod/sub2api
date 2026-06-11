<template>
  <div class="ud-tab-content">
    <div v-if="loading" class="ud-loading">{{ t('admin.userTabs.loading') }}</div>
    <div v-else-if="error" class="ud-error">{{ error }}</div>
    <div v-else-if="!items.length" class="ud-empty">{{ t('admin.userTabs.noOrders') }}</div>
    <div v-else class="ud-list">
      <div v-for="order in items" :key="order.id" class="ud-order-card">
        <div class="ud-order-header">
          <span class="ud-mono ud-order-no">{{ order.out_trade_no || ('#' + order.id) }}</span>
          <span
            class="ud-badge"
            :class="statusClass(order.status)"
          >{{ order.status }}</span>
        </div>
        <div class="ud-order-meta">
          <span class="ud-meta-item q-money">${{ (order.amount / 100).toFixed(2) }}</span>
          <span class="ud-meta-item" v-if="order.payment_type">{{ order.payment_type }}</span>
          <span class="ud-meta-item">{{ fmt(order.created_at) }}</span>
        </div>
      </div>
    </div>
    <div v-if="total > items.length" class="ud-more">{{ t('admin.userTabs.totalCountPartial', { total, shown: items.length }) }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { AdminUser } from '@/types'
import type { PaymentOrder } from '@/types/payment'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const props = defineProps<{ user: AdminUser; active: boolean }>()

const loading = ref(false)
const error = ref<string | null>(null)
const items = ref<PaymentOrder[]>([])
const total = ref(0)
const loaded = ref(false)

function fmt(iso: string | null | undefined) { return iso ? formatDateTime(iso) : '-' }
function statusClass(s: string) {
  if (s === 'paid' || s === 'completed') return 'ud-badge-ok'
  if (s === 'pending') return 'ud-badge-warn'
  return 'ud-badge-bad'
}

async function load() {
  if (loaded.value) return
  loading.value = true; error.value = null
  try {
    const res = await adminAPI.payment.getOrders({ user_id: props.user.id, page: 1, page_size: 20 })
    // getOrders returns AxiosResponse — unwrap .data
    const payload = (res as any).data ?? res
    items.value = payload?.items ?? []; total.value = payload?.total ?? 0; loaded.value = true
  } catch { error.value = t('admin.userTabs.loadFailed') } finally { loading.value = false }
}

watch(() => props.active, (v) => { if (v) load() })
onMounted(() => { if (props.active) load() })
</script>

<style scoped>
.ud-tab-content { display: flex; flex-direction: column; gap: 10px; }
.ud-loading, .ud-empty { color: var(--ink-2); font-size: 12.5px; padding: 20px 0; text-align: center; }
.ud-error { color: var(--bad); font-size: 12.5px; }
.ud-list { display: flex; flex-direction: column; gap: 8px; }
.ud-order-card {
  padding: 12px 14px;
  background: var(--bg-2);
  border: 1px solid var(--line-0);
  border-radius: 10px;
  display: flex; flex-direction: column; gap: 6px;
}
.ud-order-header { display: flex; align-items: center; justify-content: space-between; }
.ud-order-no { font-size: 12px; color: var(--ink-1); }
.ud-order-meta { display: flex; gap: 12px; flex-wrap: wrap; align-items: center; }
.ud-meta-item { font-size: 12px; color: var(--ink-2); }
.ud-badge {
  font-size: 10.5px; font-weight: 600; padding: 2px 7px;
  border-radius: 5px; letter-spacing: 0.04em;
}
.ud-badge-ok { background: var(--ok-dim); color: var(--ok); border: 1px solid rgba(70,201,140,.3); }
.ud-badge-warn { background: var(--warn-dim); color: var(--warn); border: 1px solid rgba(224,179,78,.3); }
.ud-badge-bad { background: var(--bad-dim); color: var(--bad); border: 1px solid rgba(242,92,105,.3); }
.ud-mono { font-family: 'IBM Plex Mono', monospace; }
.ud-more { font-size: 11.5px; color: var(--ink-2); text-align: center; }
</style>
