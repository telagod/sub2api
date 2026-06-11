<template>
  <!-- 抽屉遮罩 -->
  <Transition name="fade">
    <div
      v-if="visible"
      class="fixed inset-0 z-40 bg-black/40"
      @click="$emit('close')"
    />
  </Transition>

  <!-- 抽屉面板 -->
  <Transition name="slide-right">
    <div
      v-if="visible"
      class="fixed right-0 top-0 z-50 h-full w-full max-w-md overflow-y-auto shadow-2xl"
      :style="{ background: 'var(--bg-1)', borderLeft: '1px solid var(--line-1)' }"
    >
      <!-- 头部 -->
      <div
        class="flex items-center justify-between px-6 py-4"
        :style="{ borderBottom: '1px solid var(--line-0)' }"
      >
        <div class="flex items-center gap-2">
          <CalculatorIcon class="h-5 w-5" :style="{ color: 'var(--azure)' }" />
          <h2 class="text-base font-semibold" :style="{ color: 'var(--ink-0)' }">{{ t('admin.pricingDesk.simTitle') }}</h2>
        </div>
        <button
          class="rounded-lg p-1.5 transition-colors"
          :style="{ color: 'var(--ink-2)' }"
          @click="$emit('close')"
        >
          <XIcon class="h-5 w-5" />
        </button>
      </div>

      <!-- 表单区 -->
      <div class="space-y-5 px-6 py-5">
        <!-- 选模型 -->
        <div>
          <label class="mb-1.5 block text-sm font-medium" :style="{ color: 'var(--ink-1)' }">{{ t('admin.pricingDesk.simModelLabel') }}</label>
          <select
            v-model="selectedModel"
            class="w-full rounded-lg px-3 py-2 text-sm"
            :style="{
              background: 'var(--bg-0)',
              border: '1px solid var(--line-1)',
              color: 'var(--ink-0)'
            }"
          >
            <option value="">{{ t('admin.pricingDesk.simModelPlaceholder') }}</option>
            <optgroup v-for="platform in platforms" :key="platform" :label="platform">
              <option
                v-for="model in modelsByPlatform[platform]"
                :key="model"
                :value="model"
              >
                {{ model }}
              </option>
            </optgroup>
          </select>
        </div>

        <!-- 选分组 -->
        <div>
          <label class="mb-1.5 block text-sm font-medium" :style="{ color: 'var(--ink-1)' }">{{ t('admin.pricingDesk.simGroupLabel') }}</label>
          <select
            v-model="selectedGroupId"
            class="w-full rounded-lg px-3 py-2 text-sm"
            :style="{
              background: 'var(--bg-0)',
              border: '1px solid var(--line-1)',
              color: 'var(--ink-0)'
            }"
          >
            <option :value="null">{{ t('admin.pricingDesk.simGroupPlaceholder') }}</option>
            <option v-for="g in activeGroups" :key="g.id" :value="g.id">
              {{ g.name }} (×{{ g.rate_multiplier.toFixed(2) }})
            </option>
          </select>
        </div>

        <!-- Token 量 -->
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="mb-1.5 block text-sm font-medium" :style="{ color: 'var(--ink-1)' }">Input tokens</label>
            <input
              v-model.number="inputTokens"
              type="number"
              min="0"
              class="w-full rounded-lg px-3 py-2 text-sm"
              :style="{
                background: 'var(--bg-0)',
                border: '1px solid var(--line-1)',
                color: 'var(--ink-0)'
              }"
              placeholder="1000000"
            />
          </div>
          <div>
            <label class="mb-1.5 block text-sm font-medium" :style="{ color: 'var(--ink-1)' }">Output tokens</label>
            <input
              v-model.number="outputTokens"
              type="number"
              min="0"
              class="w-full rounded-lg px-3 py-2 text-sm"
              :style="{
                background: 'var(--bg-0)',
                border: '1px solid var(--line-1)',
                color: 'var(--ink-0)'
              }"
              placeholder="200000"
            />
          </div>
        </div>

        <!-- Cache 命中滑杆 -->
        <div>
          <label class="mb-1.5 flex items-center justify-between text-sm font-medium" :style="{ color: 'var(--ink-1)' }">
            <span>{{ t('admin.pricingDesk.simCacheHit') }}</span>
            <span :style="{ color: 'var(--azure)' }">{{ (cacheHitRatio * 100).toFixed(0) }}%</span>
          </label>
          <input
            v-model.number="cacheHitRatio"
            type="range"
            min="0"
            max="1"
            step="0.01"
            class="w-full accent-[var(--azure)]"
          />
          <div class="mt-1 flex justify-between text-xs" :style="{ color: 'var(--ink-2)' }">
            <span>0%</span>
            <span>100%</span>
          </div>
        </div>

        <!-- 结果展示 -->
        <div
          v-if="selectedModel && selectedGroupId !== null && cell"
          class="rounded-xl p-4 space-y-3"
          :style="{ background: 'var(--bg-2)', border: '1px solid var(--line-0)' }"
        >
          <div class="text-xs font-semibold uppercase tracking-wider mb-3" :style="{ color: 'var(--ink-2)' }">{{ t('admin.pricingDesk.simResultTitle') }}</div>
          <div class="space-y-2">
            <SimResultRow :label="t('admin.pricingDesk.simInputCost')" :value="inputCost" />
            <SimResultRow :label="t('admin.pricingDesk.simOutputCost')" :value="outputCost" />
            <SimResultRow v-if="cacheHitRatio > 0" :label="t('admin.pricingDesk.simCacheCost')" :value="cacheCost" />
            <div class="border-t pt-2" :style="{ borderColor: 'var(--line-1)' }">
              <SimResultRow :label="t('admin.pricingDesk.simTotal')" :value="totalCost" :large="true" />
            </div>
          </div>

          <!-- 对比官方价 — Major fix: replace hardcoded Tailwind color classes with design-system tokens -->
          <div
            v-if="officialTotal !== null"
            class="mt-3 rounded-lg p-3"
            :style="totalCost <= officialTotal
              ? { background: 'var(--ok-dim)' }
              : { background: 'var(--bad-dim)' }"
          >
            <div class="text-xs" :style="{ color: 'var(--ink-1)' }">{{ t('admin.pricingDesk.simOfficialTotal') }}<span class="q-money">{{ fmtUSD(officialTotal) }}</span></div>
            <div
              class="mt-1 text-sm font-medium"
              :style="totalCost <= officialTotal ? { color: 'var(--ok)' } : { color: 'var(--bad)' }"
            >
              {{ totalCost <= officialTotal ? t('admin.pricingDesk.simCheaper', { diff: fmtUSD(officialTotal - totalCost), pct: ((1 - totalCost / officialTotal) * 100).toFixed(1) }) : t('admin.pricingDesk.simDearer', { diff: fmtUSD(totalCost - officialTotal), pct: ((totalCost / officialTotal - 1) * 100).toFixed(1) }) }}
            </div>
          </div>
          <p v-else class="text-xs" :style="{ color: 'var(--ink-2)' }">{{ t('admin.pricingDesk.simNoOfficialRef') }}</p>
        </div>

        <!-- 未选模型/分组提示 -->
        <div
          v-else
          class="rounded-xl p-6 text-center"
          :style="{ background: 'var(--bg-2)', color: 'var(--ink-2)' }"
        >
          <CalculatorIcon class="mx-auto mb-2 h-8 w-8 opacity-40" />
          <p class="text-sm">{{ t('admin.pricingDesk.simSelectHint') }}</p>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { CalculatorIcon, XIcon } from 'lucide-vue-next'
