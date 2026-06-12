<template>
  <Transition name="pvd-fade"><div v-if="open" class="pvd-overlay" aria-hidden="true" @click="$emit('close')" /></Transition>
  <Transition name="pvd-slide">
    <div v-if="open" class="pvd-drawer" role="dialog" aria-modal="true" @keydown.esc="$emit('close')">
      <div class="pvd-head">
        <div class="pvd-head-left">
          <LayersIcon class="pvd-head-ico" />
          <div class="pvd-head-text">
            <h2 class="pvd-head-title">{{ detail?.name || modelName || slug || '—' }}</h2>
            <span v-if="detail?.id || slug" class="pvd-head-slug q-money">{{ detail?.id || slug }}</span>
          </div>
        </div>
        <button class="pvd-close q-focus-glow" :aria-label="t('admin.providerVerify.close')" @click="$emit('close')"><XIcon class="pvd-close-ico" /></button>
      </div>

      <div v-if="loading" class="pvd-body">
        <div class="pvd-skel-block"></div><div class="pvd-skel-short"></div>
        <div class="pvd-skel-table"><div v-for="i in 4" :key="i" class="pvd-skel-row"></div></div>
      </div>
      <div v-else-if="error" class="pvd-body"><div class="pvd-error">{{ t('admin.providerVerify.loadError') }}{{ error }}</div></div>
      <div v-else-if="detail" class="pvd-body">
        <div class="pvd-meta-row">
          <span v-if="detail.context_len" class="pvd-meta-chip"><HashIcon class="pvd-chip-ico" />{{ fmtCtx(detail.context_len) }} ctx</span>
          <span v-for="cap in filteredCaps" :key="cap" class="pvd-meta-chip pvd-cap-chip">{{ capLabel(cap) }}</span>
          <!-- 已覆盖徽章 -->
          <span v-if="detail.overridden" class="pvd-meta-chip pvd-override-badge">
            <ShieldCheckIcon class="pvd-chip-ico" />{{ t('admin.providerVerify.overriddenBadge') }}
          </span>
        </div>
        <div v-if="detail.description" class="pvd-desc-block">
          <p class="pvd-desc-text" :class="descExpanded ? '' : 'pvd-desc-clamp'">{{ detail.description }}</p>
          <button class="pvd-desc-more" @click="descExpanded = !descExpanded">{{ descExpanded ? t('admin.providerVerify.showLess') : t('admin.providerVerify.showMore') }}</button>
        </div>
        <div v-if="detail.baseline" class="pvd-baseline-note" :class="detail.overridden ? 'pvd-baseline-note--overridden' : ''">
          <InfoIcon class="pvd-note-ico" />
          <span>{{ t('admin.providerVerify.baselineNote') }}</span>
          <span v-if="detail.baseline.source" class="pvd-note-source">{{ detail.baseline.source }}</span>
          <span v-if="detail.overridden" class="pvd-note-manual-tag">{{ t('admin.providerVerify.overriddenLabel') }}</span>
        </div>
        <div class="pvd-table-wrap">
          <div v-if="syncing" class="pvd-sync-overlay"><RefreshCwIcon class="pvd-sync-ico" /><span class="pvd-sync-txt">{{ t('admin.providerVerify.syncing') }}</span></div>
          <table v-if="sortedProviders.length" class="pvd-table">
            <thead><tr class="pvd-thead-row">
              <th class="pvd-th pvd-th-provider">{{ t('admin.providerVerify.colProvider') }}</th>
              <th class="pvd-th pvd-th-r">{{ t('admin.providerVerify.colIn') }}</th>
              <th class="pvd-th pvd-th-r">{{ t('admin.providerVerify.colOut') }}</th>
              <th class="pvd-th pvd-th-r">{{ t('admin.providerVerify.colCacheRead') }}</th>
              <th class="pvd-th pvd-th-r">{{ t('admin.providerVerify.colCacheWrite') }}</th>
              <th class="pvd-th pvd-th-r">{{ t('admin.providerVerify.colUptime') }}</th>
              <th class="pvd-th pvd-th-c">{{ t('admin.providerVerify.colQuant') }}</th>
            </tr></thead>
            <tbody>
              <tr v-for="prov in sortedProviders" :key="prov.tag" class="pvd-tr" :class="[isBaseline(prov) ? 'pvd-tr-bl' : '', isPinned(prov) ? 'pvd-tr-pinned' : '']">
                <td class="pvd-td">
                  <div class="pvd-prov-cell">
                    <span v-if="isBaseline(prov)" class="pvd-bl-edge" aria-hidden="true"></span>
                    <span class="pvd-prov-name">{{ prov.provider || prov.tag }}</span>
                    <span v-if="isBaseline(prov) && !detail.overridden" class="pvd-bl-badge">{{ t('admin.providerVerify.baselineBadge') }}</span>
                    <span v-if="isPinned(prov)" class="pvd-pinned-badge">{{ t('admin.providerVerify.pinnedBadge') }}</span>
                  </div>
                </td>
                <td class="pvd-td pvd-td-r"><span class="q-money pvd-price">{{ fmtP(prov.input) }}</span></td>
                <td class="pvd-td pvd-td-r"><span class="q-money pvd-price">{{ fmtP(prov.output) }}</span></td>
                <td class="pvd-td pvd-td-r"><span class="q-money pvd-price pvd-muted">{{ fmtP(prov.cache_read) }}</span></td>
                <td class="pvd-td pvd-td-r"><span class="q-money pvd-price pvd-muted">{{ fmtP(prov.cache_write) }}</span></td>
                <td class="pvd-td pvd-td-r"><span class="pvd-uptime" :class="uptimeCls(prov.uptime_1d)">{{ fmtUp(prov.uptime_1d) }}</span></td>
                <td class="pvd-td pvd-td-c"><span v-if="prov.quant" class="pvd-quant">{{ prov.quant }}</span><span v-else class="pvd-muted">—</span></td>
              </tr>
            </tbody>
          </table>
          <div v-else-if="!syncing" class="pvd-empty-prov"><PackageSearchIcon class="pvd-empty-ico" /><p class="pvd-empty-txt">{{ t('admin.providerVerify.noProviders') }}</p></div>
        </div>

        <!-- 编辑覆盖价区域 -->
        <div class="pvd-foot">
          <button class="pvd-edit-btn q-focus-glow" :class="editOpen ? 'pvd-edit-btn--active' : ''" @click="toggleEdit">
            <EditIcon class="pvd-edit-ico" />{{ t('admin.providerVerify.editBtn') }}
            <ChevronDownIcon class="pvd-chevron" :class="editOpen ? 'pvd-chevron--up' : ''" />
          </button>
        </div>

        <!-- 覆盖编辑面板（展开） -->
        <Transition name="pvd-panel">
          <div v-if="editOpen" class="pvd-edit-panel">
            <p class="pvd-panel-title">{{ t('admin.providerVerify.panelTitle') }}</p>

            <!-- 三态单选 -->
            <div class="pvd-radio-group">
              <label class="pvd-radio-item" :class="overrideMode === 'auto' ? 'pvd-radio-item--active' : ''">
                <input v-model="overrideMode" type="radio" value="auto" class="pvd-radio-input" />
                <span class="pvd-radio-dot"></span>
                <span class="pvd-radio-label">{{ t('admin.providerVerify.modeAuto') }}</span>
                <span class="pvd-radio-hint">{{ t('admin.providerVerify.modeAutoHint') }}</span>
              </label>
              <label class="pvd-radio-item" :class="overrideMode === 'pinned' ? 'pvd-radio-item--active' : ''">
                <input v-model="overrideMode" type="radio" value="pinned" class="pvd-radio-input" />
                <span class="pvd-radio-dot"></span>
                <span class="pvd-radio-label">{{ t('admin.providerVerify.modePinned') }}</span>
              </label>
              <!-- 指定供应商下拉 -->
              <div v-if="overrideMode === 'pinned'" class="pvd-pinned-select-wrap">
                <select v-model="pinnedTag" class="pvd-select q-focus-glow">
                  <option value="">{{ t('admin.providerVerify.selectProviderPlaceholder') }}</option>
                  <option v-for="p in detail.providers" :key="p.tag" :value="p.tag">
                    {{ p.provider || p.tag }} ({{ fmtP(p.input) }} / {{ fmtP(p.output) }})
                  </option>
                </select>
              </div>
              <label class="pvd-radio-item" :class="overrideMode === 'manual' ? 'pvd-radio-item--active' : ''">
                <input v-model="overrideMode" type="radio" value="manual" class="pvd-radio-input" />
                <span class="pvd-radio-dot"></span>
                <span class="pvd-radio-label">{{ t('admin.providerVerify.modeManual') }}</span>
                <span class="pvd-radio-hint">{{ t('admin.providerVerify.modeManualHint') }}</span>
              </label>
              <!-- 手动输入框（per-MTok） -->
              <div v-if="overrideMode === 'manual'" class="pvd-manual-grid">
                <div class="pvd-manual-field">
                  <label class="pvd-field-label">{{ t('admin.providerVerify.fieldInput') }}</label>
                  <div class="pvd-input-wrap">
                    <span class="pvd-input-prefix">$</span>
                    <input v-model="manualInputMtok" type="number" min="0" step="0.001" class="pvd-number-input q-money q-focus-glow" :placeholder="t('admin.providerVerify.fieldPlaceholder')" />
                    <span class="pvd-input-suffix">/M</span>
                  </div>
                </div>
                <div class="pvd-manual-field">
                  <label class="pvd-field-label">{{ t('admin.providerVerify.fieldOutput') }}</label>
                  <div class="pvd-input-wrap">
                    <span class="pvd-input-prefix">$</span>
                    <input v-model="manualOutputMtok" type="number" min="0" step="0.001" class="pvd-number-input q-money q-focus-glow" :placeholder="t('admin.providerVerify.fieldPlaceholder')" />
                    <span class="pvd-input-suffix">/M</span>
                  </div>
                </div>
                <div class="pvd-manual-field">
                  <label class="pvd-field-label">{{ t('admin.providerVerify.fieldCacheRead') }}</label>
                  <div class="pvd-input-wrap">
                    <span class="pvd-input-prefix">$</span>
                    <input v-model="manualCacheReadMtok" type="number" min="0" step="0.001" class="pvd-number-input q-money q-focus-glow" :placeholder="t('admin.providerVerify.fieldPlaceholder')" />
                    <span class="pvd-input-suffix">/M</span>
                  </div>
                </div>
                <div class="pvd-manual-field">
                  <label class="pvd-field-label">{{ t('admin.providerVerify.fieldCacheWrite') }}</label>
                  <div class="pvd-input-wrap">
                    <span class="pvd-input-prefix">$</span>
                    <input v-model="manualCacheWriteMtok" type="number" min="0" step="0.001" class="pvd-number-input q-money q-focus-glow" :placeholder="t('admin.providerVerify.fieldPlaceholder')" />
                    <span class="pvd-input-suffix">/M</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- 备注 -->
            <div class="pvd-note-field">
              <label class="pvd-field-label">{{ t('admin.providerVerify.noteLabel') }}</label>
              <input v-model="noteText" type="text" maxlength="200" class="pvd-note-input q-focus-glow" :placeholder="t('admin.providerVerify.notePlaceholder')" />
            </div>

            <!-- 保存/恢复操作栏 -->
            <div class="pvd-action-row">
              <span v-if="saveError" class="pvd-save-error">{{ saveError }}</span>
              <span v-if="saveOk" class="pvd-save-ok">{{ t('admin.providerVerify.saveSuccess') }}</span>
              <button v-if="overrideMode === 'auto'" class="pvd-restore-btn q-focus-glow" :disabled="saving || !detail.overridden" @click="handleRestore">
                <Trash2Icon class="pvd-action-ico" />{{ t('admin.providerVerify.restoreBtn') }}
              </button>
              <button v-else class="pvd-save-btn q-focus-glow" :disabled="saving || !canSave" @click="handleSave">
                <SaveIcon v-if="!saving" class="pvd-action-ico" />
                <RefreshCwIcon v-else class="pvd-action-ico pvd-saving-spin" />
                {{ saving ? t('admin.providerVerify.saving') : t('admin.providerVerify.saveBtn') }}
              </button>
            </div>
          </div>
        </Transition>
      </div>
      <div v-else class="pvd-body pvd-empty-state"><PackageSearchIcon class="pvd-empty-ico" /><p class="pvd-empty-txt">{{ t('admin.providerVerify.noSlug') }}</p></div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  XIcon, LayersIcon, HashIcon, InfoIcon, RefreshCwIcon, EditIcon,
  PackageSearchIcon, ChevronDownIcon, SaveIcon, Trash2Icon, ShieldCheckIcon
} from 'lucide-vue-next'
import modelCatalogAPI from '@/api/admin/modelCatalog'
import type { CatalogModelDetail, CatalogProvider } from '@/api/admin/modelCatalog'

