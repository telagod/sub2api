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
        </div>
        <div v-if="detail.description" class="pvd-desc-block">
          <p class="pvd-desc-text" :class="descExpanded ? '' : 'pvd-desc-clamp'">{{ detail.description }}</p>
          <button class="pvd-desc-more" @click="descExpanded = !descExpanded">{{ descExpanded ? t('admin.providerVerify.showLess') : t('admin.providerVerify.showMore') }}</button>
        </div>
        <div v-if="detail.baseline" class="pvd-baseline-note">
          <InfoIcon class="pvd-note-ico" />
          <span>{{ t('admin.providerVerify.baselineNote') }}</span>
          <span v-if="detail.baseline.source" class="pvd-note-source">{{ detail.baseline.source }}</span>
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
              <tr v-for="prov in sortedProviders" :key="prov.tag" class="pvd-tr" :class="isBaseline(prov) ? 'pvd-tr-bl' : ''">
                <td class="pvd-td">
                  <div class="pvd-prov-cell">
                    <span v-if="isBaseline(prov)" class="pvd-bl-edge" aria-hidden="true"></span>
                    <span class="pvd-prov-name">{{ prov.provider || prov.tag }}</span>
                    <span v-if="isBaseline(prov)" class="pvd-bl-badge">{{ t('admin.providerVerify.baselineBadge') }}</span>
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
        <div class="pvd-foot">
          <button class="pvd-edit-btn" disabled :title="t('admin.providerVerify.editDisabledTip')"><EditIcon class="pvd-edit-ico" />{{ t('admin.providerVerify.editBtn') }}</button>
          <span class="pvd-foot-hint">{{ t('admin.providerVerify.editDisabledTip') }}</span>
        </div>
      </div>
      <div v-else class="pvd-body pvd-empty-state"><PackageSearchIcon class="pvd-empty-ico" /><p class="pvd-empty-txt">{{ t('admin.providerVerify.noSlug') }}</p></div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { XIcon, LayersIcon, HashIcon, InfoIcon, RefreshCwIcon, EditIcon, PackageSearchIcon } from 'lucide-vue-next'
import modelCatalogAPI from '@/api/admin/modelCatalog'
import type { CatalogModelDetail, CatalogProvider } from '@/api/admin/modelCatalog'

const props = defineProps<{ open: boolean; slug: string | null; modelName?: string | null }>()
defineEmits<{ (e: 'close'): void }>()
const { t } = useI18n()

const detail = ref<CatalogModelDetail | null>(null)
const loading = ref(false)
const syncing = ref(false)
const error = ref<string | null>(null)
const descExpanded = ref(false)

watch(() => [props.open, props.slug] as const, ([nowOpen, nowSlug]) => {
  if (nowOpen && nowSlug) { void load(nowSlug) }
  else if (!nowOpen) { detail.value = null; error.value = null; descExpanded.value = false }
}, { immediate: true })

async function load(slug: string) {
  loading.value = true; error.value = null; detail.value = null
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
  } catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
  finally { loading.value = false }
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
const CAP_LABELS: Record<string, string> = { reasoning: 'Reasoning', tools: 'Tool Use', structured_outputs: 'Structured Out', vision: 'Vision', 'image-generation': 'Image Gen' }
const filteredCaps = computed(() => detail.value?.capabilities?.filter(c => SHOW_CAPS.includes(c)) ?? [])
function capLabel(c: string) { return CAP_LABELS[c] ?? c }
function isBaseline(p: CatalogProvider) { return p.tag === detail.value?.baseline?.source }

