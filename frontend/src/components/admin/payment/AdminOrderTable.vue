<template>
  <div>
    <!-- 筛选栏 -->
    <div class="oq-filter" style="margin-bottom:14px">
      <div class="oq-search" :class="{ 'oq-search-focus': searchFocused }">
        <svg width="13" height="13" viewBox="0 0 13 13" fill="none" aria-hidden="true"><circle cx="5.5" cy="5.5" r="4" stroke="currentColor" stroke-width="1.2"/><path d="M9 9L11.5 11.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
        <input
          v-model="searchQuery"
          class="oq-search-input"
          :placeholder="t('payment.admin.searchOrders')"
          @focus="searchFocused = true"
          @blur="searchFocused = false"
          @input="handleSearch"
        />
      </div>

      <!-- 状态筛选 -->
      <div class="oq-chip-wrap" v-click-outside="() => showStatusMenu = false">
        <button class="oq-chip" :class="{ 'oq-chip-on': filters.status }" @click="showStatusMenu = !showStatusMenu">
          状态 <b>{{ statusLabel }}</b>
          <svg width="10" height="10" viewBox="0 0 10 10" fill="none"><path d="M2 3.5L5 6.5L8 3.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
        </button>
        <div v-if="showStatusMenu" class="oq-menu">
          <button v-for="opt in statusFilterOptions" :key="opt.value" class="oq-menu-item" :class="{ on: filters.status === opt.value }" @click="filters.status = opt.value; showStatusMenu = false; emitFiltersChanged()">{{ opt.label }}</button>
        </div>
      </div>

      <!-- 支付方式筛选 -->
      <div class="oq-chip-wrap" v-click-outside="() => showPayTypeMenu = false">
        <button class="oq-chip" :class="{ 'oq-chip-on': filters.payment_type }" @click="showPayTypeMenu = !showPayTypeMenu">
          支付 <b>{{ payTypeLabel }}</b>
          <svg width="10" height="10" viewBox="0 0 10 10" fill="none"><path d="M2 3.5L5 6.5L8 3.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
        </button>
        <div v-if="showPayTypeMenu" class="oq-menu">
          <button v-for="opt in paymentTypeFilterOptions" :key="opt.value" class="oq-menu-item" :class="{ on: filters.payment_type === opt.value }" @click="filters.payment_type = opt.value; showPayTypeMenu = false; emitFiltersChanged()">{{ opt.label }}</button>
        </div>
      </div>

      <!-- 订单类型筛选 -->
      <div class="oq-chip-wrap" v-click-outside="() => showOrderTypeMenu = false">
        <button class="oq-chip" :class="{ 'oq-chip-on': filters.order_type }" @click="showOrderTypeMenu = !showOrderTypeMenu">
          类型 <b>{{ orderTypeLabel }}</b>
          <svg width="10" height="10" viewBox="0 0 10 10" fill="none"><path d="M2 3.5L5 6.5L8 3.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
        </button>
        <div v-if="showOrderTypeMenu" class="oq-menu">
          <button v-for="opt in orderTypeFilterOptions" :key="opt.value" class="oq-menu-item" :class="{ on: filters.order_type === opt.value }" @click="filters.order_type = opt.value; showOrderTypeMenu = false; emitFiltersChanged()">{{ opt.label }}</button>
        </div>
      </div>

      <div style="margin-left:auto">
        <button class="oq-btn" :disabled="loading" @click="emit('refresh')">
          <svg width="13" height="13" viewBox="0 0 13 13" fill="none" :class="loading ? 'oq-spin-icon' : ''"><path d="M11 6.5A4.5 4.5 0 1 1 6.5 2a4.5 4.5 0 0 1 3.18 1.32" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/><path d="M11 2v2.5H8.5" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/></svg>
          {{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <!-- 表格 -->
    <div class="oq-card">
      <DataTable :columns="columns" :data="orders" :loading="loading">
        <template #cell-id="{ value }">
          <span class="oq-mono oq-xs oq-muted">#{{ value }}</span>
        </template>

        <template #cell-user_id="{ value }">
          <span class="oq-mono oq-xs oq-muted">#{{ value }}</span>
        </template>

        <template #cell-pay_amount="{ value, row }">
          <span class="oq-amount" style="font-size:13px">¥{{ value.toFixed(2) }}</span>
          <span v-if="row.fee_rate > 0" class="oq-amount-sub" :title="t('payment.orders.fee') + ': ' + row.fee_rate + '%'">({{ row.fee_rate }}%)</span>
          <div v-if="row.amount !== row.pay_amount" class="oq-amount-sub">
            {{ t('payment.orders.creditedAmount') }}: {{ row.order_type === 'balance' ? '$' : '¥' }}{{ row.amount.toFixed(2) }}
          </div>
        </template>

        <template #cell-payment_type="{ value }">
          <span class="oq-xs" style="color:var(--ink-0)">{{ t('payment.methods.' + value, value) }}</span>
        </template>

        <template #cell-status="{ value }">
          <span :class="['oq-badge', statusBadgeQuench(value)]">{{ t('payment.status.' + value.toLowerCase(), value) }}</span>
        </template>

        <template #cell-order_type="{ value }">
          <span class="oq-xs" style="color:var(--ink-1)">{{ t('payment.admin.' + value + 'Order', value) }}</span>
        </template>

        <template #cell-created_at="{ value }">
          <span class="oq-mono oq-xs oq-muted">{{ formatDateTime(value) }}</span>
        </template>

        <template #cell-actions="{ row }">
          <div class="oq-acts">
            <button class="oq-ib" @click="emit('detail', row)">
              <svg width="12" height="12" viewBox="0 0 12 12" fill="none"><circle cx="6" cy="6" r="4.5" stroke="currentColor" stroke-width="1.2"/><circle cx="6" cy="6" r="1.5" fill="currentColor"/></svg>
              <span>{{ t('common.view') }}</span>
            </button>
            <button v-if="row.status === 'PENDING'" class="oq-ib oq-ib-warn" @click="emit('cancel', row)">
              <svg width="12" height="12" viewBox="0 0 12 12" fill="none"><path d="M3.5 3.5L8.5 8.5M8.5 3.5L3.5 8.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
              <span>{{ t('payment.orders.cancel') }}</span>
            </button>
            <button v-if="row.status === 'FAILED'" class="oq-ib" @click="emit('retry', row)">
              <svg width="12" height="12" viewBox="0 0 12 12" fill="none"><path d="M10 5.5A3.5 3.5 0 1 1 5.5 2a3.5 3.5 0 0 1 2.47 1.03" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/><path d="M10 2v2H8" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
              <span>{{ t('payment.admin.retry') }}</span>
            </button>
            <!-- TODO(backlog): 此组件为 backup/unused，晋升为主组件时需拆分为三分支：
                 REFUND_REQUESTED → emit('approveRefund'), REFUND_FAILED → emit('retryRefund'), 其余 → emit('refund')
                 参考 AdminOrdersView.vue 中的 template v-if/v-else-if 链 -->
            <button v-if="canRefundRow(row)" class="oq-ib oq-ib-bad" @click="emit('refund', row)">
              <svg width="12" height="12" viewBox="0 0 12 12" fill="none"><path d="M6 2v8M3 5l3-3 3 3" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></svg>
              <span>{{ t('payment.admin.refund') }}</span>
            </button>
          </div>
        </template>
      </DataTable>

      <Pagination
        v-if="total > 0"
        :page="page"
        :total="total"
        :page-size="pageSize"
        @update:page="emit('update:page', $event)"
        @update:pageSize="emit('update:pageSize', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PaymentOrder } from '@/types/payment'
import type { Column } from '@/components/common/types'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import { canRefund, formatOrderDateTime } from '@/components/payment/orderUtils'

const { t } = useI18n()

defineProps<{
  orders: PaymentOrder[]
  loading: boolean
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  (e: 'detail', order: PaymentOrder): void
  (e: 'cancel', order: PaymentOrder): void
  (e: 'retry', order: PaymentOrder): void
  (e: 'refund', order: PaymentOrder): void
  (e: 'refresh'): void
  (e: 'update:page', page: number): void
  (e: 'update:pageSize', size: number): void
  (e: 'filter', filters: { keyword?: string; status?: string; payment_type?: string; order_type?: string }): void
}>()

// UI state
const searchFocused = ref(false)
const showStatusMenu = ref(false)
const showPayTypeMenu = ref(false)
const showOrderTypeMenu = ref(false)
const searchQuery = ref('')
const filters = reactive({ status: '', payment_type: '', order_type: '' })

// vClick-outside directive (inline)
const vClickOutside = {
  mounted(el: HTMLElement, binding: { value: () => void }) {
    (el as any)._clickOutside = (e: MouseEvent) => { if (!el.contains(e.target as Node)) binding.value() }
    document.addEventListener('click', (el as any)._clickOutside)
  },
  unmounted(el: HTMLElement) { document.removeEventListener('click', (el as any)._clickOutside) }
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null
function handleSearch() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => emitFiltersChanged(), 300)
}

function emitFiltersChanged() {
  emit('filter', {
    keyword: searchQuery.value || undefined,
    status: filters.status || undefined,
    payment_type: filters.payment_type || undefined,
    order_type: filters.order_type || undefined,
  })
}

const columns = computed<Column[]>(() => [
  { key: 'id', label: t('payment.orders.orderId') },
  { key: 'user_id', label: t('payment.orders.userId') },
  { key: 'pay_amount', label: t('payment.orders.payAmount') },
  { key: 'payment_type', label: t('payment.orders.paymentMethod') },
  { key: 'status', label: t('payment.orders.status') },
  { key: 'order_type', label: t('payment.orders.orderType') },
  { key: 'created_at', label: t('payment.orders.createdAt') },
  { key: 'actions', label: t('payment.orders.actions') },
])

const statusFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allStatuses') },
  { value: 'PENDING', label: t('payment.status.pending') },
  { value: 'PAID', label: t('payment.status.paid') },
  { value: 'COMPLETED', label: t('payment.status.completed') },
  { value: 'EXPIRED', label: t('payment.status.expired') },
  { value: 'CANCELLED', label: t('payment.status.cancelled') },
  { value: 'FAILED', label: t('payment.status.failed') },
  { value: 'REFUNDED', label: t('payment.status.refunded') },
  { value: 'REFUND_REQUESTED', label: t('payment.status.refund_requested') },
  { value: 'REFUND_FAILED', label: t('payment.status.refund_failed') },
])

const paymentTypeFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allPaymentTypes') },
  { value: 'alipay', label: t('payment.methods.alipay') },
  { value: 'wxpay', label: t('payment.methods.wxpay') },
  { value: 'stripe', label: t('payment.methods.stripe') },
  { value: 'airwallex', label: t('payment.methods.airwallex') },
])

const orderTypeFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allOrderTypes') },
  { value: 'balance', label: t('payment.admin.balanceOrder') },
  { value: 'subscription', label: t('payment.admin.subscriptionOrder') },
])

const statusLabel = computed(() => statusFilterOptions.value.find(o => o.value === filters.status)?.label ?? '全部')
const payTypeLabel = computed(() => paymentTypeFilterOptions.value.find(o => o.value === filters.payment_type)?.label ?? '全部')
const orderTypeLabel = computed(() => orderTypeFilterOptions.value.find(o => o.value === filters.order_type)?.label ?? '全部')

function statusBadgeQuench(status: string): string {
  const s = status.toUpperCase()
  if (s === 'COMPLETED' || s === 'PAID') return 'oq-badge-ok'
  if (s === 'PENDING' || s === 'REFUND_REQUESTED') return 'oq-badge-warn'
  if (s === 'FAILED' || s === 'REFUND_FAILED' || s === 'CANCELLED' || s === 'EXPIRED') return 'oq-badge-bad'
  if (s === 'REFUNDED') return 'oq-badge-azure'
  return 'oq-badge-dim'
}

function canRefundRow(order: PaymentOrder): boolean {
  return canRefund(order.status)
}

function formatDateTime(dateStr: string): string {
  return formatOrderDateTime(dateStr)
}
</script>

<style scoped>
.oq-spin-icon { animation: oq-icon-spin .7s linear infinite; }
@keyframes oq-icon-spin { to { transform: rotate(360deg); } }
</style>