const props = defineProps<{ open: boolean; slug: string | null; modelName?: string | null }>()
const emit = defineEmits<{
  (e: 'close'): void
  /** 覆盖保存或删除后通知父组件失效官方价缓存 */
  (e: 'override-saved', modelId: string): void
}>()
const { t } = useI18n()

const detail = ref<CatalogModelDetail | null>(null)
const loading = ref(false)
const syncing = ref(false)
const error = ref<string | null>(null)
const descExpanded = ref(false)

// ── 编辑面板状态 ──
const editOpen = ref(false)
type OverrideMode = 'auto' | 'pinned' | 'manual'
const overrideMode = ref<OverrideMode>('auto')
const pinnedTag = ref('')
const manualInputMtok = ref<string>('')
const manualOutputMtok = ref<string>('')
const manualCacheReadMtok = ref<string>('')
const manualCacheWriteMtok = ref<string>('')
const noteText = ref('')
const saving = ref(false)
const saveError = ref<string | null>(null)
const saveOk = ref(false)

watch(() => [props.open, props.slug] as const, ([nowOpen, nowSlug]) => {
  if (nowOpen && nowSlug) { void load(nowSlug) }
  else if (!nowOpen) {
    detail.value = null
    error.value = null
    descExpanded.value = false
    editOpen.value = false
  }
}, { immediate: true })