import SimResultRow from './SimResultRow.vue'
import type { MatrixRow, OfficialPricing } from './usePricingMatrix'
import type { AdminGroup } from '@/types'

const props = defineProps<{
  visible: boolean
  platforms: string[]
  matrix: MatrixRow[]
  activeGroups: AdminGroup[]
  officialPricingCache: Record<string, OfficialPricing | 'loading'>
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'need-official-pricing', model: string): void
}>()

const { t } = useI18n()
const selectedModel = ref('')
const selectedGroupId = ref<number | null>(null)
const inputTokens = ref(1_000_000)
const outputTokens = ref(200_000)
const cacheHitRatio = ref(0)

const modelsByPlatform = computed(() => {
  const map: Record<string, string[]> = {}
  for (const row of props.matrix) {
    if (!map[row.platform]) map[row.platform] = []
    map[row.platform].push(row.model)
  }
  return map
})

const cell = computed(() => {
  if (!selectedModel.value || selectedGroupId.value === null) return null
  const row = props.matrix.find(r => r.model === selectedModel.value)
  return row?.cells[selectedGroupId.value] ?? null
})

// 当选中模型变化时触发官方价加载
watch(selectedModel, (model) => {
  if (model) emit('need-official-pricing', model)
})

const inputCost = computed(() => {
  // Minor fix: inputCost only covers non-cached tokens. Cache cost is broken
  // out separately so the Cache 读取费用 row does not double-count.
  const c = cell.value
  if (!c || c.inputPrice == null) return 0
  const nonCache = inputTokens.value * (1 - cacheHitRatio.value)
  return nonCache * c.inputPrice
})

const outputCost = computed(() => {
  const c = cell.value
  if (!c || c.outputPrice == null) return 0
  return outputTokens.value * c.outputPrice
})

const cacheCost = computed(() => {
  const c = cell.value
  if (!c) return 0
  return inputTokens.value * cacheHitRatio.value * (c.cacheReadPrice ?? c.inputPrice ?? 0)
})

// Minor fix: totalCost now includes cacheCost as an independent addend.
const totalCost = computed(() => inputCost.value + outputCost.value + cacheCost.value)

const officialPricing = computed<OfficialPricing | null>(() => {
  if (!selectedModel.value) return null
  const op = props.officialPricingCache[selectedModel.value]
  if (!op || op === 'loading' || !op.found) return null
  return op
})

const officialTotal = computed<number | null>(() => {
  const op = officialPricing.value
  if (!op) return null
  const inputP = op.inputPrice ?? 0
  const outputP = op.outputPrice ?? 0
  const cacheP = op.cacheReadPrice ?? inputP * 0.1
  const cacheRead = inputTokens.value * cacheHitRatio.value
  const nonCache = inputTokens.value - cacheRead
  return nonCache * inputP + cacheRead * cacheP + outputTokens.value * outputP
})

function fmtUSD(v: number): string {
  return `$${v.toFixed(6)}`
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-right-enter-active,
.slide-right-leave-active {
  transition: transform 0.25s ease;
}
.slide-right-enter-from,
.slide-right-leave-to {
  transform: translateX(100%);
}
</style>
