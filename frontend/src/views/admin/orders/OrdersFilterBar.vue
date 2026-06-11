<template>
  <div class="oq-filter">
    <!-- 搜索框 -->
    <div class="oq-search" :class="{ 'oq-search-focus': focused }">
      <svg width="13" height="13" viewBox="0 0 13 13" fill="none" aria-hidden="true"><circle cx="5.5" cy="5.5" r="4" stroke="currentColor" stroke-width="1.2"/><path d="M9 9L11.5 11.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
      <input
        :value="search"
        class="oq-search-input"
        placeholder="搜索订单号 / 用户…"
        @focus="focused = true"
        @blur="focused = false"
        @input="$emit('update:search', ($event.target as HTMLInputElement).value)"
      />
    </div>

    <!-- 状态 -->
    <div class="oq-chip-wrap">
      <button class="oq-chip" :class="{ 'oq-chip-on': status }" @click.stop="showStatus = !showStatus; showPayType = false; showOrderType = false">
        状态 <b>{{ statusLabel }}</b>
        <svg width="10" height="10" viewBox="0 0 10 10" fill="none"><path d="M2 3.5L5 6.5L8 3.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
      </button>
      <div v-if="showStatus" class="oq-menu" @click.stop>
        <button v-for="opt in statusOptions" :key="opt.value" class="oq-menu-item" :class="{ on: status === opt.value }" @click="$emit('update:status', opt.value); showStatus = false">{{ opt.label }}</button>
      </div>
    </div>

    <!-- 支付方式 -->
    <div class="oq-chip-wrap">
      <button class="oq-chip" :class="{ 'oq-chip-on': payType }" @click.stop="showPayType = !showPayType; showStatus = false; showOrderType = false">
        支付 <b>{{ payTypeLabel }}</b>
        <svg width="10" height="10" viewBox="0 0 10 10" fill="none"><path d="M2 3.5L5 6.5L8 3.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
      </button>
      <div v-if="showPayType" class="oq-menu" @click.stop>
        <button v-for="opt in payTypeOptions" :key="opt.value" class="oq-menu-item" :class="{ on: payType === opt.value }" @click="$emit('update:payType', opt.value); showPayType = false">{{ opt.label }}</button>
      </div>
    </div>

    <!-- 订单类型 -->
    <div class="oq-chip-wrap">
      <button class="oq-chip" :class="{ 'oq-chip-on': orderType }" @click.stop="showOrderType = !showOrderType; showStatus = false; showPayType = false">
        类型 <b>{{ orderTypeLabel }}</b>
        <svg width="10" height="10" viewBox="0 0 10 10" fill="none"><path d="M2 3.5L5 6.5L8 3.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
      </button>
      <div v-if="showOrderType" class="oq-menu" @click.stop>
        <button v-for="opt in orderTypeOptions" :key="opt.value" class="oq-menu-item" :class="{ on: orderType === opt.value }" @click="$emit('update:orderType', opt.value); showOrderType = false">{{ opt.label }}</button>
      </div>
    </div>

    <!-- 清空 -->
    <button v-if="hasFilters" class="oq-clear-all" @click="$emit('clear')">清空筛选</button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  search: string
  status: string
  payType: string
  orderType: string
}>()

defineEmits<{
  'update:search': [v: string]
  'update:status': [v: string]
  'update:payType': [v: string]
  'update:orderType': [v: string]
  'clear': []
}>()

const focused = ref(false)
const showStatus = ref(false)
const showPayType = ref(false)
const showOrderType = ref(false)

const statusOptions = computed(() => [
  { value: '', label: '全部状态' },
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

const payTypeOptions = computed(() => [
  { value: '', label: t('payment.admin.allPaymentTypes') },
  { value: 'alipay', label: t('payment.methods.alipay') },
  { value: 'wxpay', label: t('payment.methods.wxpay') },
  { value: 'stripe', label: t('payment.methods.stripe') },
  { value: 'airwallex', label: t('payment.methods.airwallex') },
])

const orderTypeOptions = computed(() => [
  { value: '', label: t('payment.admin.allOrderTypes') },
  { value: 'balance', label: t('payment.admin.balanceOrder') },
  { value: 'subscription', label: t('payment.admin.subscriptionOrder') },
])

const statusLabel = computed(() => statusOptions.value.find(o => o.value === props.status)?.label ?? '全部')
const payTypeLabel = computed(() => payTypeOptions.value.find(o => o.value === props.payType)?.label ?? '全部')
const orderTypeLabel = computed(() => orderTypeOptions.value.find(o => o.value === props.orderType)?.label ?? '全部')
const hasFilters = computed(() => !!(props.search || props.status || props.payType || props.orderType))

function onDocClick() { showStatus.value = false; showPayType.value = false; showOrderType.value = false }
onMounted(() => document.addEventListener('click', onDocClick))
onUnmounted(() => document.removeEventListener('click', onDocClick))
</script>
