<template>
  <div class="oq-chart-card">
    <h3 class="oq-chart-title">{{ t('payment.admin.dailyRevenue') }}</h3>
    <div class="oq-chart-h">
      <div v-if="loading" class="oq-no-data">
        <div class="oq-spinner"></div>
      </div>
      <Line v-else-if="chartData" :data="chartData" :options="chartOptions" />
      <div v-else class="oq-no-data">{{ t('payment.admin.noData') }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler)

const { t } = useI18n()

// ── QUENCH 图表配色（取自 tokens.css）──────────────────────────────────
// --azure: #5CA8FF  → 主数据线（收入）
// --ok:    #46C98C  → 次数据线（订单数）
// --bg-2:  #171A20  → 网格线 / 轴线
// --ink-2: #5C6470  → 轴文字
// --ink-1: #97A0AF  → 图例文字
const CHART_COLORS = {
  azureLine:      '#5CA8FF',                       // --azure
  azureFill:      'rgba(92,168,255,0.08)',          // --azure-dim × 0.66
  okLine:         '#46C98C',                        // --ok
  okFill:         'rgba(70,201,140,0.06)',          // --ok-dim × 0.4
  gridLine:       'rgba(32,36,44,0.8)',             // --line-0
  axisText:       '#5C6470',                        // --ink-2
  legendText:     '#97A0AF',                        // --ink-1
  tooltipBg:      '#171A20',                        // --bg-2
  tooltipBorder:  '#2F3540',                        // --line-1
} as const

const props = defineProps<{
  data: { date: string; amount: number; count: number }[]
  loading?: boolean
}>()

const chartData = computed(() => {
  if (!props.data || props.data.length === 0) return null
  return {
    labels: props.data.map(d => d.date),
    datasets: [
      {
        label: t('payment.admin.revenue'),
        data: props.data.map(d => d.amount),
        borderColor: CHART_COLORS.azureLine,
        backgroundColor: CHART_COLORS.azureFill,
        fill: true,
        tension: 0.35,
        pointRadius: 3,
        pointHoverRadius: 5,
        pointBackgroundColor: CHART_COLORS.azureLine,
        borderWidth: 1.8,
      },
      {
        label: t('payment.admin.orderCount'),
        data: props.data.map(d => d.count),
        borderColor: CHART_COLORS.okLine,
        backgroundColor: CHART_COLORS.okFill,
        fill: false,
        tension: 0.35,
        pointRadius: 3,
        pointHoverRadius: 5,
        pointBackgroundColor: CHART_COLORS.okLine,
        borderWidth: 1.5,
        yAxisID: 'y1',
      }
    ]
  }
})

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  interaction: { mode: 'index' as const, intersect: false },
  scales: {
    x: {
      grid: { color: CHART_COLORS.gridLine },
      ticks: { color: CHART_COLORS.axisText, font: { size: 11 } },
    },
    y: {
      type: 'linear' as const,
      display: true,
      position: 'left' as const,
      title: { display: true, text: t('payment.admin.revenue'), color: CHART_COLORS.axisText, font: { size: 11 } },
      grid: { color: CHART_COLORS.gridLine },
      ticks: { color: CHART_COLORS.axisText, font: { size: 11 } },
    },
    y1: {
      type: 'linear' as const,
      display: true,
      position: 'right' as const,
      title: { display: true, text: t('payment.admin.orderCount'), color: CHART_COLORS.axisText, font: { size: 11 } },
      grid: { drawOnChartArea: false },
      ticks: { color: CHART_COLORS.axisText, font: { size: 11 } },
    }
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: { color: CHART_COLORS.legendText, font: { size: 12 }, boxWidth: 10, boxHeight: 3, useBorderRadius: true, borderRadius: 2 },
    },
    tooltip: {
      backgroundColor: CHART_COLORS.tooltipBg,
      borderColor: CHART_COLORS.tooltipBorder,
      borderWidth: 1,
      titleColor: '#E8EBF0',  // --ink-0
      bodyColor: '#97A0AF',   // --ink-1
      padding: 10,
    }
  }
}
</script>
