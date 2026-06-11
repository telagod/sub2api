<template>
  <div class="ud-tab-content">
    <!-- KPI 三格 -->
    <div class="ud-kpi-row">
      <div class="ud-kpi-card">
        <span class="ud-kpi-label">{{ t('admin.userTabs.kpiBalance') }}</span>
        <span class="ud-kpi-value q-money">${{ formatBal(user.balance) }}</span>
      </div>
      <div class="ud-kpi-card">
        <span class="ud-kpi-label">{{ t('admin.userTabs.kpiMonthCost') }}</span>
        <span class="ud-kpi-value q-money" v-if="!statsLoading">${{ formatBal(monthStats.total_cost) }}</span>
        <span class="ud-kpi-value ud-muted" v-else>…</span>
      </div>
      <div class="ud-kpi-card">
        <span class="ud-kpi-label">{{ t('admin.userTabs.kpiMonthRequests') }}</span>
        <span class="ud-kpi-value" v-if="!statsLoading">{{ monthStats.total_requests.toLocaleString() }}</span>
        <span class="ud-kpi-value ud-muted" v-else>…</span>
      </div>
    </div>

    <!-- 近 30 日消耗折线图（SVG 简版） -->
    <div class="ud-chart-wrap">
      <p class="ud-section-label">{{ t('admin.userTabs.chart30dTitle') }}</p>
      <div v-if="chartLoading" class="ud-chart-placeholder">{{ t('admin.userTabs.loading') }}</div>
      <div v-else-if="chartError" class="ud-chart-placeholder ud-muted">{{ chartError }}</div>
      <svg
        v-else
        class="ud-chart-svg"
        viewBox="0 0 480 100"
        preserveAspectRatio="none"
        :aria-label="t('admin.userTabs.chart30dTitle')"
      >
        <!-- 网格线 -->
        <line x1="0" y1="25" x2="480" y2="25" stroke="var(--line-0)" stroke-width="1"/>
        <line x1="0" y1="50" x2="480" y2="50" stroke="var(--line-0)" stroke-width="1"/>
        <line x1="0" y1="75" x2="480" y2="75" stroke="var(--line-0)" stroke-width="1"/>
        <!-- 面积填充 -->
        <path v-if="chartPath" :d="chartFillPath" fill="rgba(92,168,255,0.08)" />
        <!-- 折线 -->
        <path v-if="chartPath" :d="chartPath" fill="none" stroke="var(--azure)" stroke-width="1.5" stroke-linejoin="round" stroke-linecap="round"/>
        <!-- 无数据占位 -->
        <text v-if="!chartPath" x="240" y="55" text-anchor="middle" fill="var(--ink-2)" font-size="11">{{ t('admin.userTabs.chartNoData') }}</text>
      </svg>
    </div>

    <!-- 基础信息 -->
    <div class="ud-info-grid">
      <div class="ud-info-row">
        <span class="ud-info-key">{{ t('admin.userTabs.infoUserId') }}</span>
        <span class="ud-info-val ud-mono">#{{ user.id }}</span>
      </div>
      <div class="ud-info-row">
        <span class="ud-info-key">{{ t('admin.userTabs.infoEmail') }}</span>
        <span class="ud-info-val">{{ user.email }}</span>
      </div>
      <div class="ud-info-row" v-if="user.username">
        <span class="ud-info-key">{{ t('admin.userTabs.infoUsername') }}</span>
        <span class="ud-info-val">{{ user.username }}</span>
      </div>
      <div class="ud-info-row">
        <span class="ud-info-key">{{ t('admin.userTabs.infoConcurrency') }}</span>
        <span class="ud-info-val ud-mono">{{ user.concurrency }}</span>
      </div>
      <div class="ud-info-row" v-if="user.rpm_limit !== undefined">
        <span class="ud-info-key">{{ t('admin.userTabs.infoRpm') }}</span>
        <span class="ud-info-val ud-mono">{{ user.rpm_limit === 0 ? t('admin.userTabs.infoRpmUnlimited') : user.rpm_limit }}</span>
      </div>
      <div class="ud-info-row">
        <span class="ud-info-key">{{ t('admin.userTabs.infoRegistered') }}</span>
        <span class="ud-info-val ud-muted">{{ fmt(user.created_at) }}</span>
      </div>
      <div class="ud-info-row" v-if="user.last_active_at">
        <span class="ud-info-key">{{ t('admin.userTabs.infoLastActive') }}</span>
        <span class="ud-info-val ud-muted">{{ fmt(user.last_active_at) }}</span>
      </div>
      <div class="ud-info-row" v-if="user.notes">
        <span class="ud-info-key">{{ t('admin.userTabs.infoNotes') }}</span>
        <span class="ud-info-val">{{ user.notes }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { AdminUser } from '@/types'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const props = defineProps<{ user: AdminUser }>()

const statsLoading = ref(false)
const chartLoading = ref(false)
const chartError = ref<string | null>(null)
const monthStats = ref({ total_cost: 0, total_requests: 0, total_tokens: 0 })

// 日粒度数据点 [{date, cost}]
const dailyPoints = ref<{ date: string; cost: number }[]>([])

function formatBal(v: number) {
  if (!v) return '0.00'
  const s = v.toFixed(8).replace(/\.?0+$/, '')
  const parts = s.split('.')
  if (parts.length === 1) return s + '.00'
  if (parts[1].length < 2) return s + '0'
  return s
}

function fmt(iso: string | null | undefined) {
  if (!iso) return '-'
  return formatDateTime(iso)
}

// 构建折线 path（视口 480×100）
const chartPath = computed(() => {
  const pts = dailyPoints.value
  if (!pts.length) return ''
  const maxVal = Math.max(...pts.map(p => p.cost), 0.000001)
  const n = pts.length
  const coords = pts.map((p, i) => {
    const x = (i / (n - 1 || 1)) * 476 + 2
    const y = 96 - (p.cost / maxVal) * 86
    return [x, y] as [number, number]
  })
  return coords.reduce((acc, [x, y], i) => acc + (i === 0 ? `M${x},${y}` : ` L${x},${y}`), '')
})

const chartFillPath = computed(() => {
  if (!chartPath.value) return ''
  const pts = dailyPoints.value
  const n = pts.length
  if (!n) return ''
  const maxVal = Math.max(...pts.map(p => p.cost), 0.000001)
  const last = pts.map((p, i) => {
    const x = (i / (n - 1 || 1)) * 476 + 2
    const y = 96 - (p.cost / maxVal) * 86
    return [x, y] as [number, number]
  })
  return `${chartPath.value} L${last[last.length - 1][0]},98 L${last[0][0]},98 Z`
})

async function loadStats() {
  statsLoading.value = true
  try {
    const res = await adminAPI.users.getUserUsageStats(props.user.id, 'month')
    monthStats.value = res
  } catch { /* ignore */ } finally {
    statsLoading.value = false
  }
}

async function loadChart() {
  chartLoading.value = true
  chartError.value = null
  try {
    // 拉近 30 日的日粒度 usage stats
    const now = new Date()
    // end/start vars unused — sampling uses per-window date ranges below
    // 用 adminUsageAPI.getStats 按天分组（通过遍历 7 日 × 5 批）
    // 由于接口不支持日粒度分组，改为拉 30 天内逐日统计（不发 30 次请求）
    // 实际用单次 getStats + period=month 做概览，图表展示本月累计趋势占位
    // 若要精确日粒度需后端支持，此处用近似点（7个采样点）
    const promises = Array.from({ length: 5 }, (_, w) => {
      const eDate = new Date(now.getTime() - w * 6 * 86400000)
      const sDate = new Date(eDate.getTime() - 5 * 86400000)
      return adminAPI.usage.getStats({
        user_id: props.user.id,
        start_date: sDate.toISOString().split('T')[0],
        end_date: eDate.toISOString().split('T')[0],
      }).then(r => ({ date: eDate.toISOString().split('T')[0], cost: r.total_cost }))
    })
    const results = await Promise.all(promises)
    dailyPoints.value = results.sort((a, b) => a.date.localeCompare(b.date))
  } catch (e) {
    chartError.value = t('admin.userTabs.chartLoadError')
  } finally {
    chartLoading.value = false
  }
}

watch(() => props.user.id, () => { loadStats(); loadChart() })
onMounted(() => { loadStats(); loadChart() })
</script>

<style scoped>
.ud-tab-content { display: flex; flex-direction: column; gap: 20px; }

.ud-kpi-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}
.ud-kpi-card {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 14px 16px;
  background: var(--bg-2);
  border: 1px solid var(--line-0);
  border-radius: 10px;
}
.ud-kpi-label { font-size: 11px; color: var(--ink-2); letter-spacing: 0.03em; }
.ud-kpi-value { font-size: 16px; font-weight: 700; color: var(--ink-0); }

.ud-chart-wrap { display: flex; flex-direction: column; gap: 8px; }
.ud-section-label { font-size: 11.5px; color: var(--ink-2); margin: 0; }
.ud-chart-svg { width: 100%; height: 100px; display: block; }
.ud-chart-placeholder {
  height: 100px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: var(--ink-2);
  background: var(--bg-2);
  border-radius: 8px;
  border: 1px solid var(--line-0);
}

.ud-info-grid { display: flex; flex-direction: column; gap: 0; }
.ud-info-row {
  display: flex;
  align-items: baseline;
  gap: 12px;
  padding: 9px 0;
  border-bottom: 1px solid var(--line-0);
  font-size: 12.5px;
}
.ud-info-row:last-child { border-bottom: none; }
.ud-info-key { width: 80px; flex-shrink: 0; color: var(--ink-2); font-size: 11.5px; }
.ud-info-val { color: var(--ink-0); flex: 1; word-break: break-all; }
.ud-mono { font-family: 'IBM Plex Mono', monospace; font-size: 12px; }
.ud-muted { color: var(--ink-2); }
</style>
