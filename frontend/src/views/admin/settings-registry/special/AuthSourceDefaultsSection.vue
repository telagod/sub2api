<template>
  <div class="as-body">
    <!-- Require email on third-party signup toggle -->
    <div class="as-toggle-row as-toggle-row--top">
      <div>
        <label class="as-toggle-label">{{ t('admin.settings.authSourceDefaults.requireEmailLabel') }}</label>
        <p class="as-hint">{{ t('admin.settings.authSourceDefaults.requireEmailHint') }}</p>
      </div>
      <Toggle
        :model-value="localForceEmail"
        @update:model-value="onForceEmailChange"
      />
    </div>

    <!-- Auth source cards -->
    <div class="as-sources">
      <div
        v-for="meta in sourceMeta"
        :key="meta.source"
        class="as-source-card"
      >
        <!-- Card header: title + description + grant_on_signup toggle -->
        <div class="as-source-head">
          <div class="as-source-info">
            <div class="as-source-title">{{ meta.title }}</div>
            <p class="as-source-desc">{{ meta.description }}</p>
          </div>
          <Toggle
            v-model="localState[meta.source].grant_on_signup"
            :data-testid="`auth-source-${meta.source}-enabled`"
            @update:model-value="emitAll"
          />
        </div>

        <!-- Expanded panel (shown only when grant_on_signup = true) -->
        <div
          v-if="localState[meta.source].grant_on_signup"
          :data-testid="`auth-source-${meta.source}-panel`"
          class="as-source-panel"
        >
          <p class="as-hint as-hint--top">{{ t('admin.settings.authSourceDefaults.enabledHint') }}</p>

          <!-- Balance + Concurrency -->
          <div class="as-grid-2">
            <div class="as-field">
              <label class="as-field-label">{{ t('admin.settings.defaults.defaultBalance') }}</label>
              <input
                v-model.number="localState[meta.source].balance"
                type="number"
                step="0.01"
                min="0"
                class="as-input"
                placeholder="0.00"
                @change="emitAll"
              />
            </div>
            <div class="as-field">
              <label class="as-field-label">{{ t('admin.settings.defaults.defaultConcurrency') }}</label>
              <input
                v-model.number="localState[meta.source].concurrency"
                type="number"
                min="1"
                class="as-input"
                placeholder="5"
                @change="emitAll"
              />
            </div>
          </div>

          <!-- grant_on_first_bind -->
          <div class="as-toggle-row">
            <div>
              <label class="as-toggle-label">{{ t('admin.settings.authSourceDefaults.grantOnFirstBindLabel') }}</label>
              <p class="as-hint">{{ t('admin.settings.authSourceDefaults.grantOnFirstBindHint') }}</p>
            </div>
            <Toggle
              v-model="localState[meta.source].grant_on_first_bind"
              @update:model-value="emitAll"
            />
          </div>

          <!-- Default subscriptions -->
          <div class="as-sub-section">
            <div class="as-sub-header">
              <div>
                <label class="as-sub-label">{{ t('admin.settings.authSourceDefaults.defaultSubscriptionsLabel') }}</label>
                <p class="as-hint">{{ t('admin.settings.authSourceDefaults.defaultSubscriptionsHint') }}</p>
              </div>
              <button
                type="button"
                class="as-btn as-btn-sm"
                :disabled="subscriptionGroups.length === 0"
                @click="addSubscription(meta.source)"
              >
                {{ t('admin.settings.defaults.addDefaultSubscription') }}
              </button>
            </div>

            <div
              v-if="localState[meta.source].subscriptions.length === 0"
              class="as-empty"
            >
              {{ t('admin.settings.authSourceDefaults.noSourceSubscriptions') }}
            </div>

            <div v-else class="as-sub-list">
              <div
                v-for="(item, index) in localState[meta.source].subscriptions"
                :key="`${meta.source}-sub-${index}`"
                class="as-sub-row"
              >
                <div class="as-sub-group">
                  <label class="as-field-label">{{ t('admin.settings.defaults.subscriptionGroup') }}</label>
                  <Select
                    v-model="item.group_id"
                    class="default-sub-group-select"
                    :options="groupOptions"
                    :placeholder="t('admin.settings.defaults.subscriptionGroup')"
                    @update:model-value="emitAll"
                  >
                    <template #selected="{ option }">
                      <GroupBadge
                        v-if="option"
                        :name="(option as GroupOption).label"
                        :platform="(option as GroupOption).platform"
                        :subscription-type="(option as GroupOption).subscriptionType"
                        :rate-multiplier="(option as GroupOption).rate"
                      />
                      <span v-else class="as-muted">{{ t('admin.settings.defaults.subscriptionGroup') }}</span>
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
                <div class="as-sub-days">
                  <label class="as-field-label">{{ t('admin.settings.defaults.subscriptionValidityDays') }}</label>
                  <input
                    v-model.number="item.validity_days"
                    type="number"
                    min="1"
                    max="36500"
                    class="as-input as-input--h42"
                    @change="emitAll"
                  />
                </div>
                <div class="as-sub-del">
                  <button
                    type="button"
                    class="as-btn as-btn-danger"
                    @click="removeSubscription(meta.source, index)"
                  >
                    {{ t('common.delete') }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Platform quotas override matrix -->
          <div class="as-quota-section">
            <div class="as-quota-header">
              <label class="as-sub-label">{{ t('admin.settings.authSourceDefaults.platformQuotasOverride') }}</label>
              <p class="as-hint">{{ t('admin.settings.authSourceDefaults.platformQuotasOverrideHint') }}</p>
            </div>
            <div class="as-table-wrap">
              <table class="as-table">
                <thead>
                  <tr class="as-table-head">
                    <th class="as-th">{{ t('admin.settings.platformQuota.platform') }}</th>
                    <th class="as-th">{{ t('admin.settings.platformQuota.daily') }}</th>
                    <th class="as-th">{{ t('admin.settings.platformQuota.weekly') }}</th>
                    <th class="as-th">{{ t('admin.settings.platformQuota.monthly') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="p in PLATFORMS"
                    :key="`${meta.source}-pq-${p}`"
                    class="as-table-row"
                  >
                    <td class="as-td-platform">
                      <span class="as-mono">{{ p }}</span>
                    </td>
                    <td class="as-td-quota">
                      <input
                        v-model.number="localState[meta.source].platform_quotas[p]!.daily"
                        type="number"
                        step="0.01"
                        min="0"
                        class="as-input as-input--quota"
                        :placeholder="t('admin.settings.platformQuota.placeholder')"
                        @change="emitAll"
                      />
                    </td>
                    <td class="as-td-quota">
                      <input
                        v-model.number="localState[meta.source].platform_quotas[p]!.weekly"
                        type="number"
                        step="0.01"
                        min="0"
                        class="as-input as-input--quota"
                        :placeholder="t('admin.settings.platformQuota.placeholder')"
                        @change="emitAll"
                      />
                    </td>
                    <td class="as-td-quota">
                      <input
                        v-model.number="localState[meta.source].platform_quotas[p]!.monthly"
                        type="number"
                        step="0.01"
                        min="0"
                        class="as-input as-input--quota"
                        :placeholder="t('admin.settings.platformQuota.placeholder')"
                        @change="emitAll"
                      />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import Select from '@/components/common/Select.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
import { adminAPI } from '@/api'
import {
  buildAuthSourceDefaultsState,
  appendAuthSourceDefaultsToUpdateRequest,
  normalizePlatformQuotasMap,
} from '@/api/admin/settings'
import type {
  AuthSourceType,
  AuthSourceDefaultsState,
  DefaultSubscriptionSetting,
  PlatformType,
  DefaultPlatformQuotasMap,
} from '@/api/admin/settings'
import type { AdminGroup } from '@/types'

// ── Constants ──────────────────────────────────────────────────────────────
const PLATFORMS: PlatformType[] = ['anthropic', 'openai', 'gemini', 'antigravity']

// ── Props / emits ──────────────────────────────────────────────────────────
const props = defineProps<{
  settings: Record<string, unknown>
  formValues?: Record<string, unknown>
}>()

const emit = defineEmits<{
  'update:field': [key: string, value: unknown]
}>()

const { t, locale } = useI18n()

const isZhLocale = computed(() => locale.value.startsWith('zh'))
function localText(zh: string, en: string): string {
  return isZhLocale.value ? zh : en
}

// ── Auth source meta (title + description per source) ─────────────────────
const sourceMeta = computed<Array<{ source: AuthSourceType; title: string; description: string }>>(() => [
  {
    source: 'email',
    title: t('admin.settings.authSourceDefaults.sources.email.title'),
    description: t('admin.settings.authSourceDefaults.sources.email.description'),
  },
  {
    source: 'linuxdo',
    title: t('admin.settings.authSourceDefaults.sources.linuxdo.title'),
    description: t('admin.settings.authSourceDefaults.sources.linuxdo.description'),
  },
  {
    source: 'oidc',
    title: t('admin.settings.authSourceDefaults.sources.oidc.title'),
    description: t('admin.settings.authSourceDefaults.sources.oidc.description'),
  },
  {
    source: 'wechat',
    title: t('admin.settings.authSourceDefaults.sources.wechat.title'),
    description: t('admin.settings.authSourceDefaults.sources.wechat.description'),
  },
  {
    source: 'github',
    title: 'GitHub',
    description: localText(
      '通过 GitHub 已验证邮箱首次注册或首次绑定时应用。',
      'Applied on first signup or first bind through a verified GitHub email.',
    ),
  },
  {
    source: 'google',
    title: 'Google',
    description: localText(
      '通过 Google 已验证邮箱首次注册或首次绑定时应用。',
      'Applied on first signup or first bind through a verified Google email.',
    ),
  },
  {
    source: 'dingtalk',
    title: localText('钉钉', 'DingTalk'),
    description: localText(
      '通过钉钉首次注册或首次绑定时应用。',
      'Applied on first signup or first bind through DingTalk.',
    ),
  },
])

// ── Local state: force_email_on_third_party_signup ────────────────────────
// This flat boolean lives in the same card in SettingsView (line 3356)
const localForceEmail = ref<boolean>(false)

function onForceEmailChange(val: boolean) {
  localForceEmail.value = val
  emit('update:field', 'force_email_on_third_party_signup', val)
}

// ── Local reactive state (mirrors authSourceDefaults in SettingsView) ──────
const localState = reactive<AuthSourceDefaultsState>(
  buildAuthSourceDefaultsState({})
)

function syncFromSettings(s: Record<string, unknown>) {
  // force_email_on_third_party_signup
  localForceEmail.value = s['force_email_on_third_party_signup'] === true

  const built = buildAuthSourceDefaultsState(s)
  for (const source of Object.keys(built) as AuthSourceType[]) {
    const src = built[source]
    const dst = localState[source]
    dst.balance = src.balance
    dst.concurrency = src.concurrency
    dst.grant_on_signup = src.grant_on_signup
    dst.grant_on_first_bind = src.grant_on_first_bind
    // deep-replace subscriptions
    dst.subscriptions.splice(0, dst.subscriptions.length, ...src.subscriptions)
    // deep-replace platform_quotas (normalizePlatformQuotasMap ensures all 4 entries exist)
    const normalized = normalizePlatformQuotasMap(src.platform_quotas as DefaultPlatformQuotasMap | undefined)
    for (const p of PLATFORMS) {
      dst.platform_quotas[p] = { ...normalized[p]! }
    }
  }
}

// ── Subscription groups ────────────────────────────────────────────────────
const subscriptionGroups = reactive<AdminGroup[]>([])

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
  subscriptionGroups.map((g) => ({
    value: g.id,
    label: g.name,
    description: g.description ?? null,
    platform: g.platform,
    subscriptionType: g.subscription_type,
    rate: g.rate_multiplier,
  }))
)

// ── Lifecycle ──────────────────────────────────────────────────────────────
onMounted(async () => {
  // Load active subscription groups
  try {
    const all = await adminAPI.groups.getAll()
    const active = all.filter(
      (g) => g.subscription_type === 'subscription' && g.status === 'active'
    )
    subscriptionGroups.splice(0, subscriptionGroups.length, ...active)
  } catch {
    subscriptionGroups.splice(0, subscriptionGroups.length)
  }
  // Sync initial state from saved settings
  syncFromSettings(props.settings)
})

// Re-sync when parent discards changes (settings prop flips to saved snapshot)
watch(
  () => props.settings,
  (next) => syncFromSettings(next),
  { deep: true }
)

// ── Emit all flat auth-source keys via the API helper ─────────────────────
function emitAll() {
  const payload = appendAuthSourceDefaultsToUpdateRequest({}, localState) as Record<string, unknown>
  for (const [key, value] of Object.entries(payload)) {
    emit('update:field', key, value)
  }
}

// ── Subscription CRUD ──────────────────────────────────────────────────────
function findNextAvailableGroup(existingIds: number[]): AdminGroup | undefined {
  const set = new Set(existingIds)
  return subscriptionGroups.find((g) => !set.has(g.id))
}

function addSubscription(source: AuthSourceType) {
  if (subscriptionGroups.length === 0) return
  const candidate = findNextAvailableGroup(
    localState[source].subscriptions.map((s: DefaultSubscriptionSetting) => s.group_id)
  )
  if (!candidate) return
  localState[source].subscriptions.push({ group_id: candidate.id, validity_days: 30 })
  emitAll()
}

function removeSubscription(source: AuthSourceType, index: number) {
  localState[source].subscriptions.splice(index, 1)
  emitAll()
}
</script>

<style scoped>
/* QUENCH surface — consistent with UserDefaultsSection */
.as-body {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* ── Source list ───────────────────────────────────────────────────────── */
.as-sources { display: flex; flex-direction: column; gap: 12px; }

.as-source-card {
  border: 1px solid var(--line-1, #2F3540);
  border-radius: 10px;
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 0;
}

/* Card header */
.as-source-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.as-source-info { flex: 1; min-width: 0; }
.as-source-title { font-size: 13.5px; font-weight: 600; color: var(--ink-0, #E8EBF0); }
.as-source-desc  { font-size: 12px; color: var(--ink-2, #5C6470); margin: 3px 0 0; line-height: 1.5; }

/* Expanded panel */
.as-source-panel {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid var(--line-0, #20242C);
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* ── Typography helpers ────────────────────────────────────────────────── */
.as-hint       { font-size: 11.5px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 0; }
.as-hint--top  { margin-bottom: 4px; }
.as-field-label { font-size: 11.5px; font-weight: 500; color: var(--ink-1, #97A0AF); display: block; margin-bottom: 4px; }
.as-muted      { color: var(--ink-2, #5C6470); }
.as-mono       { font-family: var(--font-mono, ui-monospace, monospace); font-size: 12px; color: var(--ink-0, #E8EBF0); opacity: .85; }

/* ── Balance / concurrency grid ────────────────────────────────────────── */
.as-grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
@media (max-width: 540px) { .as-grid-2 { grid-template-columns: 1fr; } }

.as-field { display: flex; flex-direction: column; }

/* ── Toggle row ────────────────────────────────────────────────────────── */
.as-toggle-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  border: 1px solid var(--line-0, #20242C);
  border-radius: 8px;
  padding: 10px 12px;
}
.as-toggle-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--ink-0, #E8EBF0);
  display: block;
  margin-bottom: 3px;
}

/* ── Subscriptions ─────────────────────────────────────────────────────── */
.as-sub-section { display: flex; flex-direction: column; gap: 10px; }
.as-sub-label   { font-size: 13px; font-weight: 600; color: var(--ink-0, #E8EBF0); display: block; margin-bottom: 2px; }

.as-sub-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.as-empty {
  border: 1px dashed var(--line-1, #2F3540);
  border-radius: 8px;
  padding: 10px 14px;
  font-size: 12.5px;
  color: var(--ink-2, #5C6470);
}

.as-sub-list { display: flex; flex-direction: column; gap: 8px; }

.as-sub-row {
  display: grid;
  grid-template-columns: 1fr 160px auto;
  gap: 10px;
  align-items: end;
  border: 1px solid var(--line-0, #20242C);
  border-radius: 8px;
  padding: 10px 12px;
}
@media (max-width: 640px) { .as-sub-row { grid-template-columns: 1fr; } }

.as-sub-group { display: flex; flex-direction: column; }
.as-sub-days  { display: flex; flex-direction: column; }
.as-sub-del   { display: flex; align-items: flex-end; }

/* ── Platform quota matrix ─────────────────────────────────────────────── */
.as-quota-section { display: flex; flex-direction: column; gap: 10px; }
.as-quota-header  { display: flex; flex-direction: column; gap: 3px; }

.as-table-wrap { overflow-x: auto; }
.as-table { width: 100%; border-collapse: collapse; font-size: 12.5px; }
.as-table-head {}
.as-th {
  text-align: left; padding: 0 14px 8px 0;
  font-size: 11px; font-weight: 600; color: var(--ink-2, #5C6470);
  text-transform: uppercase; letter-spacing: .05em;
}
.as-table-row { vertical-align: top; }
.as-td-platform { padding: 4px 14px 4px 0; vertical-align: middle; }
.as-td-quota    { padding: 4px 14px 4px 0; }

/* ── Inputs ────────────────────────────────────────────────────────────── */
.as-input {
  width: 100%; padding: 7px 10px; border-radius: 7px;
  border: 1px solid var(--line-1, #2F3540); background: var(--bg-0, #0C0E12);
  color: var(--ink-0, #E8EBF0); font-size: 13px; font-family: inherit; outline: none;
  transition: border-color .15s, box-shadow .15s; box-sizing: border-box;
}
.as-input:focus { border-color: var(--azure, #5CA8FF); box-shadow: 0 0 0 3px rgba(92,168,255,.13); }
.as-input--h42  { height: 42px; }
.as-input--quota { height: 32px; width: 112px; font-size: 12.5px; padding: 4px 8px; }

/* ── Buttons ───────────────────────────────────────────────────────────── */
.as-btn {
  display: inline-flex; align-items: center; justify-content: center;
  padding: 6px 13px; border-radius: 7px; font-size: 12.5px; font-weight: 500;
  font-family: inherit; cursor: pointer; user-select: none;
  border: 1px solid var(--line-1, #2F3540); background: transparent;
  color: var(--ink-1, #97A0AF);
  transition: border-color .12s, color .12s, background .12s;
  white-space: nowrap;
}
.as-btn:hover:not(:disabled) { border-color: var(--line-0, #20242C); color: var(--ink-0, #E8EBF0); background: var(--bg-2, #171A20); }
.as-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.as-btn:disabled { opacity: .45; cursor: not-allowed; }

.as-btn-sm { padding: 4px 10px; font-size: 11.5px; flex-shrink: 0; }

.as-btn-danger { color: rgba(242, 92, 105, .8); width: 100%; }
.as-btn-danger:hover:not(:disabled) { color: var(--bad, #F25C69); background: rgba(242,92,105,.07); border-color: rgba(242,92,105,.3); }
</style>
