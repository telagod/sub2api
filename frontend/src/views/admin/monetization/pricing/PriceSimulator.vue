<template>
  <!-- 抽屉遮罩 -->
  <Transition name="ps-fade">
    <div
      v-if="visible"
      class="ps-overlay"
      @click="$emit('close')"
    />
  </Transition>

  <!-- 抽屉面板 -->
  <Transition name="ps-slide">
    <div
      v-if="visible"
      class="ps-drawer"
    >
      <!-- 头部 -->
      <div class="ps-head">
        <div class="ps-head-left">
          <CalculatorIcon class="ps-head-ico" />
          <h2 class="ps-head-title">{{ t('admin.pricingDesk.simTitle') }}</h2>
        </div>
        <button
          class="ps-close q-focus-glow"
          :aria-label="t('admin.pricingDesk.simClose')"
          @click="$emit('close')"
        >
          <XIcon class="ps-close-ico" />
        </button>
      </div>

      <!-- 表单区 -->
      <div class="ps-body">
        <!-- 选模型 -->
        <div class="ps-field">
          <label class="ps-label">{{ t('admin.pricingDesk.simModelLabel') }}</label>
          <select
            v-model="selectedModel"
            class="ps-select q-focus-glow"
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
        <div class="ps-field">
          <label class="ps-label">{{ t('admin.pricingDesk.simGroupLabel') }}</label>
          <select
            v-model="selectedGroupId"
            class="ps-select q-focus-glow"
          >
            <option :value="null">{{ t('admin.pricingDesk.simGroupPlaceholder') }}</option>
            <option v-for="g in activeGroups" :key="g.id" :value="g.id">
              {{ g.name }} (×{{ g.rate_multiplier.toFixed(2) }})
            </option>
          </select>
        </div>

        <!-- Token 量 -->
        <div class="ps-token-row">
          <div class="ps-field">
            <label class="ps-label">{{ t('admin.pricingDesk.simInputTokens') }}</label>
            <input
              v-model.number="inputTokens"
              type="number"
              min="0"
              class="ps-input q-focus-glow"
              placeholder="1000000"
            />
          </div>
          <div class="ps-field">
            <label class="ps-label">{{ t('admin.pricingDesk.simOutputTokens') }}</label>
            <input
              v-model.number="outputTokens"
              type="number"
              min="0"
              class="ps-input q-focus-glow"
              placeholder="200000"
            />
          </div>
        </div>

        <!-- Cache 命中滑杆 -->
        <div class="ps-field">
          <label class="ps-label ps-label-row">
            <span>{{ t('admin.pricingDesk.simCacheHit') }}</span>
            <span class="ps-cache-pct">{{ (cacheHitRatio * 100).toFixed(0) }}%</span>
          </label>
          <input
            v-model.number="cacheHitRatio"
            type="range"
            min="0"
            max="1"
            step="0.01"
            class="ps-range"
          />
          <div class="ps-range-hint">
            <span>0%</span>
            <span>100%</span>
          </div>
        </div>

        <!-- 结果展示 -->
        <div
          v-if="selectedModel && selectedGroupId !== null && cell"
          class="ps-result"
        >
          <div class="ps-result-label">{{ t('admin.pricingDesk.simResultTitle') }}</div>
          <div class="ps-result-rows">
            <SimResultRow :label="t('admin.pricingDesk.simInputCost')" :value="inputCost" />
            <SimResultRow :label="t('admin.pricingDesk.simOutputCost')" :value="outputCost" />
            <SimResultRow v-if="cacheHitRatio > 0" :label="t('admin.pricingDesk.simCacheCost')" :value="cacheCost" />
            <div class="ps-result-divider">
              <SimResultRow :label="t('admin.pricingDesk.simTotal')" :value="totalCost" :large="true" />
            </div>
          </div>

          <!-- 对比官方价 -->
          <div
            v-if="officialTotal !== null"
            class="ps-cmp"
            :class="totalCost <= officialTotal ? 'ps-cmp-ok' : 'ps-cmp-bad'"
          >
            <div class="ps-cmp-ref">{{ t('admin.pricingDesk.simOfficialTotal') }}<span class="q-money">{{ fmtUSD(officialTotal) }}</span></div>
            <div
              class="ps-cmp-verdict"
              :class="totalCost <= officialTotal ? 'ps-ok' : 'ps-bad'"
            >
              {{ totalCost <= officialTotal
                ? t('admin.pricingDesk.simCheaper', { diff: fmtUSD(officialTotal - totalCost), pct: ((1 - totalCost / officialTotal) * 100).toFixed(1) })
                : t('admin.pricingDesk.simDearer', { diff: fmtUSD(totalCost - officialTotal), pct: ((totalCost / officialTotal - 1) * 100).toFixed(1) }) }}
            </div>
          </div>
          <p v-else class="ps-no-ref">{{ t('admin.pricingDesk.simNoOfficialRef') }}</p>
        </div>

        <!-- 未选模型/分组提示 -->
        <div
          v-else
          class="ps-placeholder"
        >
          <CalculatorIcon class="ps-placeholder-ico" />
          <p class="ps-placeholder-txt">{{ t('admin.pricingDesk.simSelectHint') }}</p>
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
/* ── 遮罩 ── */
.ps-overlay {
  position: fixed; inset: 0; z-index: 40;
  background: rgba(0, 0, 0, .45);
}

