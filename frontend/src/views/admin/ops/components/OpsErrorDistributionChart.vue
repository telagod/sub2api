<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Chart as ChartJS, ArcElement, Legend, Tooltip } from 'chart.js'
import { Doughnut } from 'vue-chartjs'
import type { OpsErrorDistributionResponse } from '@/api/admin/ops'
import type { ChartState } from '../types'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import EmptyState from '@/components/common/EmptyState.vue'

ChartJS.register(ArcElement, Tooltip, Legend)

interface Props {
  data: OpsErrorDistributionResponse | null
  loading: boolean
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'openDetails'): void
}>()
const { t } = useI18n()

const colors = computed(() => ({
  blue: '#5CA8FF',     /* azure 主系 */
  red: '#F25C69',      /* bad */
  orange: '#E0B34E',   /* warn */
  gray: '#97A0AF',     /* steel 次系 */
  text: '#5C6470'      /* ink-2 */
}))

const totalSlaErrors = computed(() =>
  (props.data?.items ?? []).reduce((total, item) => total + Number(item.sla || 0), 0)
)

const hasData = computed(() => totalSlaErrors.value > 0)

const state = computed<ChartState>(() => {
  if (hasData.value) return 'ready'
  if (props.loading) return 'loading'
  return 'empty'
})

interface ErrorCategory {
  label: string
  count: number
  color: string
}

const categories = computed<ErrorCategory[]>(() => {
  if (!props.data) return []

  let upstream = 0 // 502, 503, 504
  let client = 0 // 4xx
  let system = 0 // 500
  let other = 0

  for (const item of props.data.items || []) {
    const code = Number(item.status_code || 0)
    const count = Number(item.sla || 0)
    if (!Number.isFinite(code) || !Number.isFinite(count)) continue

    if ([502, 503, 504].includes(code)) upstream += count
    else if (code >= 400 && code < 500) client += count
    else if (code === 500) system += count
    else other += count
  }

  const out: ErrorCategory[] = []
  if (upstream > 0) out.push({ label: t('admin.ops.upstream'), count: upstream, color: colors.value.orange })
  if (client > 0) out.push({ label: t('admin.ops.client'), count: client, color: colors.value.blue })
  if (system > 0) out.push({ label: t('admin.ops.system'), count: system, color: colors.value.red })
  if (other > 0) out.push({ label: t('admin.ops.other'), count: other, color: colors.value.gray })
  return out
})

const topReason = computed(() => {
  if (categories.value.length === 0) return null
  return categories.value.reduce((prev, cur) => (cur.count > prev.count ? cur : prev))
})

const chartData = computed(() => {
  if (!hasData.value || categories.value.length === 0) return null
  return {
    labels: categories.value.map((c) => c.label),
    datasets: [
      {
        data: categories.value.map((c) => c.count),
        backgroundColor: categories.value.map((c) => c.color),
        borderWidth: 0
      }
    ]
  }
})

const options = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: '#111111',
      titleColor: '#ededed',
      bodyColor: '#a3a3a3'
    }
  }
}))
</script>

<template>
  <div class="od-chart-card">
    <div class="od-chart-head">
      <h3 class="od-chart-title">
        <svg class="od-chart-icon" width="14" height="14" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        {{ t('admin.ops.errorDistribution') }}
        <HelpTooltip :content="t('admin.ops.tooltips.errorDistribution')" />
      </h3>
      <button type="button" class="od-btn" style="padding:3px 8px;font-size:11px;" :disabled="state !== 'ready'" :title="t('admin.ops.errorTrend')" @click="emit('openDetails')">
        {{ t('admin.ops.requestDetails.details') }}
      </button>
    </div>

    <div style="position:relative;flex:1;min-height:0;">
      <div v-if="state === 'ready' && chartData" style="display:flex;flex-direction:column;height:100%;">
        <div style="flex:1;">
          <Doughnut :data="chartData" :options="{ ...options, cutout: '65%' }" />
        </div>
        <div style="margin-top:12px;display:flex;flex-direction:column;align-items:center;gap:6px;">
          <div v-if="topReason" style="font-size:11.5px;font-weight:700;color:var(--ink-0,#E8EBF0);">
            {{ t('admin.ops.top') }}: <span :style="{ color: topReason.color }">{{ topReason.label }}</span>
          </div>
          <div style="display:flex;flex-wrap:wrap;justify-content:center;gap:8px;">
            <div v-for="item in categories" :key="item.label" style="display:flex;align-items:center;gap:5px;font-size:11px;">
              <span style="width:7px;height:7px;border-radius:50%;display:inline-block;flex-shrink:0;" :style="{ backgroundColor: item.color }"></span>
              <span style="color:var(--ink-2,#5C6470);">{{ item.count }}</span>
            </div>
          </div>
        </div>
      </div>

      <div v-else style="display:flex;height:100%;align-items:center;justify-content:center;">
        <div v-if="state === 'loading'" style="font-size:13px;color:var(--ink-2,#5C6470);" class="animate-pulse">{{ t('common.loading') }}</div>
        <EmptyState v-else :title="t('common.noData')" :description="t('admin.ops.charts.emptyError')" />
      </div>
    </div>
  </div>
</template>

<style src="../ops-quench.css"></style>
