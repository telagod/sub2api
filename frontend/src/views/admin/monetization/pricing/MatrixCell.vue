<template>
  <div class="relative flex flex-col gap-0.5">
    <!-- input 价 -->
    <div class="flex items-center justify-end gap-1">
      <span
        class="q-money text-xs"
        :style="{ fontSize: '11px' }"
      >
        {{ fmtPrice(cell.inputPrice) }}
      </span>
      <span
        v-if="discountIn !== null"
        class="text-[10px] font-medium rounded px-0.5"
        :style="discountStyle(discountIn)"
      >
        {{ discountLabel(discountIn) }}
      </span>
    </div>

    <!-- output 价 -->
    <div class="flex items-center justify-end gap-1">
      <span class="q-money text-xs" :style="{ fontSize: '11px' }">
        {{ fmtPrice(cell.outputPrice) }}
      </span>
      <span
        v-if="discountOut !== null"
        class="text-[10px] font-medium rounded px-0.5"
        :style="discountStyle(discountOut)"
      >
        {{ discountLabel(discountOut) }}
      </span>
    </div>

    <!-- 分档角标 -->
    <div
      v-if="cell.hasIntervals"
      class="absolute -right-1 -top-1 cursor-pointer"
      @click.stop="(e) => openIntervals(e)"
    >
      <span class="text-[10px]" :style="{ color: 'var(--azure)' }">▾</span>
    </div>

    <!-- 分档浮层 -->
    <Teleport to="body">
      <div
        v-if="showIntervals"
        v-click-outside="() => { showIntervals = false }"
        class="fixed z-50 rounded-xl shadow-xl p-3 text-xs min-w-[200px]"
        :style="{
          background: 'var(--bg-1)',
          border: '1px solid var(--line-1)',
          top: popoverTop + 'px',
          left: popoverLeft + 'px'
        }"
      >
        <div class="mb-2 font-semibold" :style="{ color: 'var(--ink-0)' }">分档定价</div>
        <table class="w-full">
          <thead>
            <tr :style="{ color: 'var(--ink-2)' }">
              <th class="text-left pr-3 pb-1">档位</th>
              <th class="text-right pr-2 pb-1">input</th>
              <th class="text-right pb-1">output</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="iv in cell.intervals" :key="iv.sort_order ?? iv.min_tokens" class="border-t" :style="{ borderColor: 'var(--line-0)' }">
              <td class="pr-3 py-0.5" :style="{ color: 'var(--ink-1)' }">
                {{ iv.tier_label || fmtRange(iv.min_tokens, iv.max_tokens) }}
              </td>
              <td class="text-right pr-2 py-0.5 q-money" :style="{ fontSize: '11px' }">{{ fmtPrice(iv.input_price) }}</td>
              <td class="text-right py-0.5 q-money" :style="{ fontSize: '11px' }">{{ fmtPrice(iv.output_price) }}</td>
            </tr>
          </tbody>
        </table>
        <div class="mt-2 text-right">
          <span class="text-[10px]" :style="{ color: 'var(--ink-2)' }">来自渠道：{{ cell.channelName }}</span>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { MatrixCell } from './usePricingMatrix'
import type { OfficialPricing } from './usePricingMatrix'

const props = defineProps<{
  cell: MatrixCell
  model: string
  officialPricing?: OfficialPricing | 'loading'
}>()

const showIntervals = ref(false)
const popoverTop = ref(0)
const popoverLeft = ref(0)

// Blocker fix: define v-click-outside locally (mirrors AdminOrderTable.vue pattern)
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

// Minor fix: compute popover position from click target rect
function openIntervals(e: MouseEvent) {
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
  popoverTop.value = rect.bottom + 4
  popoverLeft.value = rect.left
  showIntervals.value = !showIntervals.value
}

function fmtPrice(v: number | null | undefined): string {
  if (v == null) return '—'
  // 价格单位：per token，转换为 per 1M token 显示
  return `$${(v * 1_000_000).toFixed(4)}/M`
}

function fmtRange(min: number, max: number | null): string {
  const fmt = (n: number) => n >= 1_000_000 ? `${(n / 1_000_000).toFixed(1)}M` : n >= 1000 ? `${(n / 1000).toFixed(0)}k` : `${n}`
  return max == null ? `>${fmt(min)}` : `${fmt(min)}~${fmt(max)}`
}

// 折扣率计算 (实际价 / 官方价)，比较 input 和 output
const discountIn = computed<number | null>(() => {
  const op = props.officialPricing
  if (!op || op === 'loading' || !op.found) return null
  if (op.inputPrice == null || props.cell.inputPrice == null || op.inputPrice === 0) return null
  return props.cell.inputPrice / op.inputPrice
})

const discountOut = computed<number | null>(() => {
  const op = props.officialPricing
  if (!op || op === 'loading' || !op.found) return null
  if (op.outputPrice == null || props.cell.outputPrice == null || op.outputPrice === 0) return null
  return props.cell.outputPrice / op.outputPrice
})

// Major fix: replace hardcoded Tailwind color classes with design-system tokens
function discountStyle(ratio: number): Record<string, string> {
  if (ratio < 1) return { background: 'var(--ok-dim)', color: 'var(--ok)' }
  if (ratio > 1) return { background: 'var(--bad-dim)', color: 'var(--bad)' }
  return { background: 'var(--bg-2)', color: 'var(--ink-2)' }
}

function discountLabel(ratio: number): string {
  if (ratio < 1) return `-${((1 - ratio) * 100).toFixed(0)}%`
  if (ratio > 1) return `+${((ratio - 1) * 100).toFixed(0)}%`
  return '=官价'
}
</script>
