<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Chart as ChartJS, CategoryScale, Filler, Legend, LineElement, LinearScale, PointElement, Title, Tooltip } from 'chart.js'
import { Line } from 'vue-chartjs'
import type { ChartComponentRef } from 'vue-chartjs'
import type { OpsThroughputGroupBreakdownItem, OpsThroughputPlatformBreakdownItem, OpsThroughputTrendPoint } from '@/api/admin/ops'
import type { ChartState } from '../types'
import { formatHistoryLabel, sumNumbers } from '../utils/opsFormatters'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { formatNumber } from '@/utils/format'

ChartJS.register(Title, Tooltip, Legend, LineElement, LinearScale, PointElement, CategoryScale, Filler)

interface Props {
  points: OpsThroughputTrendPoint[]
  loading: boolean
  timeRange: string
  byPlatform?: OpsThroughputPlatformBreakdownItem[]
  topGroups?: OpsThroughputGroupBreakdownItem[]
  fullscreen?: boolean
}

const props = defineProps<Props>()
const { t } = useI18n()
const emit = defineEmits<{
  (e: 'selectPlatform', platform: string): void
  (e: 'selectGroup', groupId: number): void
  (e: 'openDetails'): void
}>()

const throughputChartRef = ref<ChartComponentRef | null>(null)
watch(
  () => props.timeRange,
  () => {
    setTimeout(() => {
      const chart: any = throughputChartRef.value?.chart
      if (chart && typeof chart.resetZoom === 'function') {
        chart.resetZoom()
      }
    }, 100)
  }
)

const colors = computed(() => ({
  blue: '#5CA8FF',              /* azure 主系 */
  blueAlpha: 'rgba(92,168,255,.12)',
  green: '#46C98C',             /* ok 次系 */
  greenAlpha: 'rgba(70,201,140,.12)',
  grid: '#20242C',              /* line-0 */
  text: '#5C6470'               /* ink-2 */
}))

const totalRequests = computed(() => sumNumbers(props.points.map((p) => p.request_count)))

const chartData = computed(() => {
  if (!props.points.length || totalRequests.value <= 0) return null
  return {
    labels: props.points.map((p) => formatHistoryLabel(p.bucket_start, props.timeRange)),
    datasets: [
      {
        label: 'QPS',
        data: props.points.map((p) => p.qps ?? 0),
        borderColor: colors.value.blue,
        backgroundColor: colors.value.blueAlpha,
        fill: true,
        tension: 0.4,
        pointRadius: 0,
        pointHitRadius: 10
      },
      {
        label: t('admin.ops.tpsK'),
        data: props.points.map((p) => (p.tps ?? 0) / 1000),
        borderColor: colors.value.green,
        backgroundColor: colors.value.greenAlpha,
        fill: true,
        tension: 0.4,
        pointRadius: 0,
        pointHitRadius: 10,
        yAxisID: 'y1'
      }
    ]
  }
})

const state = computed<ChartState>(() => {
  if (chartData.value) return 'ready'
  if (props.loading) return 'loading'
  return 'empty'
})

const options = computed(() => {
  const c = colors.value
  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { intersect: false, mode: 'index' as const },
    plugins: {
      legend: {
        position: 'top' as const,
        align: 'end' as const,
        labels: { color: c.text, usePointStyle: true, boxWidth: 6, font: { size: 10 } }
      },
      tooltip: {
        backgroundColor: '#111111',
        titleColor: '#ededed',
        bodyColor: '#a3a3a3',
        borderColor: c.grid,
        borderWidth: 1,
        padding: 10,
        displayColors: true,
        callbacks: {
          label: (context: any) => {
            let label = context.dataset.label || ''
            if (label) label += ': '
            if (context.raw !== null) label += context.parsed.y.toFixed(1)
            return label
          }
        }
      },
      // Optional: if chartjs-plugin-zoom is installed, these options will enable zoom/pan.
      zoom: {
        pan: { enabled: true, mode: 'x' as const, modifierKey: 'ctrl' as const },
        zoom: { wheel: { enabled: true }, pinch: { enabled: true }, mode: 'x' as const }
      }
    },
    scales: {
      x: {
        type: 'category' as const,
        grid: { display: false },
        ticks: {
          color: c.text,
          font: { size: 10 },
          maxTicksLimit: 8,
          autoSkip: true,
          autoSkipPadding: 10
        }
      },
      y: {
        type: 'linear' as const,
        display: true,
        position: 'left' as const,
        grid: { color: c.grid, borderDash: [4, 4] },
        ticks: { color: c.text, font: { size: 10 } }
      },
      y1: {
        type: 'linear' as const,
        display: true,
        position: 'right' as const,
        grid: { display: false },
        ticks: { color: c.green, font: { size: 10 } }
      }
    }
  }
})