async function load(slug: string) {
  loading.value = true; error.value = null; detail.value = null
  editOpen.value = false
  try {
    const d = await modelCatalogAPI.getModelCatalogDetail(slug)
    if (d.providers.length === 0) {
      detail.value = d; loading.value = false; syncing.value = true
      try {
        await modelCatalogAPI.syncModelEndpoints(slug)
        detail.value = await modelCatalogAPI.getModelCatalogDetail(slug)
      } catch (e) { console.warn('[ProviderVerifyDrawer] sync failed', e) }
      finally { syncing.value = false }
    } else { detail.value = d }
    // 用现有 override 填充表单
    seedFormFromDetail(detail.value)
  } catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
  finally { loading.value = false }
}

function seedFormFromDetail(d: CatalogModelDetail | null) {
  const ov = d?.override
  if (!ov) {
    overrideMode.value = 'auto'
    pinnedTag.value = ''
    manualInputMtok.value = ''
    manualOutputMtok.value = ''
    manualCacheReadMtok.value = ''
    manualCacheWriteMtok.value = ''
    noteText.value = ''
    return
  }
  noteText.value = ov.note ?? ''
  // 判断模式
  const hasManual = ov.manual_input != null || ov.manual_output != null ||
    ov.manual_cache_read != null || ov.manual_cache_write != null
  if (hasManual) {
    overrideMode.value = 'manual'
    manualInputMtok.value = ov.manual_input != null ? String(ov.manual_input * 1e6) : ''
    manualOutputMtok.value = ov.manual_output != null ? String(ov.manual_output * 1e6) : ''
    manualCacheReadMtok.value = ov.manual_cache_read != null ? String(ov.manual_cache_read * 1e6) : ''
    manualCacheWriteMtok.value = ov.manual_cache_write != null ? String(ov.manual_cache_write * 1e6) : ''
    pinnedTag.value = ''
  } else if (ov.pinned_provider_tag) {
    overrideMode.value = 'pinned'
    pinnedTag.value = ov.pinned_provider_tag
  } else {
    overrideMode.value = 'auto'
    pinnedTag.value = ''
  }
}

