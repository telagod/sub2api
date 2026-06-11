<template>
  <div class="ud-tab-content">
    <!-- KPI 三格（去重：余额/消耗已在抽屉顶部 KPI 条展示，此处放更有信息量的活动指标） -->
    <div class="ud-kpi-row">
      <div class="ud-kpi-card">
        <span class="ud-kpi-label">{{ t('admin.userTabs.kpiMonthRequests') }}</span>
        <span class="ud-kpi-value" v-if="!statsLoading">{{ monthStats.total_requests.toLocaleString() }}</span>
        <span class="ud-kpi-value ud-muted" v-else>…</span>
      </div>
      <div class="ud-kpi-card">
        <span class="ud-kpi-label">{{ t('admin.userTabs.kpiMonthTokens') }}</span>
        <span class="ud-kpi-value ud-mono" v-if="!statsLoading">{{ fmtTok(monthStats.total_tokens) }}</span>
        <span class="ud-kpi-value ud-muted" v-else>…</span>
      </div>
      <div class="ud-kpi-card">
        <span class="ud-kpi-label">{{ t('admin.userTabs.kpiConcurrency') }}</span>
        <span class="ud-kpi-value ud-mono">{{ user.current_concurrency ?? 0 }}<span class="ud-kpi-sub-inline">/{{ user.concurrency }}</span></span>
      </div>
    </div>

    <!-- 近 30 日消耗折线图（SVG 简版） -->
    <div class="ud-chart-wrap">
      <p class="ud-section-label">{{ t('admin.userTabs.chart30dTitle') }}</p>
      <div v-if="chartLoading" class="ud-chart-placeholder">
        <svg class="ud-chart-empty-ico" width="28" height="28" viewBox="0 0 28 28" fill="none" aria-hidden="true">
          <rect x="3" y="6" width="22" height="16" rx="3" stroke="currentColor" stroke-width="1.3"/>
          <circle cx="14" cy="14" r="3" stroke="currentColor" stroke-width="1.3"/>
        </svg>
        {{ t('admin.userTabs.loading') }}
      </div>
      <div v-else-if="chartError" class="ud-chart-placeholder">
        <svg class="ud-chart-empty-ico" width="28" height="28" viewBox="0 0 28 28" fill="none" aria-hidden="true">
          <circle cx="14" cy="14" r="10" stroke="currentColor" stroke-width="1.3"/>
          <path d="M14 9v6M14 18v1" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
        </svg>
        <span class="ud-muted">{{ chartError }}</span>
      </div>
      <template v-else>
        <!-- 无数据体面空态 -->
        <div v-if="!chartPath" class="ud-chart-placeholder">
          <svg class="ud-chart-empty-ico" width="28" height="28" viewBox="0 0 28 28" fill="none" aria-hidden="true">
            <path d="M4 22L10 14L15 18L20 10L24 14" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round" stroke-dasharray="3 2"/>
            <rect x="3" y="6" width="22" height="16" rx="3" stroke="currentColor" stroke-width="1.3" opacity=".4"/>
          </svg>
          <span>{{ t('admin.userTabs.chartNoData') }}</span>
        </div>
        <!-- 有数据：azure 折线 + 渐变填充 -->
        <svg
          v-else
          class="ud-chart-svg"
          viewBox="0 0 480 100"
          preserveAspectRatio="none"
          :aria-label="t('admin.userTabs.chart30dTitle')"
        >
          <defs>
            <linearGradient id="ud-chart-grad" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="var(--azure,#5CA8FF)" stop-opacity="0.22"/>
              <stop offset="100%" stop-color="var(--azure,#5CA8FF)" stop-opacity="0"/>
            </linearGradient>
          </defs>
          <!-- 网格线 -->
          <line x1="0" y1="25" x2="480" y2="25" stroke="var(--line-0)" stroke-width="1"/>
          <line x1="0" y1="50" x2="480" y2="50" stroke="var(--line-0)" stroke-width="1"/>
          <line x1="0" y1="75" x2="480" y2="75" stroke="var(--line-0)" stroke-width="1"/>
          <!-- 面积填充（渐变） -->
          <path :d="chartFillPath" fill="url(#ud-chart-grad)"/>
          <!-- 折线 azure -->
          <path :d="chartPath" fill="none" stroke="var(--azure,#5CA8FF)" stroke-width="1.8" stroke-linejoin="round" stroke-linecap="round"/>
        </svg>
      </template>
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


function fmtTok(v: number) {
  if (v >= 1e9) return (v / 1e9).toFixed(2) + 'B'
  if (v >= 1e6) return (v / 1e6).toFixed(2) + 'M'
  if (v >= 1e3) return (v / 1e3).toFixed(2) + 'K'
  return Math.round(v).toLocaleString()
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
  background: var(--metal,linear-gradient(180deg,#15181E,#0E1014));
  border: 1px solid var(--line-0);
  border-radius: 10px;
  box-shadow: var(--edge-hi,inset 0 1px 0 rgba(255,255,255,.04));
}
.ud-kpi-label { font-size: 10.5px; font-weight: 600; letter-spacing: .06em; text-transform: uppercase; color: var(--ink-2); }
.ud-kpi-value { font-family: var(--font-mono,"IBM Plex Mono",monospace); font-size: 18px; font-weight: 700; color: var(--ink-0); font-variant-numeric: tabular-nums; }
.ud-kpi-sub-inline { font-size: 12px; color: var(--ink-2); margin-left: 2px; font-weight: 400; }

.ud-chart-wrap { display: flex; flex-direction: column; gap: 8px; }
.ud-section-label { font-size: 11.5px; color: var(--ink-2); margin: 0; }
.ud-chart-svg { width: 100%; height: 100px; display: block; }
.ud-chart-placeholder {
  height: 100px;
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  gap: 8px;
  font-size: 12px; color: var(--ink-2);
  background: var(--metal,linear-gradient(180deg,#15181E,#0E1014));
  border-radius: 8px; border: 1px solid var(--line-0);
  box-shadow: var(--edge-hi,inset 0 1px 0 rgba(255,255,255,.04));
}
.ud-chart-empty-ico { opacity: .3; color: var(--ink-2); }

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