function resetZoom() {
  const chart: any = throughputChartRef.value?.chart
  if (chart && typeof chart.resetZoom === 'function') chart.resetZoom()
}

function downloadChart() {
  const chart: any = throughputChartRef.value?.chart
  if (!chart || typeof chart.toBase64Image !== 'function') return
  const url = chart.toBase64Image('image/png', 1)
  const a = document.createElement('a')
  a.href = url
  a.download = `ops-throughput-${new Date().toISOString().slice(0, 19).replace(/[:T]/g, '-')}.png`
  a.click()
}
</script>

<template>
  <div class="od-chart-card">
    <div class="od-chart-head">
      <h3 class="od-chart-title">
        <svg class="od-chart-icon" width="14" height="14" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
        </svg>
        {{ t('admin.ops.throughputTrend') }}
        <HelpTooltip v-if="!props.fullscreen" :content="t('admin.ops.tooltips.throughputTrend')" />
      </h3>
      <div style="display:flex;align-items:center;gap:6px;font-size:11px;color:var(--ink-2,#5C6470);">
        <span style="display:inline-flex;align-items:center;gap:4px;"><span style="width:7px;height:7px;border-radius:50%;background:var(--ops-azure);display:inline-block;"></span>QPS</span>
        <span style="display:inline-flex;align-items:center;gap:4px;"><span style="width:7px;height:7px;border-radius:50%;background:var(--ops-ok);display:inline-block;"></span>{{ t('admin.ops.tpsK') }}</span>
        <template v-if="!props.fullscreen">
          <button type="button" class="od-btn" style="padding:3px 8px;font-size:11px;" :disabled="state !== 'ready'" :title="t('admin.ops.requestDetails.title')" @click="emit('openDetails')">
            {{ t('admin.ops.requestDetails.details') }}
          </button>
          <button type="button" class="od-btn" style="padding:3px 8px;font-size:11px;" :disabled="state !== 'ready'" :title="t('admin.ops.charts.resetZoomHint')" @click="resetZoom">
            {{ t('admin.ops.charts.resetZoom') }}
          </button>
          <button type="button" class="od-btn" style="padding:3px 8px;font-size:11px;" :disabled="state !== 'ready'" :title="t('admin.ops.charts.downloadChartHint')" @click="downloadChart">
            {{ t('admin.ops.charts.downloadChart') }}
          </button>
        </template>
      </div>
    </div>

    <!-- Drilldown chips -->
    <div v-if="(props.topGroups?.length ?? 0) > 0" style="display:flex;flex-wrap:wrap;gap:6px;margin-bottom:10px;">
      <button
        v-for="g in props.topGroups"
        :key="g.group_id"
        type="button"
        class="od-badge od-badge-dim"
        style="cursor:pointer;border-radius:100px;"
        @click="emit('selectGroup', g.group_id)"
      >
        <span style="max-width:180px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">{{ g.group_name || `#${g.group_id}` }}</span>
        <span style="color:var(--ink-2,#5C6470);">{{ formatNumber(g.request_count) }}</span>
      </button>
    </div>

    <div v-else-if="(props.byPlatform?.length ?? 0) > 0" style="display:flex;flex-wrap:wrap;gap:6px;margin-bottom:10px;">
      <button
        v-for="p in props.byPlatform"
        :key="p.platform"
        type="button"
        class="od-badge od-badge-dim"
        style="cursor:pointer;border-radius:100px;"
        @click="emit('selectPlatform', p.platform)"
      >
        <span style="text-transform:uppercase;">{{ p.platform }}</span>
        <span style="color:var(--ink-2,#5C6470);">{{ formatNumber(p.request_count) }}</span>
      </button>
    </div>

    <div style="flex:1;min-height:0;">
      <Line v-if="state === 'ready' && chartData" ref="throughputChartRef" :data="chartData" :options="options" />
      <div v-else style="display:flex;height:100%;align-items:center;justify-content:center;">
        <div v-if="state === 'loading'" style="font-size:13px;color:var(--ink-2,#5C6470);" class="animate-pulse">{{ t('common.loading') }}</div>
        <EmptyState v-else :title="t('common.noData')" :description="t('admin.ops.charts.emptyRequest')" />
      </div>
    </div>
  </div>
</template>

<style src="../ops-quench.css"></style>
