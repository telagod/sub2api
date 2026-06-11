<template>
  <AppLayout>
    <div class="oq-root">
      <!-- 页头 -->
      <div class="oq-head">
        <div>
          <h1 class="oq-title">订单流水</h1>
          <p class="oq-desc">收入域 · 全量订单 · 点击操作列处理退款</p>
        </div>
        <div class="oq-head-acts">
          <button class="oq-btn" :disabled="ordersLoading" @click="loadOrders">
            <svg width="13" height="13" viewBox="0 0 13 13" fill="none" :class="ordersLoading ? 'spin-icon' : ''"><path d="M11 6.5A4.5 4.5 0 1 1 6.5 2a4.5 4.5 0 0 1 3.18 1.32" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/><path d="M11 2v2.5H8.5" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/></svg>
            刷新
          </button>
        </div>
      </div>

      <!-- 筛选栏 -->
      <OrdersFilterBar
        v-model:search="orderSearch"
        v-model:status="orderFilters.status"
        v-model:pay-type="orderFilters.payment_type"
        v-model:order-type="orderFilters.order_type"
        @update:search="debounceLoadOrders"
        @update:status="loadOrders"
        @update:pay-type="loadOrders"
        @update:order-type="loadOrders"
        @clear="clearFilters"
      />

      <!-- 表格卡片 -->
      <div class="oq-card">
        <OrderTable :orders="orders" :loading="ordersLoading" show-user>
          <template #actions="{ row }">
            <div class="oq-acts">
              <button class="oq-ib" @click="showOrderDetail(row)">
                <svg width="12" height="12" viewBox="0 0 12 12" fill="none"><circle cx="6" cy="6" r="4.5" stroke="currentColor" stroke-width="1.2"/><circle cx="6" cy="6" r="1.5" fill="currentColor"/></svg>
                {{ t('common.view') }}
              </button>
              <button v-if="row.status === 'PENDING'" class="oq-ib oq-ib-warn" @click="handleCancelOrder(row)">
                <svg width="12" height="12" viewBox="0 0 12 12" fill="none"><path d="M3.5 3.5L8.5 8.5M8.5 3.5L3.5 8.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
                {{ t('payment.orders.cancel') }}
              </button>
              <button v-if="row.status === 'FAILED'" class="oq-ib" @click="handleRetryOrder(row)">
                <svg width="12" height="12" viewBox="0 0 12 12" fill="none"><path d="M10 5.5A3.5 3.5 0 1 1 5.5 2a3.5 3.5 0 0 1 2.47 1.03" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/><path d="M10 2v2H8" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
                {{ t('payment.admin.retry') }}
              </button>
              <template v-if="row.status === 'REFUND_REQUESTED'">
                <span v-if="row.refund_amount" class="oq-badge oq-badge-warn oq-mono" style="font-size:10.5px">{{ row.order_type === 'balance' ? '$' : '¥' }}{{ row.refund_amount.toFixed(2) }}</span>
                <button class="oq-ib oq-ib-ok" @click="openRefundDialog(row)">
                  <svg width="12" height="12" viewBox="0 0 12 12" fill="none"><path d="M2 6.5L4.5 9L10 3.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                  {{ t('payment.admin.approveRefund') }}
                </button>
              </template>
              <button v-else-if="row.status === 'REFUND_FAILED'" class="oq-ib oq-ib-bad" @click="openRefundDialog(row)">
                <svg width="12" height="12" viewBox="0 0 12 12" fill="none"><path d="M10 5.5A3.5 3.5 0 1 1 5.5 2a3.5 3.5 0 0 1 2.47 1.03" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/><path d="M10 2v2H8" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
                {{ t('payment.admin.retryRefund') }}
              </button>
              <button v-else-if="row.status === 'COMPLETED' || row.status === 'PARTIALLY_REFUNDED'" class="oq-ib oq-ib-bad" @click="openRefundDialog(row)">
                <svg width="12" height="12" viewBox="0 0 12 12" fill="none"><path d="M6 2v8M3 5l3-3 3 3" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                {{ t('payment.admin.refund') }}
              </button>
            </div>
          </template>
        </OrderTable>
        <Pagination v-if="orderPagination.total > 0" :page="orderPagination.page" :total="orderPagination.total" :page-size="orderPagination.page_size" @update:page="handleOrderPageChange" @update:pageSize="handleOrderPageSizeChange" />
      </div>
    </div>

    <!-- 订单详情弹窗 -->
    <Teleport to="body">
      <div v-if="showDetailDialog" class="oq-overlay" @click.self="showDetailDialog = false">
        <div class="oq-dialog">
          <div class="oq-dlg-title">
            订单详情
            <button class="oq-dlg-close" @click="showDetailDialog = false"><svg width="13" height="13" viewBox="0 0 13 13" fill="none"><path d="M3 3L10 10M10 3L3 10" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></svg></button>
          </div>
          <div v-if="selectedOrder" class="oq-dlg-grid">
            <div class="oq-dlg-field"><label>{{ t('payment.orders.orderId') }}</label><p class="oq-mono">#{{ selectedOrder.id }}</p></div>
            <div class="oq-dlg-field"><label>{{ t('payment.orders.orderNo') }}</label><p class="oq-mono oq-xs">{{ selectedOrder.out_trade_no }}</p></div>
            <div class="oq-dlg-field"><label>{{ t('payment.orders.status') }}</label><span :class="['oq-badge', statusBadgeQuench(selectedOrder.status)]">{{ t('payment.status.' + selectedOrder.status.toLowerCase(), selectedOrder.status) }}</span></div>
            <div class="oq-dlg-field"><label>{{ t('payment.orders.amount') }}</label><span class="oq-amount">{{ selectedOrder.order_type === 'balance' ? '$' : '¥' }}{{ selectedOrder.amount.toFixed(2) }}</span></div>
            <div class="oq-dlg-field"><label>{{ t('payment.orders.payAmount') }}</label><span class="oq-amount">¥{{ selectedOrder.pay_amount.toFixed(2) }}</span></div>
            <div class="oq-dlg-field"><label>{{ t('payment.orders.paymentMethod') }}</label><p>{{ t('payment.methods.' + selectedOrder.payment_type, selectedOrder.payment_type) }}</p></div>
            <div class="oq-dlg-field"><label>{{ t('payment.admin.feeRate') }}</label><p>{{ selectedOrder.fee_rate }}%</p></div>
            <div class="oq-dlg-field"><label>{{ t('payment.orders.createdAt') }}</label><p class="oq-xs oq-muted">{{ formatDateTime(selectedOrder.created_at) }}</p></div>
            <div class="oq-dlg-field"><label>{{ t('payment.admin.expiresAt') }}</label><p class="oq-xs oq-muted">{{ formatDateTime(selectedOrder.expires_at) }}</p></div>
            <div v-if="selectedOrder.paid_at" class="oq-dlg-field"><label>{{ t('payment.admin.paidAt') }}</label><p class="oq-xs oq-muted">{{ formatDateTime(selectedOrder.paid_at) }}</p></div>
            <div v-if="selectedOrder.refund_amount" class="oq-refund-block">
              <h4>{{ t('payment.admin.refundInfo') }}</h4>
              <p>{{ t('payment.admin.refundAmount') }}: {{ selectedOrder.order_type === 'balance' ? '$' : '¥' }}{{ selectedOrder.refund_amount.toFixed(2) }}</p>
              <p v-if="selectedOrder.refund_reason">{{ t('payment.admin.refundReason') }}: {{ selectedOrder.refund_reason }}</p>
            </div>
            <div v-if="selectedOrder.refund_requested_at" class="oq-dlg-section">
              <div class="oq-dlg-sep"></div>
              <p class="oq-dlg-section-title">{{ t('payment.admin.refundRequestInfo') }}</p>
              <div class="oq-dlg-field"><label>{{ t('payment.admin.refundRequestedAt') }}</label><p class="oq-xs oq-muted">{{ formatDateTime(selectedOrder.refund_requested_at) }}</p></div>
              <div class="oq-dlg-field"><label>{{ t('payment.admin.refundRequestedBy') }}</label><p>#{{ selectedOrder.refund_requested_by }}</p></div>
              <div class="oq-dlg-field"><label>{{ t('payment.admin.refundRequestReason') }}</label><p>{{ selectedOrder.refund_request_reason }}</p></div>
            </div>
            <div v-if="orderAuditLogs.length > 0" class="oq-dlg-section">
              <div class="oq-dlg-sep"></div>
              <p class="oq-dlg-section-title">{{ t('payment.admin.auditLogs') }}</p>
              <div class="oq-audit-wrap">
                <div v-for="log in orderAuditLogs" :key="log.id" class="oq-audit-item">
                  <div class="oq-audit-head"><span class="oq-audit-action">{{ log.action }}</span><span class="oq-audit-time">{{ formatDateTime(log.created_at) }}</span></div>
                  <div v-if="log.detail" class="oq-audit-detail">{{ log.detail }}</div>
                  <div v-if="log.operator" class="oq-audit-detail">{{ t('payment.admin.operator') }}: {{ log.operator }}</div>
                </div>
              </div>
            </div>
          </div>
          <div class="oq-dlg-foot"><button class="oq-btn" @click="showDetailDialog = false">关闭</button></div>
        </div>
      </div>
    </Teleport>

    <AdminRefundDialog :show="showRefundDialog" :order="selectedOrder" :submitting="refundSubmitting" @confirm="handleRefund" @cancel="showRefundDialog = false" />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatOrderDateTime } from '@/components/payment/orderUtils'
