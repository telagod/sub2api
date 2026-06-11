<template>
  <div
    class="mc-root"
    @mouseenter="onEnter"
    @mouseleave="onLeave"
  >
    <!-- input 价 -->
    <div class="mc-row">
      <span class="mc-label">in</span>
      <span
        class="q-money mc-price"
        :style="priceColorStyle(discountIn)"
      >{{ fmtPrice(cell.inputPrice) }}</span>
      <span
        v-if="discountIn !== null"
        class="mc-badge"
        :style="discountStyle(discountIn)"
      >{{ discountLabel(discountIn) }}</span>
    </div>

    <!-- output 价 -->
    <div class="mc-row">
      <span class="mc-label">out</span>
      <span
        class="q-money mc-price"
        :style="priceColorStyle(discountOut)"
      >{{ fmtPrice(cell.outputPrice) }}</span>
      <span
        v-if="discountOut !== null"
        class="mc-badge"
        :style="discountStyle(discountOut)"
      >{{ discountLabel(discountOut) }}</span>
    </div>

    <!-- 分档角标 -->
    <div
      v-if="cell.hasIntervals"
      class="mc-tier-badge"
      @click.stop="openIntervals"
    >
      <span :style="{ color: 'var(--azure)' }">▾</span>
    </div>

    <!-- hover tooltip：拆分价 + 官方价对比 -->
    <Teleport to="body">
      <div
        v-if="showTooltip && tooltipReady"
        class="mc-tooltip"
        :style="{ top: tooltipTop + 'px', left: tooltipLeft + 'px' }"
      >
        <div class="mc-tt-title">{{ model }}</div>
        <div class="mc-tt-sep"></div>
        <!-- 当前价 -->
        <div class="mc-tt-section">
          <div class="mc-tt-row">
            <span class="mc-tt-key">{{ t('admin.pricingDesk.ttInput') }}</span>
            <span class="q-money mc-tt-val" :style="priceColorStyle(discountIn)">{{ fmtPrice(cell.inputPrice) }}</span>
          </div>
          <div class="mc-tt-row">
            <span class="mc-tt-key">{{ t('admin.pricingDesk.ttOutput') }}</span>
            <span class="q-money mc-tt-val" :style="priceColorStyle(discountOut)">{{ fmtPrice(cell.outputPrice) }}</span>
          </div>
          <div v-if="cell.cacheReadPrice != null" class="mc-tt-row">
            <span class="mc-tt-key">{{ t('admin.pricingDesk.ttCacheRead') }}</span>
            <span class="q-money mc-tt-val">{{ fmtPrice(cell.cacheReadPrice) }}</span>
          </div>
          <div v-if="cell.cacheWritePrice != null" class="mc-tt-row">
            <span class="mc-tt-key">{{ t('admin.pricingDesk.ttCacheWrite') }}</span>
            <span class="q-money mc-tt-val">{{ fmtPrice(cell.cacheWritePrice) }}</span>
          </div>
        </div>
        <!-- 官方价对比 -->
        <template v-if="resolvedOfficial && resolvedOfficial.found">
          <div class="mc-tt-sep"></div>
          <div class="mc-tt-section">
            <div class="mc-tt-sub">{{ t('admin.pricingDesk.ttOfficialRef') }}</div>
            <div class="mc-tt-row">
              <span class="mc-tt-key">{{ t('admin.pricingDesk.ttInput') }}</span>
              <span class="q-money mc-tt-val mc-tt-muted">{{ fmtPrice(resolvedOfficial.inputPrice ?? null) }}</span>
            </div>
            <div class="mc-tt-row">
              <span class="mc-tt-key">{{ t('admin.pricingDesk.ttOutput') }}</span>
              <span class="q-money mc-tt-val mc-tt-muted">{{ fmtPrice(resolvedOfficial.outputPrice ?? null) }}</span>
            </div>
          </div>
        </template>
        <div v-else-if="officialPricing === 'loading'" class="mc-tt-loading">{{ t('admin.pricingDesk.ttLoadingOfficial') }}</div>
        <!-- 渠道来源 -->
        <div class="mc-tt-sep"></div>
        <div class="mc-tt-channel">{{ cell.channelName }}</div>
      </div>
    </Teleport>

    <!-- 分档浮层 -->
    <Teleport to="body">
      <div
        v-if="showIntervals"
        v-click-outside="() => { showIntervals = false }"
        class="mc-intervals"
        :style="{ top: popoverTop + 'px', left: popoverLeft + 'px' }"
      >
        <div class="mc-iv-title">{{ t('admin.pricingDesk.tieredPricingTitle') }}</div>
        <table class="mc-iv-table">
          <thead>
            <tr>
              <th class="mc-iv-th">{{ t('admin.pricingDesk.tieredColTier') }}</th>
              <th class="mc-iv-th mc-iv-r">input</th>
              <th class="mc-iv-th mc-iv-r">output</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="iv in cell.intervals" :key="iv.sort_order ?? iv.min_tokens" class="mc-iv-tr">
              <td class="mc-iv-td">{{ iv.tier_label || fmtRange(iv.min_tokens, iv.max_tokens) }}</td>
              <td class="mc-iv-td mc-iv-r q-money">{{ fmtPrice(iv.input_price) }}</td>
              <td class="mc-iv-td mc-iv-r q-money">{{ fmtPrice(iv.output_price) }}</td>
            </tr>
          </tbody>
        </table>
        <div class="mc-iv-foot">{{ t('admin.pricingDesk.tieredFromChannel') }}{{ cell.channelName }}</div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MatrixCell } from './usePricingMatrix'
