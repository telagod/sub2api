<template>
  <div class="oq-chart-card">
    <h3 class="oq-chart-title">{{ t('payment.admin.paymentDistribution') }}</h3>
    <div v-if="!methods?.length" class="oq-no-data" style="min-height:120px">
      {{ t('payment.admin.noData') }}
    </div>
    <div v-else>
      <div v-for="method in methods" :key="method.type" class="oq-method-row">
        <div class="oq-method-head">
          <div class="oq-method-name">
            <span class="oq-method-dot" :style="{ background: dotColor(method.type) }"></span>
            {{ t('payment.methods.' + method.type, method.type) }}
          </div>
          <div class="oq-method-right">
            <span class="oq-method-amt">${{ method.amount.toFixed(2) }}</span>
            <span class="oq-method-cnt">({{ method.count }})</span>
          </div>
        </div>
        <div class="oq-bar-track">
          <div class="oq-bar-fill" :style="{ width: barWidth(method.amount) + '%', background: dotColor(method.type) }"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  methods: { type: string; amount: number; count: number }[]
}>()

// ── QUENCH 图表配色（取自 tokens.css + mockup）──────────────────────────
// --azure: #5CA8FF  → alipay（主系蓝）
// --ok:    #46C98C  → wxpay（绿）
// #2E6FB8           → alipay_direct（深蓝，mockup .bar-fill）
// #3DAF7A           → wxpay_direct（深绿降级）
// --ink-1: #97A0AF  → stripe（钢银次系）
// --warn:  #E0B34E  → airwallex（琥珀）
const METHOD_COLORS: Record<string, string> = {
  alipay:        '#5CA8FF',  // --azure
  wxpay:         '#46C98C',  // --ok
  alipay_direct: '#2E6FB8',  // mockup deep-blue
  wxpay_direct:  '#3DAF7A',  // ok deep-green
  stripe:        '#97A0AF',  // --ink-1 钢银
  airwallex:     '#E0B34E',  // --warn
}

const maxAmount = computed(() => {
  if (!props.methods?.length) return 1
  return Math.max(...props.methods.map(m => m.amount), 1)
})

function dotColor(type: string): string {
  return METHOD_COLORS[type] || '#5C6470'  // fallback --ink-2
}

function barWidth(amount: number): number {
  return Math.min((amount / maxAmount.value) * 100, 100)
}
</script>