function toggleEdit() {
  editOpen.value = !editOpen.value
  saveError.value = null
  saveOk.value = false
}

const canSave = computed(() => {
  if (overrideMode.value === 'pinned') return !!pinnedTag.value
  if (overrideMode.value === 'manual') {
    return !!(manualInputMtok.value || manualOutputMtok.value ||
      manualCacheReadMtok.value || manualCacheWriteMtok.value)
  }
  return false
})

function parseMtok(v: string): number | null {
  if (!v && v !== '0') return null
  const n = parseFloat(v)
  return isNaN(n) ? null : n / 1e6
}

async function handleSave() {
  if (!detail.value || saving.value) return
  saving.value = true
  saveError.value = null
  saveOk.value = false
  try {
    const payload: Parameters<typeof modelCatalogAPI.putModelOverride>[0] = {
      model_id: detail.value.id,
      note: noteText.value || undefined
    }
    if (overrideMode.value === 'pinned') {
      payload.pinned_provider_tag = pinnedTag.value
    } else if (overrideMode.value === 'manual') {
      payload.manual_input = parseMtok(manualInputMtok.value)
      payload.manual_output = parseMtok(manualOutputMtok.value)
      payload.manual_cache_read = parseMtok(manualCacheReadMtok.value)
      payload.manual_cache_write = parseMtok(manualCacheWriteMtok.value)
    }
    await modelCatalogAPI.putModelOverride(payload)
    // 重拉详情以刷新 overridden / baseline
    detail.value = await modelCatalogAPI.getModelCatalogDetail(detail.value.id)
    saveOk.value = true
    setTimeout(() => { saveOk.value = false }, 3000)
    emit('override-saved', props.modelName ?? detail.value.id)
  } catch (e: unknown) {
    saveError.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

async function handleRestore() {
  if (!detail.value || saving.value) return
  saving.value = true
  saveError.value = null
  saveOk.value = false
  try {
    await modelCatalogAPI.deleteModelOverride(detail.value.id)
    detail.value = await modelCatalogAPI.getModelCatalogDetail(detail.value.id)
    seedFormFromDetail(detail.value)
    saveOk.value = true
    setTimeout(() => { saveOk.value = false }, 3000)
    emit('override-saved', props.modelName ?? detail.value.id)
  } catch (e: unknown) {
    saveError.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

const sortedProviders = computed<CatalogProvider[]>(() => {
  if (!detail.value) return []
  const src = detail.value.baseline?.source
  return [...detail.value.providers].sort((a, b) => {
    if (a.tag === src && b.tag !== src) return -1
    if (a.tag !== src && b.tag === src) return 1
    return (a.input ?? Infinity) - (b.input ?? Infinity)
  })
})

const SHOW_CAPS = ['reasoning', 'tools', 'structured_outputs', 'vision', 'image-generation']
const CAP_LABELS: Record<string, string> = {
  reasoning: 'Reasoning', tools: 'Tool Use', structured_outputs: 'Structured Out',
  vision: 'Vision', 'image-generation': 'Image Gen'
}
const filteredCaps = computed(() => detail.value?.capabilities?.filter(c => SHOW_CAPS.includes(c)) ?? [])
function capLabel(c: string) { return CAP_LABELS[c] ?? c }
function isBaseline(p: CatalogProvider) { return p.tag === detail.value?.baseline?.source }
function isPinned(p: CatalogProvider) {
  return !!(detail.value?.override?.pinned_provider_tag && p.tag === detail.value.override.pinned_provider_tag)
}

function fmtP(v?: number | null): string {
  if (v == null) return '—'
  const m = v * 1e6
  return `$${m.toFixed(m >= 1 ? 2 : m >= 0.1 ? 3 : 4)}`
}
function fmtCtx(n: number) { return n >= 1e6 ? `${(n/1e6).toFixed(1)}M` : n >= 1e3 ? `${(n/1e3).toFixed(0)}k` : String(n) }
function fmtUp(u?: number) { return u == null ? '—' : `${u.toFixed(1)}%` }
function uptimeCls(u?: number) { return u == null ? 'pvd-up-muted' : u < 95 ? 'pvd-up-warn' : 'pvd-up-ok' }
</script>

<style scoped>
.pvd-overlay { position:fixed;inset:0;z-index:49;background:rgba(0,0,0,.45); }
.pvd-drawer {
  position:fixed;right:0;top:0;z-index:50;width:100%;max-width:560px;height:100%;overflow-y:auto;
  display:flex;flex-direction:column;
  background:var(--metal,linear-gradient(180deg,#15181E,#0E1014));
  border-left:1px solid var(--line-1);
  box-shadow:var(--edge-hi,inset 0 1px 0 rgba(255,255,255,.04)),-8px 0 40px rgba(0,0,0,.5);
}
.pvd-head { display:flex;align-items:center;justify-content:space-between;padding:16px 20px;border-bottom:1px solid var(--line-0);background:var(--bg-2);flex-shrink:0; }
.pvd-head-left { display:flex;align-items:center;gap:10px;min-width:0; }
.pvd-head-ico { width:18px;height:18px;color:var(--azure);flex-shrink:0; }
.pvd-head-text { display:flex;flex-direction:column;gap:2px;min-width:0; }
.pvd-head-title { font-size:14px;font-weight:700;color:var(--ink-0);margin:0;white-space:nowrap;overflow:hidden;text-overflow:ellipsis; }
.pvd-head-slug { font-size:10.5px;color:var(--ink-2);white-space:nowrap;overflow:hidden;text-overflow:ellipsis; }
.pvd-close { display:inline-flex;align-items:center;justify-content:center;padding:5px;border-radius:8px;border:1px solid transparent;background:transparent;color:var(--ink-2);cursor:pointer;transition:border-color .15s,color .15s,background .15s; }
.pvd-close:hover { border-color:var(--line-1);color:var(--ink-0);background:var(--bg-1); }
.pvd-close-ico { width:16px;height:16px; }
.pvd-body { display:flex;flex-direction:column;gap:14px;padding:18px 20px;flex:1; }
.pvd-skel-block,.pvd-skel-short,.pvd-skel-row { background:var(--bg-2);border-radius:8px;animation:pvd-skel 1.4s ease-in-out infinite; }
.pvd-skel-block { height:44px; }
.pvd-skel-short { height:20px;width:60%; }
.pvd-skel-table { display:flex;flex-direction:column;gap:6px; }
.pvd-skel-row { height:34px; }
@keyframes pvd-skel { 0%,100%{opacity:.5}50%{opacity:.9} }
@media (prefers-reduced-motion:reduce){.pvd-skel-block,.pvd-skel-short,.pvd-skel-row{animation:none}}
.pvd-error { padding:10px 14px;border-radius:10px;font-size:12.5px;background:var(--bad-dim);border:1px solid var(--bad);color:var(--bad); }
.pvd-meta-row { display:flex;flex-wrap:wrap;gap:6px;align-items:center; }
.pvd-meta-chip { display:inline-flex;align-items:center;gap:4px;padding:2px 8px;border-radius:5px;font-size:10.5px;font-weight:600;background:var(--bg-2);border:1px solid var(--line-1);color:var(--ink-1); }
.pvd-cap-chip { background:var(--azure-dim,rgba(92,168,255,.12));border-color:rgba(92,168,255,.3);color:var(--azure); }
.pvd-override-badge { background:rgba(92,168,255,.15);border-color:rgba(92,168,255,.4);color:var(--azure);font-weight:700; }
.pvd-chip-ico { width:11px;height:11px; }
.pvd-desc-block { padding:12px 14px;background:var(--bg-2);border:1px solid var(--line-0);border-radius:10px; }
.pvd-desc-text { font-size:12px;line-height:1.65;color:var(--ink-1);margin:0; }
.pvd-desc-clamp { display:-webkit-box;-webkit-line-clamp:3;-webkit-box-orient:vertical;overflow:hidden; }
.pvd-desc-more { margin-top:6px;padding:0;border:none;background:none;font-size:11px;color:var(--azure);cursor:pointer;font-weight:600; }
.pvd-desc-more:hover { text-decoration:underline; }
.pvd-baseline-note { display:flex;align-items:center;gap:6px;padding:8px 12px;border-radius:8px;background:var(--azure-dim,rgba(92,168,255,.08));border:1px solid rgba(92,168,255,.2);font-size:11.5px;color:var(--ink-1); }
.pvd-baseline-note--overridden { background:rgba(92,168,255,.13);border-color:rgba(92,168,255,.35); }
.pvd-note-ico { width:13px;height:13px;color:var(--azure);flex-shrink:0; }
.pvd-note-source { font-family:'IBM Plex Mono',monospace;font-size:10.5px;color:var(--azure);font-weight:600;padding:1px 5px;border-radius:4px;background:rgba(92,168,255,.12);border:1px solid rgba(92,168,255,.25); }
.pvd-note-manual-tag { font-size:10px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;padding:1px 5px;border-radius:4px;background:rgba(92,168,255,.2);border:1px solid rgba(92,168,255,.4);color:var(--azure); }
.pvd-table-wrap { position:relative;border:1px solid var(--line-0);border-radius:10px;overflow:hidden; }
.pvd-sync-overlay { position:absolute;inset:0;z-index:2;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:8px;background:rgba(14,16,20,.72);border-radius:10px; }
.pvd-sync-ico { width:22px;height:22px;color:var(--azure);animation:pvd-spin 1s linear infinite; }
@keyframes pvd-spin{to{transform:rotate(360deg)}}
@media(prefers-reduced-motion:reduce){.pvd-sync-ico{animation:none}}
.pvd-sync-txt { font-size:11.5px;color:var(--ink-1); }
.pvd-table { width:100%;border-collapse:collapse;font-size:12px; }
.pvd-thead-row { background:var(--bg-2);border-bottom:1px solid var(--line-0); }
.pvd-th { padding:8px 10px;font-size:9.5px;font-weight:600;letter-spacing:.06em;text-transform:uppercase;color:var(--ink-2);white-space:nowrap;text-align:left; }
.pvd-th-r { text-align:right; }.pvd-th-c { text-align:center; }.pvd-th-provider { min-width:140px; }
.pvd-tr { border-top:1px solid var(--line-0);transition:background .12s; }
.pvd-tr:hover { background:rgba(255,255,255,.018); }
.pvd-tr-bl { background:var(--azure-dim,rgba(92,168,255,.07)); }
.pvd-tr-bl:hover { background:rgba(92,168,255,.12); }
.pvd-tr-pinned { background:rgba(92,168,255,.09); }
.pvd-tr-pinned:hover { background:rgba(92,168,255,.14); }
.pvd-td { padding:9px 10px;vertical-align:middle; }
.pvd-td-r { text-align:right; }.pvd-td-c { text-align:center; }
.pvd-prov-cell { display:flex;align-items:center;gap:7px; }
.pvd-bl-edge { display:inline-block;width:3px;height:20px;border-radius:2px;background:var(--azure);flex-shrink:0; }
.pvd-prov-name { font-size:12px;font-weight:500;color:var(--ink-0); }
.pvd-bl-badge { font-size:9px;font-weight:700;letter-spacing:.06em;text-transform:uppercase;padding:1px 5px;border-radius:4px;background:var(--azure-dim,rgba(92,168,255,.15));border:1px solid rgba(92,168,255,.3);color:var(--azure); }
.pvd-pinned-badge { font-size:9px;font-weight:700;letter-spacing:.06em;text-transform:uppercase;padding:1px 5px;border-radius:4px;background:rgba(92,168,255,.2);border:1px solid rgba(92,168,255,.45);color:var(--azure); }
.pvd-price { font-size:11.5px;font-variant-numeric:tabular-nums; }
.pvd-muted { color:var(--ink-2)!important; }
.pvd-uptime { font-size:11px;font-variant-numeric:tabular-nums; }
.pvd-up-ok { color:var(--ok); }.pvd-up-warn { color:var(--warm,#F4A64A); }.pvd-up-muted { color:var(--ink-2); }
.pvd-quant { display:inline-block;padding:1px 5px;border-radius:4px;font-size:9.5px;font-weight:600;letter-spacing:.04em;background:var(--bg-2);border:1px solid var(--line-1);color:var(--ink-2); }
.pvd-empty-prov,.pvd-empty-state { display:flex;flex-direction:column;align-items:center;justify-content:center;gap:8px;padding:36px 16px;background:var(--bg-2); }
.pvd-empty-ico { width:32px;height:32px;color:var(--ink-2);opacity:.3; }
.pvd-empty-txt { font-size:12px;color:var(--ink-2);margin:0;text-align:center; }

/* 编辑按钮 */
.pvd-foot { display:flex;align-items:center;gap:10px;padding-top:10px;border-top:1px solid var(--line-0); }
.pvd-edit-btn {
  display:inline-flex;align-items:center;gap:6px;padding:7px 15px;border-radius:10px;
  font-size:12.5px;font-weight:600;
  background:var(--metal-raised,linear-gradient(180deg,#272D37,#14171D));
  border:1px solid var(--line-1);color:var(--ink-1);
  box-shadow:inset 0 1px 0 rgba(255,255,255,.06);
  cursor:pointer;transition:border-color .15s,color .15s,background .15s,box-shadow .15s;
}
.pvd-edit-btn:hover { border-color:var(--azure);color:var(--azure);box-shadow:inset 0 1px 0 rgba(255,255,255,.08),0 0 0 1px rgba(92,168,255,.18); }
.pvd-edit-btn--active { border-color:var(--azure);color:var(--azure);background:linear-gradient(180deg,#1e2838,#131720);box-shadow:inset 0 1px 0 rgba(255,255,255,.05),0 0 0 1px rgba(92,168,255,.25); }
.pvd-edit-ico { width:13px;height:13px; }
.pvd-chevron { width:13px;height:13px;transition:transform .2s;flex-shrink:0; }
.pvd-chevron--up { transform:rotate(180deg); }
@media(prefers-reduced-motion:reduce){.pvd-chevron{transition:none}}

/* 编辑面板 */
.pvd-edit-panel {
  padding:16px;border:1px solid rgba(92,168,255,.22);border-radius:12px;
  background:linear-gradient(180deg,rgba(92,168,255,.05),rgba(92,168,255,.02));
  display:flex;flex-direction:column;gap:14px;
}
.pvd-panel-title { font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--azure);margin:0; }
.pvd-radio-group { display:flex;flex-direction:column;gap:8px; }
.pvd-radio-item {
  display:flex;align-items:center;gap:9px;padding:9px 12px;border-radius:9px;
  border:1px solid var(--line-0);background:var(--bg-2);cursor:pointer;
  transition:border-color .14s,background .14s;
}
.pvd-radio-item:hover { border-color:var(--line-1);background:var(--bg-1); }
.pvd-radio-item--active { border-color:rgba(92,168,255,.4);background:rgba(92,168,255,.07); }
.pvd-radio-input { position:absolute;opacity:0;width:0;height:0; }
.pvd-radio-dot {
  width:14px;height:14px;border-radius:50%;border:2px solid var(--line-1);flex-shrink:0;
  transition:border-color .14s,background .14s;
}
.pvd-radio-item--active .pvd-radio-dot { border-color:var(--azure);background:var(--azure);box-shadow:0 0 0 3px rgba(92,168,255,.18); }
.pvd-radio-label { font-size:12.5px;font-weight:600;color:var(--ink-0); }
.pvd-radio-hint { font-size:10.5px;color:var(--ink-2);margin-left:auto; }

/* 指定供应商下拉 */
.pvd-pinned-select-wrap { padding:0 12px 4px 35px; }
.pvd-select {
  width:100%;padding:7px 10px;border-radius:8px;font-size:12px;
  background:var(--bg-1);border:1px solid var(--line-1);color:var(--ink-0);
  appearance:none;outline:none;cursor:pointer;
  transition:border-color .14s;
}
.pvd-select:focus { border-color:var(--azure); }

/* 手动输入网格 */
.pvd-manual-grid { display:grid;grid-template-columns:1fr 1fr;gap:10px;padding:4px 12px 4px 35px; }
.pvd-manual-field { display:flex;flex-direction:column;gap:4px; }
.pvd-field-label { font-size:10.5px;font-weight:600;letter-spacing:.04em;text-transform:uppercase;color:var(--ink-2); }
.pvd-input-wrap { display:flex;align-items:center;border-radius:8px;border:1px solid var(--line-1);background:var(--bg-1);overflow:hidden;transition:border-color .14s; }
.pvd-input-wrap:focus-within { border-color:var(--azure); }
.pvd-input-prefix,.pvd-input-suffix { padding:0 7px;font-size:11.5px;color:var(--ink-2);background:var(--bg-2);border:none;flex-shrink:0;line-height:32px; }
.pvd-input-prefix { border-right:1px solid var(--line-0); }
.pvd-input-suffix { border-left:1px solid var(--line-0); }
.pvd-number-input {
  flex:1;padding:6px 8px;border:none;background:transparent;font-size:12px;
  font-variant-numeric:tabular-nums;color:var(--ink-0);outline:none;
  min-width:0;
}
.pvd-number-input::placeholder { color:var(--ink-2); }

/* 备注 */
.pvd-note-field { display:flex;flex-direction:column;gap:5px; }
.pvd-note-input {
  padding:7px 10px;border-radius:8px;border:1px solid var(--line-1);
  background:var(--bg-1);font-size:12px;color:var(--ink-0);outline:none;
  transition:border-color .14s;
}
.pvd-note-input:focus { border-color:var(--azure); }
.pvd-note-input::placeholder { color:var(--ink-2); }

/* 操作栏 */
.pvd-action-row { display:flex;align-items:center;gap:10px;justify-content:flex-end; }
.pvd-save-error { font-size:11.5px;color:var(--bad);margin-right:auto;flex:1;min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap; }
.pvd-save-ok { font-size:11.5px;color:var(--ok);margin-right:auto; }
.pvd-save-btn {
  display:inline-flex;align-items:center;gap:6px;padding:8px 18px;border-radius:10px;
  font-size:12.5px;font-weight:700;
  background:linear-gradient(180deg,var(--azure,#5CA8FF),#3A8EE8);
  border:1px solid rgba(92,168,255,.6);color:#fff;
  box-shadow:inset 0 1px 0 rgba(255,255,255,.18),0 2px 8px rgba(92,168,255,.22);
  cursor:pointer;transition:opacity .14s,box-shadow .14s;
}
.pvd-save-btn:hover:not(:disabled) { box-shadow:inset 0 1px 0 rgba(255,255,255,.22),0 3px 14px rgba(92,168,255,.35); }
.pvd-save-btn:disabled { opacity:.45;cursor:not-allowed; }
.pvd-restore-btn {
  display:inline-flex;align-items:center;gap:6px;padding:8px 16px;border-radius:10px;
  font-size:12.5px;font-weight:600;
  background:var(--metal-raised,linear-gradient(180deg,#272D37,#14171D));
  border:1px solid var(--bad,#E05A5A);color:var(--bad,#E05A5A);
  box-shadow:inset 0 1px 0 rgba(255,255,255,.04);
  cursor:pointer;transition:opacity .14s,background .14s;
}
.pvd-restore-btn:hover:not(:disabled) { background:linear-gradient(180deg,#2d1c1c,#1a1010); }
.pvd-restore-btn:disabled { opacity:.38;cursor:not-allowed; }
.pvd-action-ico { width:13px;height:13px; }
.pvd-saving-spin { animation:pvd-spin 1s linear infinite; }
@media(prefers-reduced-motion:reduce){.pvd-saving-spin{animation:none}}

/* 面板折叠动画 */
.pvd-panel-enter-active,.pvd-panel-leave-active { transition:opacity .2s,transform .22s cubic-bezier(.22,.68,0,1.1),max-height .25s ease; max-height:600px; overflow:hidden; }
.pvd-panel-enter-from,.pvd-panel-leave-to { opacity:0;transform:translateY(-6px);max-height:0; }
@media(prefers-reduced-motion:reduce){.pvd-panel-enter-active,.pvd-panel-leave-active{transition:none}}

.pvd-fade-enter-active,.pvd-fade-leave-active{transition:opacity .2s}
.pvd-fade-enter-from,.pvd-fade-leave-to{opacity:0}
.pvd-slide-enter-active,.pvd-slide-leave-active{transition:transform .28s cubic-bezier(.22,.68,0,1.2)}
.pvd-slide-enter-from,.pvd-slide-leave-to{transform:translateX(100%)}
@media(prefers-reduced-motion:reduce){.pvd-fade-enter-active,.pvd-fade-leave-active,.pvd-slide-enter-active,.pvd-slide-leave-active{transition:none}}
</style>
