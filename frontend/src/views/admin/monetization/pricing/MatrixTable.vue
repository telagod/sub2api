<template>
  <!-- 空态 -->
  <div
    v-if="!loading && platforms.length === 0"
    class="mt-empty rise"
  >
    <div class="mt-empty-ico">⬡</div>
    <p class="mt-empty-title">{{ t('admin.pricingDesk.noData') }}</p>
    <p class="mt-empty-hint">{{ t('admin.pricingDesk.noDataHint') }}</p>
    <RouterLink
      to="/admin/channels/pricing"
      class="mt-empty-btn"
    >
      <CalculatorIcon class="mt-empty-btn-ico" />
      {{ t('admin.pricingDesk.goConfigBtn') }}
    </RouterLink>
  </div>

  <!-- 矩阵面板 -->
  <div v-else class="mt-panel rise">
    <!-- 图例 + 同步目录按钮 -->
    <div class="mt-legend">
      <span class="mt-legend-item">
        <span class="mt-legend-dot mt-ok"></span>
        {{ t('admin.pricingDesk.legendBelow') }}
      </span>
      <span class="mt-legend-sep">·</span>
      <span class="mt-legend-item">
        <span class="mt-legend-dot mt-neutral"></span>
        {{ t('admin.pricingDesk.legendEqual') }}
      </span>
      <span class="mt-legend-sep">·</span>
      <span class="mt-legend-item">
        <span class="mt-legend-dot mt-bad"></span>
        {{ t('admin.pricingDesk.legendAbove') }}
      </span>
      <span class="mt-legend-spacer"></span>
      <button
        class="mt-sync-btn"
        :disabled="syncLoading"
        :title="t('admin.pricingDesk.syncCatalogTitle')"
        @click="$emit('sync-catalog')"
      >
        <RefreshCwIcon class="mt-sync-ico" :class="syncLoading ? 'mt-spinning' : ''" />
        {{ t('admin.pricingDesk.syncCatalogBtn') }}
      </button>
    </div>

    <div class="mt-scroll-wrap">
      <table class="mt-table">
        <!-- 表头 -->
        <thead>
          <tr class="mt-head-row">
            <th class="mt-th mt-th-model mt-sticky">
              {{ t('admin.pricingDesk.colModel') }}
            </th>
            <th
              v-for="group in activeGroups"
              :key="group.id"
              class="mt-th mt-th-group"
            >
              <div class="mt-th-inner">
                <span class="mt-group-name">{{ group.name }}</span>
                <!-- ×倍率就地编辑 -->
                <div class="mt-multiplier-wrap">
                  <span
                    v-if="editingGroupId !== group.id"
                    class="mt-multiplier"
                    :title="t('admin.pricingDesk.dblClickToEdit')"
                    @dblclick="startEditMultiplier(group)"
                  >×{{ group.rate_multiplier.toFixed(2) }}</span>
                  <template v-else>
                    <input
                      :ref="(el) => { if (group.id === editingGroupId) { multiplierInputRef = el as HTMLInputElement } }"
                      v-model.number="editingMultiplierValue"
                      type="number"
                      step="0.01"
                      min="0"
                      class="mt-multiplier-input q-focus-glow"
                      @keydown.enter="commitMultiplier(group.id)"
                      @keydown.esc="cancelEditMultiplier"
                      @blur="commitMultiplier(group.id)"
                    />
                  </template>
                </div>
              </div>
            </th>
          </tr>
        </thead>

        <!-- 按 platform 分组折叠 -->
        <tbody>
          <template v-for="platform in platforms" :key="platform">
            <!-- platform 行组标题 -->
            <tr
              class="mt-platform-row"
              @click="togglePlatform(platform)"
            >
              <td
                :colspan="activeGroups.length + 1"
                class="mt-platform-cell"
              >
                <div class="mt-platform-inner">
                  <ChevronDownIcon
                    class="mt-chevron"
                    :class="collapsedPlatforms.has(platform) ? 'mt-collapsed' : ''"
                  />
                  <span class="mt-platform-name">{{ platform }}</span>
                  <span class="mt-platform-count">
                    {{ t('admin.pricingDesk.modelCount', { n: rowsByPlatform[platform]?.length ?? 0 }) }}
                  </span>
                </div>
              </td>
            </tr>

            <!-- 模型行 -->
            <template v-if="!collapsedPlatforms.has(platform)">
              <tr
                v-for="row in rowsByPlatform[platform]"
                :key="row.model"
                class="mt-model-row"
                @mouseenter="onRowHover(row.model)"
              >
                <!-- 模型名列 -->
                <td class="mt-td mt-td-model mt-sticky">
                  <span class="mt-model-name">{{ row.model }}</span>
                  <!-- 官方价基准条 -->
                  <template v-if="officialPricingCache[row.model] && officialPricingCache[row.model] !== 'loading' && (officialPricingCache[row.model] as OfficialPricing).found">
                    <div class="mt-official-strip">
                      <span class="mt-off-label">{{ t('admin.pricingDesk.officialStripLabel') }}</span>
                      <span class="q-money mt-off-price">{{ fmtPrice((officialPricingCache[row.model] as OfficialPricing).inputPrice) }}</span>
                      <span class="mt-off-sep">/</span>
                      <span class="q-money mt-off-price">{{ fmtPrice((officialPricingCache[row.model] as OfficialPricing).outputPrice) }}</span>
                      <span class="mt-off-unit">{{ t('admin.pricingDesk.officialStripPerM') }}</span>
                      <!-- 来源 chip（可点击打开供应商核对抽屉） -->
                      <span
                        v-if="(officialPricingCache[row.model] as OfficialPricing).slug"
                        class="mt-source-chip"
                        :title="t('admin.pricingDesk.sourceChipTitle')"
                        @click.stop="$emit('open-detail', { slug: (officialPricingCache[row.model] as OfficialPricing).slug!, model: row.model })"
                      >{{ (officialPricingCache[row.model] as OfficialPricing).source || 'openrouter' }}</span>
                    </div>
                  </template>
                </td>

                <!-- 每个分组的单元格 -->
                <td
                  v-for="group in activeGroups"
                  :key="group.id"
                  class="mt-td mt-td-cell"
                >
                  <MatrixCell
                    v-if="row.cells[group.id]"
                    :cell="row.cells[group.id]"
                    :model="row.model"
                    :official-pricing="officialPricingCache[row.model]"
                  />
                  <span v-else class="mt-empty-cell">—</span>
                </td>
              </tr>
            </template>
          </template>
        </tbody>
      </table>

      <!-- 加载骨架 -->
      <div v-if="loading" class="mt-skeleton">
        <div v-for="i in 6" :key="i" class="mt-skel-row"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'
