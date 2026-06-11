<template>
  <BaseDialog
    :show="show"
    :title="t('payment.admin.refundOrder')"
    width="normal"
    @close="emit('cancel')"
  >
    <form id="refund-form" class="oq-refund-form" @submit.prevent="handleSubmit">
      <!-- 退款申请信息 -->
      <div
        v-if="order?.refund_requested_at || order?.refund_request_reason"
        class="oq-info-block"
        style="border-color:rgba(92,168,255,.2);background:var(--azure-dim,rgba(92,168,255,.08))"
      >
        <div style="display:flex;align-items:center;gap:7px;font-size:12.5px;font-weight:600;color:var(--azure,#5CA8FF);margin-bottom:6px">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none"><circle cx="7" cy="7" r="5.5" stroke="currentColor" stroke-width="1.2"/><path d="M7 6.5v3M7 4.5h.01" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
          {{ t('payment.admin.refundRequestInfo') }}
        </div>
        <div v-if="order?.refund_requested_at" class="oq-info-row">
          <span>{{ t('payment.admin.refundRequestedAt') }}</span>
          <span>{{ formatDateTime(order.refund_requested_at) }}</span>
        </div>
        <div v-if="order?.refund_request_reason" class="oq-info-row" style="flex-direction:column;align-items:flex-start;gap:3px">
          <span>{{ t('payment.admin.refundRequestReason') }}:</span>
          <span style="color:var(--ink-0)">{{ order.refund_request_reason }}</span>
        </div>
      </div>

      <!-- 订单信息 -->
      <div class="oq-info-block">
        <div class="oq-info-row">
          <span>{{ t('payment.orders.orderId') }}</span>
          <span class="oq-mono">#{{ order?.id }}</span>
        </div>
        <div class="oq-info-row">
          <span>{{ t('payment.orders.creditedAmount') }}</span>
          <span>{{ order?.order_type === 'balance' ? '$' : '¥' }}{{ order?.amount?.toFixed(2) }}</span>
        </div>
        <div class="oq-info-row">
          <span>{{ t('payment.orders.payAmount') }}</span>
          <span>¥{{ order?.pay_amount?.toFixed(2) }}</span>
        </div>
        <div v-if="actuallyRefunded > 0" class="oq-info-row">
          <span>{{ t('payment.admin.alreadyRefunded') }}</span>
          <span class="c-bad">{{ order?.order_type === 'balance' ? '$' : '¥' }}{{ actuallyRefunded.toFixed(2) }}</span>
        </div>
      </div>

      <!-- 扣减余额 -->
      <div class="oq-form-group">
        <div class="oq-checkbox-row">
          <input id="deduct-balance" v-model="form.deduct_balance" type="checkbox" />
          <label for="deduct-balance">{{ t('payment.admin.deductBalance') }}</label>
          <span class="oq-checkbox-hint">{{ t('payment.admin.deductBalanceHint') }}</span>
        </div>

        <!-- 用户余额信息 -->
        <div v-if="form.deduct_balance && userBalance != null" style="display:grid;grid-template-columns:1fr 1fr;gap:10px;margin-top:8px">
          <div class="oq-info-block" style="padding:10px 12px">
            <div class="oq-hint" style="margin-bottom:4px">{{ t('payment.admin.userBalance') }}</div>
            <div class="oq-mono" style="color:var(--money);font-size:14px;font-weight:700">${{ userBalance.toFixed(2) }}</div>
          </div>
          <div class="oq-info-block" style="padding:10px 12px">
            <div class="oq-hint" style="margin-bottom:4px">{{ t('payment.admin.orderAmount') }}</div>
            <div class="oq-mono" style="color:var(--money);font-size:14px;font-weight:700">{{ order?.order_type === 'balance' ? '$' : '¥' }}{{ order?.amount?.toFixed(2) }}</div>
          </div>
        </div>

        <!-- 余额不足警告 -->
        <div v-if="form.deduct_balance && balanceInsufficient" class="oq-warn-block" style="margin-top:8px">
          {{ t('payment.admin.insufficientBalance') }}
        </div>

        <!-- 不扣减说明 -->
        <div v-if="!form.deduct_balance" class="oq-info-block" style="margin-top:8px;font-size:12.5px">
          {{ t('payment.admin.noDeduction') }}
        </div>
      </div>

      <!-- 退款金额 -->
      <div class="oq-form-group">
        <label class="oq-label">{{ t('payment.admin.refundAmount') }}</label>
        <div class="oq-input-prefix">
          <span class="oq-pfx">{{ order?.order_type === 'balance' ? '$' : '¥' }}</span>
          <input
            v-model.number="form.amount"
            class="oq-input"
            type="number"
            step="0.01"
            min="0.01"
            :max="maxRefundable"
            required
          />
        </div>
        <p class="oq-hint">{{ t('payment.admin.maxRefundable') }}: {{ order?.order_type === 'balance' ? '$' : '¥' }}{{ maxRefundable.toFixed(2) }}</p>
      </div>

      <!-- 退款原因 -->
      <div class="oq-form-group">
        <label class="oq-label">{{ t('payment.admin.refundReason') }}</label>
        <textarea
          v-model="form.reason"
          class="oq-input"
          rows="3"
          :placeholder="t('payment.admin.refundReasonPlaceholder')"
          required
        ></textarea>
      </div>

      <!-- 警告信息 -->
      <div v-if="warning" class="oq-warn-block">{{ warning }}</div>

      <!-- 强制退款 -->
      <div v-if="requireForce" class="oq-checkbox-row">
        <input id="force-refund" v-model="form.force" type="checkbox" style="accent-color:var(--bad,#F25C69)" />
        <label for="force-refund" style="color:var(--bad,#F25C69);font-weight:600">{{ t('payment.admin.forceRefund') }}</label>
      </div>
    </form>

    <template #footer>
      <div style="display:flex;justify-content:flex-end;gap:10px">
        <button type="button" class="oq-btn" @click="emit('cancel')">
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          form="refund-form"
          class="oq-btn oq-btn-danger"
          :disabled="submitting || form.amount <= 0 || (requireForce && !form.force)"
        >
          {{ submitting ? t('common.processing') : t('payment.admin.confirmRefund') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { PaymentOrder } from '@/types/payment'
import { formatOrderDateTime } from '@/components/payment/orderUtils'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  order: PaymentOrder | null
  submitting?: boolean
  userBalance?: number | null
  requireForce?: boolean
  warning?: string
}>()

const emit = defineEmits<{
  (e: 'confirm', data: { amount: number; reason: string; deduct_balance: boolean; force: boolean }): void
  (e: 'cancel'): void
}>()

const form = reactive({
  amount: 0,
  reason: '',
  deduct_balance: true,
  force: false,
})

// In REFUND_REQUESTED status, refund_amount is the REQUESTED amount, not actually refunded.
// Only PARTIALLY_REFUNDED / REFUNDED have real refund amounts.
const actuallyRefunded = computed(() => {
  if (!props.order) return 0
  const s = props.order.status
  if (s === 'PARTIALLY_REFUNDED' || s === 'REFUNDED') return props.order.refund_amount || 0
  return 0
})

const maxRefundable = computed(() => {
  if (!props.order) return 0
  return props.order.amount - actuallyRefunded.value
})

const balanceInsufficient = computed(() => {
  if (props.userBalance == null || !props.order) return false
  return props.userBalance < props.order.amount
})

watch(() => props.show, (val) => {
  if (val && props.order) {
    // For REFUND_REQUESTED, pre-fill with the requested amount
    if (props.order.status === 'REFUND_REQUESTED' && props.order.refund_amount) {
      form.amount = props.order.refund_amount
    } else {
      form.amount = maxRefundable.value
    }
    form.reason = props.order.refund_request_reason || ''
    form.deduct_balance = true
    form.force = false
  }
})

function formatDateTime(dateStr: string): string {
  return formatOrderDateTime(dateStr)
}

function handleSubmit() {
  if (form.amount <= 0 || form.amount > maxRefundable.value) return
  if (props.requireForce && !form.force) return
  emit('confirm', { ...form })
}
</script>

<style scoped>
.oq-refund-form { display: flex; flex-direction: column; gap: 14px; }
</style>