import type { OfficialPricing } from './usePricingMatrix'

const props = defineProps<{
  cell: MatrixCell
  model: string
  officialPricing?: OfficialPricing | 'loading'
}>()

const { t } = useI18n()
const showIntervals = ref(false)
const popoverTop = ref(0)
const popoverLeft = ref(0)
const showTooltip = ref(false)
const tooltipTop = ref(0)
const tooltipLeft = ref(0)
const tooltipReady = ref(false)
let tooltipTimer: ReturnType<typeof setTimeout> | null = null

const vClickOutside = {
  mounted(el: HTMLElement, binding: { value: () => void }) {
    ;(el as any)._clickOutside = (e: MouseEvent) => {
      if (!el.contains(e.target as Node)) binding.value()
    }
    document.addEventListener('click', (el as any)._clickOutside)
  },
  unmounted(el: HTMLElement) {
    document.removeEventListener('click', (el as any)._clickOutside)
  }
}

function openIntervals(e: MouseEvent) {
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
  popoverTop.value = rect.bottom + 4
  popoverLeft.value = rect.left
  showIntervals.value = !showIntervals.value
}

function onEnter(e: MouseEvent) {
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
  // Position tooltip: try right of cell, clamp to viewport
  const ttW = 220
  let left = rect.right + 8
  if (left + ttW > window.innerWidth - 12) left = rect.left - ttW - 8
  tooltipLeft.value = Math.max(8, left)
  tooltipTop.value = rect.top
  showTooltip.value = true
  tooltipTimer = setTimeout(() => { tooltipReady.value = true }, 120)
}

function onLeave() {
  if (tooltipTimer) { clearTimeout(tooltipTimer); tooltipTimer = null }
  showTooltip.value = false
  tooltipReady.value = false
}

/**
 * 智能价格截断：≥1 两位，<1 按量级三~四位
 * 单位：per token → per 1M token
 */
function fmtPrice(v: number | null | undefined): string {
  if (v == null) return '—'
  const perM = v * 1_000_000
  let decimals: number
  if (perM >= 1) decimals = 2
  else if (perM >= 0.1) decimals = 3
  else decimals = 4
  return `$${perM.toFixed(decimals)}/M`
}

function fmtRange(min: number, max: number | null): string {
  const fmt = (n: number) =>
    n >= 1_000_000 ? `${(n / 1_000_000).toFixed(1)}M` : n >= 1000 ? `${(n / 1000).toFixed(0)}k` : `${n}`
  return max == null ? `>${fmt(min)}` : `${fmt(min)}~${fmt(max)}`
}

const resolvedOfficial = computed<OfficialPricing | null>(() => {
  const op = props.officialPricing
  if (!op || op === 'loading' || !op.found) return null
  return op
})

const discountIn = computed<number | null>(() => {
  const op = resolvedOfficial.value
  if (!op) return null
  if (op.inputPrice == null || props.cell.inputPrice == null || op.inputPrice === 0) return null
  return props.cell.inputPrice / op.inputPrice
})

const discountOut = computed<number | null>(() => {
  const op = resolvedOfficial.value
  if (!op) return null
  if (op.outputPrice == null || props.cell.outputPrice == null || op.outputPrice === 0) return null
  return props.cell.outputPrice / op.outputPrice
})

function discountStyle(ratio: number): Record<string, string> {
  if (ratio < 0.999) return { background: 'var(--ok-dim)', color: 'var(--ok)' }
  if (ratio > 1.001) return { background: 'var(--bad-dim)', color: 'var(--bad)' }
  return { background: 'var(--bg-2)', color: 'var(--ink-2)' }
}