function fmtP(v?: number | null): string {
  if (v == null) return '—'
  const m = v * 1e6
  return `$${m.toFixed(m >= 1 ? 2 : m >= 0.1 ? 3 : 4)}`
}
function fmtCtx(n: number) { return n >= 1e6 ? `${(n/1e6).toFixed(1)}M` : n >= 1e3 ? `${(n/1e3).toFixed(0)}k` : String(n) }
// uptime_1d 已是 0-100 的百分数，直接显示，不再 ×100
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
.pvd-chip-ico { width:11px;height:11px; }
.pvd-desc-block { padding:12px 14px;background:var(--bg-2);border:1px solid var(--line-0);border-radius:10px; }
.pvd-desc-text { font-size:12px;line-height:1.65;color:var(--ink-1);margin:0; }
.pvd-desc-clamp { display:-webkit-box;-webkit-line-clamp:3;-webkit-box-orient:vertical;overflow:hidden; }
.pvd-desc-more { margin-top:6px;padding:0;border:none;background:none;font-size:11px;color:var(--azure);cursor:pointer;font-weight:600; }
.pvd-desc-more:hover { text-decoration:underline; }
.pvd-baseline-note { display:flex;align-items:center;gap:6px;padding:8px 12px;border-radius:8px;background:var(--azure-dim,rgba(92,168,255,.08));border:1px solid rgba(92,168,255,.2);font-size:11.5px;color:var(--ink-1); }
.pvd-note-ico { width:13px;height:13px;color:var(--azure);flex-shrink:0; }
.pvd-note-source { font-family:'IBM Plex Mono',monospace;font-size:10.5px;color:var(--azure);font-weight:600;padding:1px 5px;border-radius:4px;background:rgba(92,168,255,.12);border:1px solid rgba(92,168,255,.25); }
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
.pvd-td { padding:9px 10px;vertical-align:middle; }
.pvd-td-r { text-align:right; }.pvd-td-c { text-align:center; }
.pvd-prov-cell { display:flex;align-items:center;gap:7px; }
.pvd-bl-edge { display:inline-block;width:3px;height:20px;border-radius:2px;background:var(--azure);flex-shrink:0; }
.pvd-prov-name { font-size:12px;font-weight:500;color:var(--ink-0); }
.pvd-bl-badge { font-size:9px;font-weight:700;letter-spacing:.06em;text-transform:uppercase;padding:1px 5px;border-radius:4px;background:var(--azure-dim,rgba(92,168,255,.15));border:1px solid rgba(92,168,255,.3);color:var(--azure); }
.pvd-price { font-size:11.5px;font-variant-numeric:tabular-nums; }
.pvd-muted { color:var(--ink-2)!important; }
.pvd-uptime { font-size:11px;font-variant-numeric:tabular-nums; }
.pvd-up-ok { color:var(--ok); }.pvd-up-warn { color:var(--warm,#F4A64A); }.pvd-up-muted { color:var(--ink-2); }
.pvd-quant { display:inline-block;padding:1px 5px;border-radius:4px;font-size:9.5px;font-weight:600;letter-spacing:.04em;background:var(--bg-2);border:1px solid var(--line-1);color:var(--ink-2); }
.pvd-empty-prov,.pvd-empty-state { display:flex;flex-direction:column;align-items:center;justify-content:center;gap:8px;padding:36px 16px;background:var(--bg-2); }
.pvd-empty-ico { width:32px;height:32px;color:var(--ink-2);opacity:.3; }
.pvd-empty-txt { font-size:12px;color:var(--ink-2);margin:0;text-align:center; }
.pvd-foot { display:flex;align-items:center;gap:10px;padding-top:10px;border-top:1px solid var(--line-0);margin-top:auto; }
.pvd-edit-btn { display:inline-flex;align-items:center;gap:6px;padding:7px 15px;border-radius:10px;font-size:12.5px;font-weight:600;background:var(--metal-raised,linear-gradient(180deg,#272D37,#14171D));border:1px solid var(--line-1);color:var(--ink-2);box-shadow:inset 0 1px 0 rgba(255,255,255,.04);cursor:not-allowed;opacity:.45; }
.pvd-edit-ico { width:13px;height:13px; }
.pvd-foot-hint { font-size:10.5px;color:var(--ink-2); }
.pvd-fade-enter-active,.pvd-fade-leave-active{transition:opacity .2s}
.pvd-fade-enter-from,.pvd-fade-leave-to{opacity:0}
.pvd-slide-enter-active,.pvd-slide-leave-active{transition:transform .28s cubic-bezier(.22,.68,0,1.2)}
.pvd-slide-enter-from,.pvd-slide-leave-to{transform:translateX(100%)}
@media(prefers-reduced-motion:reduce){.pvd-fade-enter-active,.pvd-fade-leave-active,.pvd-slide-enter-active,.pvd-slide-leave-active{transition:none}}
</style>