import { ChevronDownIcon, CalculatorIcon, RefreshCwIcon } from 'lucide-vue-next'
import MatrixCell from './MatrixCell.vue'
import type { MatrixRow, OfficialPricing } from './usePricingMatrix'
import type { AdminGroup } from '@/types'

const props = defineProps<{
  loading: boolean
  platforms: string[]
  activeGroups: AdminGroup[]
  matrix: MatrixRow[]
  officialPricingCache: Record<string, OfficialPricing | 'loading'>
  syncLoading?: boolean
}>()

const emit = defineEmits<{
  (e: 'hover-model', model: string): void
  (e: 'update-multiplier', groupId: number, value: number): void
  (e: 'sync-catalog'): void
  (e: 'open-detail', payload: { slug: string; model: string }): void
}>()

const { t } = useI18n()

const collapsedPlatforms = ref(new Set<string>())
function togglePlatform(p: string) {
  if (collapsedPlatforms.value.has(p)) {
    collapsedPlatforms.value.delete(p)
  } else {
    collapsedPlatforms.value.add(p)
  }
}

const rowsByPlatform = computed(() => {
  const map: Record<string, MatrixRow[]> = {}
  for (const row of props.matrix) {
    if (!map[row.platform]) map[row.platform] = []
    map[row.platform].push(row)
  }
  return map
})

function onRowHover(model: string) {
  emit('hover-model', model)
}

const editingGroupId = ref<number | null>(null)
const editingMultiplierValue = ref(1)
const multiplierInputRef = ref<HTMLInputElement | null>(null)

function startEditMultiplier(group: AdminGroup) {
  editingGroupId.value = group.id
  editingMultiplierValue.value = group.rate_multiplier
  nextTick(() => multiplierInputRef.value?.select())
}

function cancelEditMultiplier() {
  editingGroupId.value = null
}

async function commitMultiplier(groupId: number) {
  if (editingGroupId.value !== groupId) return
  editingGroupId.value = null
  const v = Number(editingMultiplierValue.value)
  if (!isNaN(v) && v > 0) {
    emit('update-multiplier', groupId, v)
  }
}

/** 价格格式化 per-token → /M */
function fmtPrice(v: number | null | undefined): string {
  if (v == null) return '—'
  const perM = v * 1_000_000
  const decimals = perM >= 1 ? 2 : perM >= 0.1 ? 3 : 4
  return `$${perM.toFixed(decimals)}`
}
</script>