function priceColorStyle(ratio: number | null): Record<string, string> {
  if (ratio === null) return {}
  if (ratio < 0.999) return { color: 'var(--ok)' }
  if (ratio > 1.001) return { color: 'var(--bad)' }
  return {}
}

function discountLabel(ratio: number): string {
  if (ratio < 0.999) return `-${((1 - ratio) * 100).toFixed(0)}%`
  if (ratio > 1.001) return `+${((ratio - 1) * 100).toFixed(0)}%`
  return t('admin.pricingDesk.officialPrice')
}
</script>

<style scoped>
.mc-root {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding: 4px 6px;
  border-radius: 8px;
  cursor: default;
  transition: background .15s, box-shadow .15s;
}
.mc-root:hover {
  background: var(--azure-dim, rgba(92,168,255,.1));
  box-shadow: 0 0 0 1px rgba(92,168,255,.22);
}
@media (prefers-reduced-motion: reduce) {
  .mc-root { transition: none; }
}

/* 固定三列子栅格：标签 | 价格(右对齐) | 徽章(定宽)
   —— IN/OUT 共享同一 template，价格小数与徽章竖直对齐；整块居中，落在居中列头下方 */
.mc-row {
  display: grid;
  grid-template-columns: 18px minmax(58px, auto) 40px;
  align-items: center;
  gap: 4px;
  width: fit-content;
  margin: 0 auto;
}
.mc-label {
  font-size: 9px;
  font-weight: 600;
  letter-spacing: .06em;
  text-transform: uppercase;
  color: var(--ink-2);
  text-align: left;
}
.mc-price { font-size: 11px; text-align: right; white-space: nowrap; }
.mc-badge {
  font-size: 9.5px;
  font-weight: 600;
  border-radius: 4px;
  padding: 0 3px;
  line-height: 1.5;
  text-align: center;
  justify-self: start;
}

.mc-tier-badge {
  position: absolute;
  right: 2px;
  top: 2px;
  cursor: pointer;
  font-size: 10px;
}

/* ── Tooltip ── */
.mc-tooltip {
  position: fixed;
  z-index: 9999;
  min-width: 200px;
  max-width: 240px;
  padding: 10px 12px;
  background: var(--metal, linear-gradient(180deg,#15181E,#0E1014));
  border: 1px solid var(--line-1);
  border-radius: var(--q-radius, 12px);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.06), 0 12px 32px rgba(0,0,0,.5);
  pointer-events: none;
}
.mc-tt-title {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 10.5px;
  color: var(--azure);
  font-weight: 600;
  margin-bottom: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.mc-tt-sep {
  height: 1px;
  background: var(--line-0);
  margin: 6px 0;
}
.mc-tt-section { display: flex; flex-direction: column; gap: 4px; }
.mc-tt-sub {
  font-size: 9.5px;
  font-weight: 600;
  letter-spacing: .07em;
  text-transform: uppercase;
  color: var(--ink-2);
  margin-bottom: 2px;
}
.mc-tt-row { display: flex; justify-content: space-between; align-items: center; gap: 8px; }
.mc-tt-key { font-size: 10.5px; color: var(--ink-1); }
.mc-tt-val { font-size: 11px; }
.mc-tt-muted { color: var(--ink-2) !important; text-shadow: none !important; }
.mc-tt-loading { font-size: 10px; color: var(--ink-2); text-align: center; padding: 4px 0; }
.mc-tt-channel { font-size: 9.5px; color: var(--ink-2); text-align: right; }

/* ── 分档浮层 ── */
.mc-intervals {
  position: fixed;
  z-index: 9998;
  min-width: 200px;
  padding: 10px 12px;
  background: var(--bg-1);
  border: 1px solid var(--line-1);
  border-radius: var(--q-radius, 12px);
  box-shadow: 0 12px 32px rgba(0,0,0,.4);
}
.mc-iv-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--ink-0);
  margin-bottom: 8px;
}
.mc-iv-table { width: 100%; border-collapse: collapse; }
.mc-iv-th {
  font-size: 9.5px;
  font-weight: 600;
  letter-spacing: .06em;
  text-transform: uppercase;
  color: var(--ink-2);
  padding-bottom: 4px;
  text-align: left;
}
.mc-iv-r { text-align: right; }
.mc-iv-tr { border-top: 1px solid var(--line-0); }
.mc-iv-td { padding: 3px 0; font-size: 11px; color: var(--ink-1); }
.mc-iv-foot { margin-top: 6px; font-size: 9.5px; color: var(--ink-2); text-align: right; }
</style>