import type { PaymentOrder } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import AdminRefundDialog from '@/components/admin/payment/AdminRefundDialog.vue'
import OrderTable from '@/components/payment/OrderTable.vue'
import OrdersFilterBar from './OrdersFilterBar.vue'

interface AuditLog { id: number; action: string; detail: string | null; operator: string | null; created_at: string }

const { t } = useI18n()
const appStore = useAppStore()

const ordersLoading = ref(false)
const orders = ref<PaymentOrder[]>([])
const orderSearch = ref('')
const orderFilters = reactive({ status: '', payment_type: '', order_type: '' })
const orderPagination = reactive({ page: 1, page_size: 20, total: 0 })
const selectedOrder = ref<PaymentOrder | null>(null)
const showDetailDialog = ref(false)
const showRefundDialog = ref(false)
const refundSubmitting = ref(false)
const orderAuditLogs = ref<AuditLog[]>([])

let debounceTimer: ReturnType<typeof setTimeout> | null = null
function debounceLoadOrders() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => loadOrders(), 300)
}

async function loadOrders() {
  ordersLoading.value = true
  try {
    const res = await adminPaymentAPI.getOrders({
      page: orderPagination.page, page_size: orderPagination.page_size,
      keyword: orderSearch.value || undefined, status: orderFilters.status || undefined,
      payment_type: orderFilters.payment_type || undefined, order_type: orderFilters.order_type || undefined,
    })
    orders.value = res.data.items || []
    orderPagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally { ordersLoading.value = false }
}

function handleOrderPageChange(page: number) { orderPagination.page = page; loadOrders() }
function handleOrderPageSizeChange(size: number) { orderPagination.page_size = size; orderPagination.page = 1; loadOrders() }
function clearFilters() { orderSearch.value = ''; orderFilters.status = ''; orderFilters.payment_type = ''; orderFilters.order_type = ''; orderPagination.page = 1; loadOrders() }

function statusBadgeQuench(status: string): string {
  const s = status.toUpperCase()
  if (s === 'COMPLETED' || s === 'PAID') return 'oq-badge-ok'
  if (s === 'PENDING' || s === 'REFUND_REQUESTED') return 'oq-badge-warn'
  if (s === 'FAILED' || s === 'REFUND_FAILED' || s === 'CANCELLED' || s === 'EXPIRED') return 'oq-badge-bad'
  if (s === 'REFUNDED') return 'oq-badge-azure'
  return 'oq-badge-dim'
}

async function showOrderDetail(order: PaymentOrder) {
  selectedOrder.value = order; orderAuditLogs.value = []; showDetailDialog.value = true
  try {
    const res = await adminPaymentAPI.getOrder(order.id)
    const data = res.data as unknown as Record<string, unknown>
    if (data.order) selectedOrder.value = data.order as PaymentOrder
    orderAuditLogs.value = ((data.auditLogs || data.audit_logs || []) as unknown) as AuditLog[]
  } catch { /* keep cached */ }
}

async function handleCancelOrder(order: PaymentOrder) {
  try { await adminPaymentAPI.cancelOrder(order.id); appStore.showSuccess(t('payment.admin.orderCancelled')); loadOrders() }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
}

async function handleRetryOrder(order: PaymentOrder) {
  try { await adminPaymentAPI.retryRecharge(order.id); appStore.showSuccess(t('payment.admin.retrySuccess')); loadOrders() }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
}

function openRefundDialog(order: PaymentOrder) { selectedOrder.value = order; showRefundDialog.value = true }

async function handleRefund(data: { amount: number; reason: string; deduct_balance: boolean; force: boolean }) {
  if (!selectedOrder.value) return
  refundSubmitting.value = true
  try {
    await adminPaymentAPI.refundOrder(selectedOrder.value.id, data)
    appStore.showSuccess(t('payment.admin.refundSuccess')); showRefundDialog.value = false; loadOrders()
  } catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { refundSubmitting.value = false }
}

function formatDateTime(dateStr: string): string { return formatOrderDateTime(dateStr) }

onMounted(() => loadOrders())
</script>

<style src="./orders-quench.css"></style>
<style scoped>
.spin-icon { animation: icon-spin .7s linear infinite; }
@keyframes icon-spin { to { transform: rotate(360deg); } }
</style>