<style scoped>
/* ── 入场动画 ── */
.rise { opacity: 0; transform: translateY(8px); animation: rise .45s cubic-bezier(.22,.68,0,1.2) forwards; }
@keyframes rise { to { opacity: 1; transform: none; } }
@media (prefers-reduced-motion: reduce) { .rise { animation: none; opacity: 1; transform: none; } }

/* ── 空态 ── */
.mt-empty {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 10px; padding: 72px 24px;
  background: var(--metal, linear-gradient(180deg,#15181E,#0E1014));
  border: 1px dashed var(--line-1); border-radius: var(--q-radius, 12px);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.04);
}
.mt-empty-ico { font-size: 36px; opacity: .2; line-height: 1; }
.mt-empty-title { font-size: 14px; font-weight: 600; color: var(--ink-1); margin: 0; }
.mt-empty-hint { font-size: 12px; color: var(--ink-2); margin: 0; text-align: center; }
.mt-empty-btn {
  display: inline-flex; align-items: center; gap: 6px;
  margin-top: 6px; padding: 7px 16px; border-radius: 10px;
  font-size: 12.5px; font-weight: 600; text-decoration: none;
  background: var(--metal-raised, linear-gradient(180deg,#272D37,#14171D));
  border: 1px solid var(--line-1);
  color: var(--ink-0);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.06), 0 2px 8px rgba(0,0,0,.3);
  transition: border-color .18s, box-shadow .18s;
}
.mt-empty-btn:hover { border-color: rgba(92,168,255,.45); box-shadow: inset 0 1px 0 rgba(255,255,255,.06), 0 0 12px rgba(92,168,255,.18); }
.mt-empty-btn:focus-visible { outline: none; box-shadow: var(--glow-focus); }
.mt-empty-btn-ico { width: 14px; height: 14px; color: var(--azure); }

/* ── 矩阵面板 ── */
.mt-panel {
  display: flex; flex-direction: column; gap: 0;
  background: var(--metal, linear-gradient(180deg,#15181E,#0E1014));
  border: 1px solid var(--line-0); border-radius: var(--q-radius, 12px);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.04), 0 8px 22px rgba(0,0,0,.28);
  overflow: hidden;
}

/* ── 图例 ── */
.mt-legend {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 16px 7px;
  border-bottom: 1px solid var(--line-0);
  background: var(--bg-2);
}
.mt-legend-item { display: inline-flex; align-items: center; gap: 5px; font-size: 10.5px; color: var(--ink-2); }
.mt-legend-sep { color: var(--ink-2); opacity: .4; font-size: 10px; }
.mt-legend-dot { display: inline-block; width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.mt-ok { background: var(--ok); }
.mt-neutral { background: var(--ink-2); }
.mt-bad { background: var(--bad); }
.mt-legend-spacer { flex: 1; }

/* 同步目录按钮 */
.mt-sync-btn {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 4px 11px; border-radius: 8px;
  font-size: 11px; font-weight: 600;
  background: var(--metal-raised, linear-gradient(180deg,#272D37,#14171D));
  border: 1px solid var(--line-1); color: var(--ink-1);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.06), 0 2px 6px rgba(0,0,0,.25);
  cursor: pointer; transition: border-color .18s, box-shadow .18s, color .18s;
  white-space: nowrap;
}
.mt-sync-btn:hover:not(:disabled) { border-color: rgba(92,168,255,.45); color: var(--ink-0); box-shadow: inset 0 1px 0 rgba(255,255,255,.06), 0 0 10px rgba(92,168,255,.18); }
.mt-sync-btn:disabled { opacity: .5; cursor: default; }
.mt-sync-btn:focus-visible { outline: none; box-shadow: var(--glow-focus); }
.mt-sync-ico { width: 12px; height: 12px; flex-shrink: 0; }
.mt-spinning { animation: mt-spin 1s linear infinite; }
@keyframes mt-spin { to { transform: rotate(360deg); } }
@media (prefers-reduced-motion: reduce) { .mt-spinning { animation: none; } }

/* ── 表格滚动容器 ── */
.mt-scroll-wrap { overflow-x: auto; }

/* ── 表格 ── */
.mt-table { width: 100%; border-collapse: collapse; font-size: 12px; }
.mt-sticky {
  position: sticky; left: 0; z-index: 2;
  background: var(--metal, linear-gradient(180deg,#15181E,#0E1014));
}

/* 表头行 */
.mt-head-row { background: var(--bg-2); border-bottom: 1px solid var(--line-0); }
.mt-th {
  padding: 10px 12px; font-size: 10.5px; font-weight: 600;
  letter-spacing: .05em; text-transform: uppercase;
  color: var(--ink-2); white-space: nowrap;
}
.mt-th-model { text-align: left; min-width: 200px; background: var(--bg-2); }
.mt-th-group { text-align: center; min-width: 130px; }
.mt-th-inner { display: flex; flex-direction: column; align-items: center; gap: 4px; }
.mt-group-name { color: var(--ink-0); font-size: 11px; font-weight: 600; text-transform: none; letter-spacing: 0; }
.mt-multiplier {
  color: var(--azure); font-size: 10.5px; cursor: pointer;
  padding: 1px 5px; border-radius: 4px;
  border: 1px solid transparent;
  transition: border-color .15s, background .15s;
}
.mt-multiplier:hover { border-color: rgba(92,168,255,.4); background: var(--azure-dim); }
.mt-multiplier-wrap { display: flex; align-items: center; justify-content: center; }
.mt-multiplier-input {
  width: 62px; padding: 2px 4px; border-radius: 5px;
  font-size: 11px; text-align: center; font-family: inherit;
  background: var(--bg-0); border: 1px solid var(--azure);
  color: var(--ink-0); outline: none;
}

/* platform 行组标题 */
.mt-platform-row {
  cursor: pointer; user-select: none;
  background: var(--bg-2); border-top: 1px solid var(--line-0);
}
.mt-platform-row:hover { background: var(--bg-1); }
.mt-platform-cell { padding: 7px 16px; }
.mt-platform-inner { display: flex; align-items: center; gap: 8px; }
.mt-chevron {
  width: 14px; height: 14px; color: var(--ink-2);
  transition: transform .2s cubic-bezier(.22,.68,0,1.2);
  flex-shrink: 0;
}
@media (prefers-reduced-motion: reduce) { .mt-chevron { transition: none; } }
.mt-collapsed { transform: rotate(-90deg); }
.mt-platform-name {
  font-size: 10.5px; font-weight: 700; letter-spacing: .08em;
  text-transform: uppercase; color: var(--ink-1);
}
.mt-platform-count { font-size: 10.5px; color: var(--ink-2); }

/* 模型行 */
.mt-model-row {
  border-bottom: 1px solid var(--line-0);
  transition: background .12s;
}
.mt-model-row:hover { background: rgba(255,255,255,.015); }
.mt-model-row:hover .mt-sticky { background: rgba(22,26,33,.98); }
.mt-td { padding: 5px 8px; vertical-align: middle; }
.mt-td-model {
  padding: 5px 16px;
}
.mt-model-name {
  display: block;
  font-family: 'IBM Plex Mono', monospace; font-size: 11.5px;
  color: var(--ink-0); white-space: nowrap;
}

/* 官方价基准条 */
.mt-official-strip {
  display: flex; align-items: center; gap: 4px;
  margin-top: 2px; flex-wrap: nowrap; white-space: nowrap;
}
.mt-off-label {
  font-size: 9px; font-weight: 600; letter-spacing: .06em;
  text-transform: uppercase; color: var(--ink-2);
}
.mt-off-price { font-size: 10px; color: var(--ink-1); }
.mt-off-sep { font-size: 9px; color: var(--ink-2); opacity: .5; }
.mt-off-unit { font-size: 9px; color: var(--ink-2); }

/* 来源 chip */
.mt-source-chip {
  display: inline-flex; align-items: center;
  padding: 0 5px; border-radius: 4px; height: 14px;
  font-size: 8.5px; font-weight: 700; letter-spacing: .05em; text-transform: uppercase;
  background: var(--azure-dim, rgba(92,168,255,.12));
  border: 1px solid rgba(92,168,255,.25);
  color: var(--azure);
  cursor: pointer; user-select: none;
  transition: background .15s, border-color .15s;
}
.mt-source-chip:hover { background: rgba(92,168,255,.2); border-color: rgba(92,168,255,.5); }
.mt-td-cell { text-align: center; }
.mt-empty-cell { font-size: 11px; color: var(--ink-2); }

/* 骨架 */
.mt-skeleton { padding: 12px 16px; display: flex; flex-direction: column; gap: 8px; }
.mt-skel-row {
  height: 38px; border-radius: 6px;
  background: var(--bg-2);
  animation: skel-pulse 1.4s ease-in-out infinite;
}
@keyframes skel-pulse { 0%,100%{ opacity:.6 } 50%{ opacity:1 } }
@media (prefers-reduced-motion: reduce) { .mt-skel-row { animation: none; } }
</style>
