<template>
  <BaseDialog
    :show="show"
    :title="t('payment.admin.orderDetail')"
    width="wide"
    @close="emit('close')"
  >
    <div v-if="order" class="oq-dlg-grid">
      <div class="oq-dlg-field">
        <label>{{ t('payment.orders.orderId') }}</label>
        <p class="oq-mono">#{{ order.id }}</p>
      </div>
      <div class="oq-dlg-field">
        <label>{{ t('payment.orders.status') }}</label>
        <span :class="['oq-badge', statusBadgeQuench(order.status)]">
          {{ t('payment.status.' + order.status.toLowerCase(), order.status) }}
        </span>
      </div>
      <div class="oq-dlg-field">
        <label>{{ t('payment.orders.baseAmount') }}</label>
        <span class="oq-amount">¥{{ baseAmount.toFixed(2) }}</span>
      </div>
      <div v-if="order.fee_rate > 0" class="oq-dlg-field">
        <label>{{ t('payment.orders.fee') }} ({{ order.fee_rate }}%)</label>
        <span class="oq-amount">¥{{ feeAmount.toFixed(2) }}</span>
      </div>
      <div class="oq-dlg-field">
        <label>{{ t('payment.orders.payAmount') }}</label>
        <span class="oq-amount">¥{{ order.pay_amount.toFixed(2) }}</span>
      </div>
      <div v-if="order.amount !== order.pay_amount" class="oq-dlg-field">
        <label>{{ t('payment.orders.creditedAmount') }}</label>
        <span class="oq-amount">{{ order.order_type === 'balance' ? '$' : '¥' }}{{ order.amount.toFixed(2) }}</span>
      </div>
      <div class="oq-dlg-field">
        <label>{{ t('payment.orders.paymentMethod') }}</label>
        <p>{{ t('payment.methods.' + order.payment_type, order.payment_type) }}</p>
      </div>
      <div class="oq-dlg-field">
        <label>{{ t('payment.admin.orderType') }}</label>
        <p>{{ t('payment.admin.' + order.order_type + 'Order', order.order_type) }}</p>
      </div>
      <div class="oq-dlg-field">
        <label>{{ t('payment.orders.userId') }}</label>
        <p class="oq-mono oq-muted">#{{ order.user_id }}</p>
      </div>
      <div class="oq-dlg-field">
        <label>{{ t('payment.orders.createdAt') }}</label>
        <p class="oq-xs oq-muted">{{ formatDateTime(order.created_at) }}</p>
      </div>
      <div class="oq-dlg-field">
        <label>{{ t('payment.admin.expiresAt') }}</label>
        <p class="oq-xs oq-muted">{{ formatDateTime(order.expires_at) }}</p>
      </div>
      <div v-if="order.paid_at" class="oq-dlg-field">
        <label>{{ t('payment.admin.paidAt') }}</label>
        <p class="oq-xs oq-muted">{{ formatDateTime(order.paid_at) }}</p>
      </div>
      <div v-if="order.completed_at" class="oq-dlg-field">
        <label>{{ t('payment.admin.completedAt') }}</label>
        <p class="oq-xs oq-muted">{{ formatDateTime(order.completed_at) }}</p>
      </div>

      <!-- 退款信息 -->
      <div v-if="order.refund_amount" class="oq-refund-block">
        <h4>{{ t('payment.admin.refundInfo') }}</h4>
        <p><strong>{{ t('payment.admin.refundAmount') }}:</strong> {{ order.order_type === 'balance' ? '$' : '¥' }}{{ order.refund_amount.toFixed(2) }}</p>
        <p v-if="order.refund_reason">{{ t('payment.admin.refundReason') }}: {{ order.refund_reason }}</p>
      </div>

      <!-- 底部操作 -->
      <div class="oq-dlg-sep"></div>
      <div class="oq-dlg-section" style="display:flex;gap:8px;justify-content:flex-end">
        <button
          v-if="order.status === 'PENDING'"
          class="oq-btn oq-btn-warn"
          @click="emit('cancel', order)"
        >
          {{ t('payment.orders.cancel') }}
        </button>
        <button
          v-if="order.status === 'FAILED'"
          class="oq-btn"
          @click="emit('retry', order)"
        >
          {{ t('payment.admin.retry') }}
        </button>
        <button
          v-if="canRefund(order)"
          class="oq-btn oq-btn-danger"
          @click="emit('refund', order)"
        >
          {{ t('payment.admin.refund') }}
        </button>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { PaymentOrder } from '@/types/payment'
import { canRefund as canRefundStatus, formatOrderDateTime } from '@/components/payment/orderUtils'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  order: PaymentOrder | null
}>()

/** 充值金额 (base amount before fee) = pay_amount / (1 + fee_rate/100) */
const baseAmount = computed(() => {
  if (!props.order) return 0
  const feeRate = Number(props.order.fee_rate) || 0
  if (feeRate <= 0) return props.order.pay_amount
  return props.order.pay_amount / (1 + feeRate / 100)
})

/** 手续费 = pay_amount - baseAmount */
const feeAmount = computed(() => {
  if (!props.order) return 0
  const feeRate = Number(props.order.fee_rate) || 0
  if (feeRate <= 0) return 0
  return props.order.pay_amount - baseAmount.value
})

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'cancel', order: PaymentOrder): void
  (e: 'retry', order: PaymentOrder): void
  (e: 'refund', order: PaymentOrder): void
}>()

function canRefund(order: PaymentOrder): boolean {
  return canRefundStatus(order.status)
}

function statusBadgeQuench(status: string): string {
  const s = status.toUpperCase()
  if (s === 'COMPLETED' || s === 'PAID') return 'oq-badge-ok'
  if (s === 'PENDING' || s === 'REFUND_REQUESTED') return 'oq-badge-warn'
  if (s === 'FAILED' || s === 'REFUND_FAILED' || s === 'CANCELLED' || s === 'EXPIRED') return 'oq-badge-bad'
  if (s === 'REFUNDED') return 'oq-badge-azure'
  return 'oq-badge-dim'
}

function formatDateTime(dateStr: string): string {
  return formatOrderDateTime(dateStr)
}
</script>

<style src="../../../views/admin/orders/orders-quench.css"></style>
<style scoped>
/* QUENCH 局部样式 — 弹窗内 grid/badge/amount 复用全局 oq-* 类由 orders-quench.css 提供 */
/* 此处只声明 BaseDialog 穿透无法覆盖的微调 */
.oq-btn-warn {
  color: var(--warn, #E0B34E);
  border-color: rgba(224,179,78,.35);
  background: var(--warn-dim, rgba(224,179,78,.12));
}
.oq-btn-warn:hover { background: rgba(224,179,78,.2); }
</style>