/* ── 抽屉面板 ── */
.ps-drawer {
  position: fixed; right: 0; top: 0; z-index: 50;
  width: 100%; max-width: 420px; height: 100%;
  overflow-y: auto;
  background: var(--metal, linear-gradient(180deg,#15181E,#0E1014));
  border-left: 1px solid var(--line-1);
  box-shadow: var(--edge-hi, inset 0 1px 0 rgba(255,255,255,.04)), -8px 0 32px rgba(0,0,0,.4);
}

/* ── 头部 ── */
.ps-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 16px 20px; border-bottom: 1px solid var(--line-0);
  background: var(--bg-2);
}
.ps-head-left { display: flex; align-items: center; gap: 8px; }
.ps-head-ico { width: 18px; height: 18px; color: var(--azure); flex-shrink: 0; }
.ps-head-title { font-size: 14px; font-weight: 600; color: var(--ink-0); margin: 0; }

.ps-close {
  display: inline-flex; align-items: center; justify-content: center;
  padding: 5px; border-radius: 8px; border: 1px solid transparent;
  background: transparent; color: var(--ink-2);
  cursor: pointer; transition: border-color .15s, color .15s, background .15s;
}
.ps-close:hover { border-color: var(--line-1); color: var(--ink-0); background: var(--bg-2); }
.ps-close-ico { width: 16px; height: 16px; }

/* ── 表单区 ── */
.ps-body { display: flex; flex-direction: column; gap: 18px; padding: 20px; }

.ps-field { display: flex; flex-direction: column; gap: 5px; }
.ps-label {
  font-size: 11.5px; font-weight: 600; color: var(--ink-1);
}
.ps-label-row { display: flex; align-items: center; justify-content: space-between; }
.ps-cache-pct { color: var(--azure); font-family: var(--font-mono, monospace); font-variant-numeric: tabular-nums; }

.ps-select,
.ps-input {
  width: 100%; padding: 7px 10px; border-radius: 8px;
  font-size: 12.5px; font-family: inherit;
  background: var(--bg-0); border: 1px solid var(--line-1);
  color: var(--ink-0);
  transition: border-color .15s;
  -webkit-appearance: none; appearance: none;
}
.ps-select:hover, .ps-input:hover { border-color: var(--line-0); }

.ps-token-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }

.ps-range {
  width: 100%; accent-color: var(--azure);
  cursor: pointer;
}
.ps-range-hint {
  display: flex; justify-content: space-between;
  font-size: 10.5px; color: var(--ink-2); margin-top: 3px;
}

/* ── 结果卡 ── */
.ps-result {
  display: flex; flex-direction: column; gap: 10px;
  padding: 14px;
  background: var(--bg-2); border: 1px solid var(--line-0);
  border-radius: var(--q-radius, 12px);
}
.ps-result-label {
  font-size: 9.5px; font-weight: 600; letter-spacing: .08em;
  text-transform: uppercase; color: var(--ink-2);
}
.ps-result-rows { display: flex; flex-direction: column; gap: 7px; }
.ps-result-divider {
  border-top: 1px solid var(--line-1); padding-top: 8px; margin-top: 2px;
}

/* 对比官方价 */
.ps-cmp {
  border-radius: 8px; padding: 10px 12px;
  display: flex; flex-direction: column; gap: 4px;
}
.ps-cmp-ok  { background: var(--ok-dim); }
.ps-cmp-bad { background: var(--bad-dim); }
.ps-cmp-ref { font-size: 11px; color: var(--ink-1); }
.ps-cmp-verdict { font-size: 12px; font-weight: 600; }
.ps-ok  { color: var(--ok); }
.ps-bad { color: var(--bad); }
.ps-no-ref { font-size: 11px; color: var(--ink-2); margin: 0; }

/* ── 空态 ── */
.ps-placeholder {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 8px; padding: 40px 16px;
  background: var(--bg-2); border: 1px dashed var(--line-1);
  border-radius: var(--q-radius, 12px); text-align: center;
}
.ps-placeholder-ico { width: 32px; height: 32px; color: var(--ink-2); opacity: .35; }
.ps-placeholder-txt { font-size: 12.5px; color: var(--ink-2); margin: 0; }

/* ── 入/出场动画 ── */
.ps-fade-enter-active,
.ps-fade-leave-active { transition: opacity .2s; }
.ps-fade-enter-from,
.ps-fade-leave-to { opacity: 0; }

.ps-slide-enter-active,
.ps-slide-leave-active { transition: transform .25s cubic-bezier(.22,.68,0,1.2); }
.ps-slide-enter-from,
.ps-slide-leave-to { transform: translateX(100%); }

@media (prefers-reduced-motion: reduce) {
  .ps-fade-enter-active,
  .ps-fade-leave-active,
  .ps-slide-enter-active,
  .ps-slide-leave-active { transition: none; }
}
</style>
