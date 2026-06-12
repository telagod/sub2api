<template>
  <div class="ud-body">
    <!-- ══ Block 1: default_subscriptions ══ -->
    <div class="ud-block">
      <div class="ud-block-header">
        <div>
          <label class="ud-label">{{ t('admin.settings.defaults.defaultSubscriptions') }}</label>
          <p class="ud-hint">{{ t('admin.settings.defaults.defaultSubscriptionsHint') }}</p>
        </div>
        <button
          type="button"
          class="ud-btn ud-btn-secondary ud-btn-sm"
          :disabled="subscriptionGroups.length === 0"
          @click="addDefaultSubscription"
        >
          {{ t('admin.settings.defaults.addDefaultSubscription') }}
        </button>
      </div>

      <div
        v-if="localSubscriptions.length === 0"
        class="ud-empty"
      >
        {{ t('admin.settings.defaults.defaultSubscriptionsEmpty') }}
      </div>

      <div v-else class="ud-sub-list">
        <div
          v-for="(item, index) in localSubscriptions"
          :key="`default-sub-${index}`"
          class="ud-sub-row"
        >
          <!-- Group selector -->
          <div class="ud-sub-group">
            <label class="ud-field-label">{{ t('admin.settings.defaults.subscriptionGroup') }}</label>
            <Select
              v-model="item.group_id"
              class="default-sub-group-select"
              :options="groupOptions"
              :placeholder="t('admin.settings.defaults.subscriptionGroup')"
              @update:model-value="emitSubscriptions"
            >
              <template #selected="{ option }">
                <GroupBadge
                  v-if="option"
                  :name="(option as GroupOption).label"
                  :platform="(option as GroupOption).platform"
                  :subscription-type="(option as GroupOption).subscriptionType"
                  :rate-multiplier="(option as GroupOption).rate"
                />
                <span v-else class="ud-muted">{{ t('admin.settings.defaults.subscriptionGroup') }}</span>
              </template>
              <template #option="{ option, selected }">
                <GroupOptionItem
                  :name="(option as GroupOption).label"
                  :platform="(option as GroupOption).platform"
                  :subscription-type="(option as GroupOption).subscriptionType"
                  :rate-multiplier="(option as GroupOption).rate"
                  :description="(option as GroupOption).description"
                  :selected="selected"
                />
              </template>
            </Select>
          </div>

          <!-- Validity days -->
          <div class="ud-sub-days">
            <label class="ud-field-label">{{ t('admin.settings.defaults.subscriptionValidityDays') }}</label>
            <input
              v-model.number="item.validity_days"
              type="number"
              min="1"
              max="36500"
              class="ud-input ud-input--h42"
              @change="emitSubscriptions"
            />
          </div>

          <!-- Delete -->
          <div class="ud-sub-del">
            <button
              type="button"
              class="ud-btn ud-btn-secondary default-sub-delete-btn ud-danger-text"
              @click="removeDefaultSubscription(index)"
            >
              {{ t('common.delete') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ══ Block 2: default_platform_quotas matrix ══ -->
    <div class="ud-block ud-block-sep">
      <div class="ud-block-header ud-block-header--col">
        <label class="ud-label">{{ t('admin.settings.defaults.defaultPlatformQuotas') }}</label>
        <p class="ud-hint">{{ t('admin.settings.defaults.defaultPlatformQuotasHint') }}</p>
        <p class="ud-warn">{{ t('admin.settings.defaults.platformQuotaNotice') }}</p>
      </div>

      <div class="ud-table-wrap">
        <table class="ud-table">
          <thead>
            <tr class="ud-table-head">
              <th class="ud-th">{{ t('admin.settings.platformQuota.platform') }}</th>
              <th class="ud-th">{{ t('admin.settings.platformQuota.daily') }}</th>
              <th class="ud-th">{{ t('admin.settings.platformQuota.weekly') }}</th>
              <th class="ud-th">{{ t('admin.settings.platformQuota.monthly') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="p in PLATFORMS"
              :key="p"
              class="ud-table-row"
            >
              <td class="ud-td-platform">
                <span class="ud-mono">{{ p }}</span>
              </td>
              <td class="ud-td-quota">
                <input
                  v-model.number="localQuotas[p]!.daily"
                  type="number"
                  step="0.01"
                  min="0"
                  class="ud-input ud-input--quota"
                  :placeholder="t('admin.settings.platformQuota.placeholder')"
                  @change="emitQuotas"
                />
              </td>
              <td class="ud-td-quota">
                <input
                  v-model.number="localQuotas[p]!.weekly"
                  type="number"
                  step="0.01"
                  min="0"
                  class="ud-input ud-input--quota"
                  :placeholder="t('admin.settings.platformQuota.placeholder')"
                  @change="emitQuotas"
                />
              </td>
              <td class="ud-td-quota">
                <input
                  v-model.number="localQuotas[p]!.monthly"
                  type="number"
                  step="0.01"
                  min="0"
                  class="ud-input ud-input--quota"
                  :placeholder="t('admin.settings.platformQuota.placeholder')"
                  @change="emitQuotas"
                />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
import { adminAPI } from '@/api'
import {
  sanitizePlatformQuotasMap,
  normalizePlatformQuotasMap,
  normalizeDefaultSubscriptionSettings,
} from '@/api/admin/settings'
import type {
  DefaultSubscriptionSetting,
  DefaultPlatformQuotasMap,
  PlatformType,
} from '@/api/admin/settings'
import type { AdminGroup } from '@/types'

const PLATFORMS: PlatformType[] = ['anthropic', 'openai', 'gemini', 'antigravity']

// ── Component interface ────────────────────────────────────────────────────
const props = defineProps<{
  settings: Record<string, unknown>
  formValues?: Record<string, unknown>
}>()

const emit = defineEmits<{
  'update:field': [key: string, value: unknown]
}>()

const { t } = useI18n()

// ── Local state: subscriptions ─────────────────────────────────────────────
const localSubscriptions = ref<DefaultSubscriptionSetting[]>([])

// ── Local state: platform quotas ───────────────────────────────────────────
// Fully-normalized 4×3 object, always non-null for template binding
const localQuotas = reactive<Record<PlatformType, { daily: number | null; weekly: number | null; monthly: number | null }>>(
  normalizePlatformQuotasMap() as Record<PlatformType, { daily: number | null; weekly: number | null; monthly: number | null }>
)

// ── Groups (loaded once on mount) ──────────────────────────────────────────
const subscriptionGroups = ref<AdminGroup[]>([])

interface GroupOption {
  value: number
  label: string
  description: string | null
  platform: AdminGroup['platform']
  subscriptionType: AdminGroup['subscription_type']
  rate: number
  [key: string]: unknown
}

const groupOptions = computed<GroupOption[]>(() =>
  subscriptionGroups.value.map((g) => ({
    value: g.id,
    label: g.name,
    description: g.description ?? null,
    platform: g.platform,
    subscriptionType: g.subscription_type,
    rate: g.rate_multiplier,
  }))
)

onMounted(async () => {
  // Load groups
  try {
    const all = await adminAPI.groups.getAll()
    subscriptionGroups.value = all.filter(
      (g) => g.subscription_type === 'subscription' && g.status === 'active'
    )
  } catch {
    subscriptionGroups.value = []
  }
  // Sync from props.settings (initial load)
  syncFromSettings(props.settings)
})

// ── Sync on settings prop change (global discard / reload) ─────────────────
watch(
  () => props.settings,
  (next) => syncFromSettings(next),
  { deep: true }
)

function syncFromSettings(s: Record<string, unknown>) {
  // subscriptions
  const rawSubs = s['default_subscriptions']
  localSubscriptions.value = normalizeDefaultSubscriptionSettings(
    Array.isArray(rawSubs) ? (rawSubs as DefaultSubscriptionSetting[]) : []
  )

  // platform quotas
  const rawQuotas = s['default_platform_quotas'] as DefaultPlatformQuotasMap | undefined
  const normalized = normalizePlatformQuotasMap(rawQuotas) as Record<
    PlatformType,
    { daily: number | null; weekly: number | null; monthly: number | null }
  >
  for (const p of PLATFORMS) {
    localQuotas[p] = { ...normalized[p]! }
  }
}

// ── Emit helpers ───────────────────────────────────────────────────────────
function emitSubscriptions() {
  emit('update:field', 'default_subscriptions', normalizeDefaultSubscriptionSettings(localSubscriptions.value))
}

function emitQuotas() {
  emit('update:field', 'default_platform_quotas', sanitizePlatformQuotasMap(localQuotas as DefaultPlatformQuotasMap))
}

// ── Add / remove subscriptions ─────────────────────────────────────────────
function findNextAvailableGroup(existingIds: number[]): AdminGroup | undefined {
  const set = new Set(existingIds)
  return subscriptionGroups.value.find((g) => !set.has(g.id))
}

function addDefaultSubscription() {
  if (subscriptionGroups.value.length === 0) return
  const candidate = findNextAvailableGroup(localSubscriptions.value.map((s) => s.group_id))
  if (!candidate) return
  localSubscriptions.value.push({ group_id: candidate.id, validity_days: 30 })
  emitSubscriptions()
}

function removeDefaultSubscription(index: number) {
  localSubscriptions.value.splice(index, 1)
  emitSubscriptions()
}
</script>

<style scoped>
/* QUENCH surface — consistent with RectifierSection / OpenaiFastPolicySection */
.ud-body {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 0;
}

/* ── Block wrapper ─────────────────────────────────────────────────────── */
.ud-block { display: flex; flex-direction: column; gap: 12px; }
.ud-block-sep { border-top: 1px solid var(--line-0, #20242C); padding-top: 20px; margin-top: 20px; }

.ud-block-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}
.ud-block-header--col {
  flex-direction: column;
  align-items: flex-start;
}

/* ── Typography ────────────────────────────────────────────────────────── */
.ud-label  { font-size: 13px; font-weight: 600; color: var(--ink-0, #E8EBF0); display: block; margin-bottom: 2px; }
.ud-hint   { font-size: 11.5px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 0; }
.ud-warn   { font-size: 11px; color: var(--amber, #FBBF24); margin: 4px 0 0; }
.ud-muted  { color: var(--ink-2, #5C6470); }
.ud-mono   { font-family: var(--font-mono, ui-monospace, monospace); font-size: 12px; color: var(--ink-0, #E8EBF0); opacity: .85; }

.ud-field-label { font-size: 11px; font-weight: 500; color: var(--ink-2, #5C6470); display: block; margin-bottom: 4px; }

/* ── Empty placeholder ─────────────────────────────────────────────────── */
.ud-empty {
  border: 1px dashed var(--line-1, #2F3540);
  border-radius: 8px;
  padding: 12px 16px;
  font-size: 12.5px;
  color: var(--ink-2, #5C6470);
}

/* ── Subscription list ─────────────────────────────────────────────────── */
.ud-sub-list { display: flex; flex-direction: column; gap: 8px; }

.ud-sub-row {
  display: grid;
  grid-template-columns: 1fr 160px auto;
  gap: 10px;
  align-items: end;
  border: 1px solid var(--line-0, #20242C);
  border-radius: 8px;
  padding: 10px 12px;
}
@media (max-width: 640px) {
  .ud-sub-row { grid-template-columns: 1fr; }
}

.ud-sub-group { display: flex; flex-direction: column; }
.ud-sub-days  { display: flex; flex-direction: column; }
.ud-sub-del   { display: flex; align-items: flex-end; }

/* ── Inputs ────────────────────────────────────────────────────────────── */
.ud-input {
  width: 100%; padding: 7px 10px; border-radius: 7px;
  border: 1px solid var(--line-1, #2F3540); background: var(--bg-0, #0C0E12);
  color: var(--ink-0, #E8EBF0); font-size: 13px; font-family: inherit; outline: none;
  transition: border-color .15s, box-shadow .15s; box-sizing: border-box;
}
.ud-input:focus { border-color: var(--azure, #5CA8FF); box-shadow: 0 0 0 3px rgba(92,168,255,.13); }
.ud-input--h42  { height: 42px; }
.ud-input--quota { height: 32px; width: 112px; font-size: 12.5px; padding: 4px 8px; }

/* ── Buttons ───────────────────────────────────────────────────────────── */
.ud-btn {
  display: inline-flex; align-items: center; justify-content: center;
  padding: 6px 13px; border-radius: 7px; font-size: 12.5px; font-weight: 500;
  font-family: inherit; cursor: pointer; user-select: none;
  border: 1px solid var(--line-1, #2F3540); background: transparent;
  color: var(--ink-1, #97A0AF);
  transition: border-color .12s, color .12s, background .12s;
}
.ud-btn:hover:not(:disabled) { border-color: var(--line-0, #20242C); color: var(--ink-0, #E8EBF0); background: var(--bg-2, #171A20); }
.ud-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.ud-btn:disabled { opacity: .45; cursor: not-allowed; }

.ud-btn-secondary { /* inherits ud-btn */ }
.ud-btn-sm { padding: 4px 10px; font-size: 11.5px; flex-shrink: 0; }

.ud-danger-text { color: rgba(242, 92, 105, .8); width: 100%; }
.ud-danger-text:hover:not(:disabled) { color: var(--bad, #F25C69); background: rgba(242,92,105,.07); border-color: rgba(242,92,105,.3); }

/* ── Platform quota table ──────────────────────────────────────────────── */
.ud-table-wrap { overflow-x: auto; }

.ud-table { width: 100%; border-collapse: collapse; font-size: 12.5px; }

.ud-table-head {}
.ud-th {
  text-align: left; padding: 0 14px 8px 0;
  font-size: 11px; font-weight: 600; color: var(--ink-2, #5C6470);
  text-transform: uppercase; letter-spacing: .05em;
}
.ud-table-row { vertical-align: top; }
.ud-td-platform { padding: 4px 14px 4px 0; vertical-align: middle; }
.ud-td-quota    { padding: 4px 14px 4px 0; }
</style>
