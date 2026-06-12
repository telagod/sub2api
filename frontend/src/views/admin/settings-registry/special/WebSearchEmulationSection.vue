<template>
  <div class="wse-body">
    <!-- Global Enable Toggle -->
    <div class="wse-row-switch">
      <div>
        <label class="wse-label">{{ t('admin.settings.webSearchEmulation.enabled') }}</label>
        <p class="wse-hint-inline">{{ t('admin.settings.webSearchEmulation.enabledHint') }}</p>
      </div>
      <Toggle v-model="config.enabled" />
    </div>

    <!-- Providers List -->
    <div v-if="config.enabled" class="wse-providers">
      <div class="wse-providers-header">
        <label class="wse-label">{{ t('admin.settings.webSearchEmulation.providers') }}</label>
        <button type="button" class="wse-btn-secondary" @click="addProvider">
          {{ t('admin.settings.webSearchEmulation.addProvider') }}
        </button>
      </div>

      <div
        v-if="config.providers.length === 0"
        class="wse-empty"
      >
        {{ t('admin.settings.webSearchEmulation.noProviders') }}
      </div>

      <div
        v-for="(provider, pIdx) in config.providers"
        :key="pIdx"
        class="wse-provider-card"
      >
        <!-- Collapsible header -->
        <div class="wse-provider-header" @click="toggleExpand(pIdx)">
          <div class="wse-provider-header-left">
            <svg
              class="wse-chevron"
              :class="{ 'wse-chevron--open': expandedProviders[pIdx] }"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
            <Select
              v-model="provider.type"
              :options="[
                { value: 'brave', label: 'Brave Search' },
                { value: 'tavily', label: 'Tavily' },
              ]"
              class="w-36"
              @click.stop
            />
            <span class="wse-quota-summary">
              {{ provider.quota_used ?? 0 }} /
              {{ provider.quota_limit != null && provider.quota_limit > 0 ? provider.quota_limit : '∞' }}
            </span>
            <span
              v-if="!expandedProviders[pIdx] && provider.api_key_configured"
              class="wse-key-configured"
            >
              {{ t('admin.settings.webSearchEmulation.apiKeyConfigured') }}
            </span>
          </div>
          <button
            type="button"
            class="wse-remove-btn"
            @click.stop="removeProvider(pIdx)"
          >
            {{ t('admin.settings.webSearchEmulation.removeProvider') }}
          </button>
        </div>

        <!-- Expanded Content -->
        <div v-if="expandedProviders[pIdx]" class="wse-provider-body">
          <!-- API Key -->
          <div>
            <label class="wse-field-label">{{ t('admin.settings.webSearchEmulation.apiKey') }}</label>
            <div class="wse-input-wrap">
              <input
                v-model="provider.api_key"
                :type="apiKeyVisible[pIdx] ? 'text' : 'password'"
                class="wse-input"
                :class="provider.api_key || provider.api_key_configured ? 'wse-input--with-actions' : ''"
                :placeholder="
                  provider.api_key_configured
                    ? '••••••••'
                    : t('admin.settings.webSearchEmulation.apiKeyPlaceholder')
                "
              />
              <div
                v-if="provider.api_key || provider.api_key_configured"
                class="wse-input-actions"
              >
                <button
                  type="button"
                  class="wse-icon-btn"
                  :title="apiKeyVisible[pIdx]
                    ? t('admin.settings.webSearchEmulation.hideApiKey')
                    : t('admin.settings.webSearchEmulation.showApiKey')"
                  @click="apiKeyVisible[pIdx] = !apiKeyVisible[pIdx]"
                >
                  <svg v-if="!apiKeyVisible[pIdx]" class="wse-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                  <svg v-else class="wse-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L3 3m6.878 6.878L21 21" />
                  </svg>
                </button>
                <button
                  type="button"
                  class="wse-icon-btn"
                  :class="{ 'wse-icon-btn--disabled': !provider.api_key }"
                  :title="t('admin.settings.webSearchEmulation.copyApiKey')"
                  :disabled="!provider.api_key"
                  @click="copyApiKey(pIdx)"
                >
                  <svg class="wse-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <!-- Quota + Subscription -->
          <div class="wse-grid-2">
            <div>
              <label class="wse-field-label">{{ t('admin.settings.webSearchEmulation.quotaLimit') }}</label>
              <input
                v-model="provider.quota_limit"
                type="number"
                min="1"
                class="wse-input wse-input--mono"
                placeholder="∞"
              />
              <p class="wse-field-hint">{{ t('admin.settings.webSearchEmulation.quotaLimitHint') }}</p>
            </div>
            <div>
              <label class="wse-field-label">{{ t('admin.settings.webSearchEmulation.subscribedAt') }}</label>
              <input
                :value="formatSubscribedAt(provider.subscribed_at)"
                type="date"
                class="wse-input"
                @input="provider.subscribed_at = parseSubscribedAt(($event.target as HTMLInputElement).value)"
              />
              <p class="wse-field-hint">{{ t('admin.settings.webSearchEmulation.subscribedAtHint') }}</p>
            </div>
          </div>

          <!-- Quota Usage Bar -->
          <div class="wse-usage-row">
            <span class="wse-field-label wse-no-margin">{{ t('admin.settings.webSearchEmulation.quotaUsage') }}:</span>
            <div
              v-if="provider.quota_limit != null && provider.quota_limit > 0"
              class="wse-quota-bar"
            >
              <div
                class="wse-quota-fill"
                :class="
                  quotaPercentage(provider) > 90
                    ? 'wse-quota-fill--danger'
                    : quotaPercentage(provider) > 70
                      ? 'wse-quota-fill--warn'
                      : 'wse-quota-fill--ok'
                "
                :style="{ width: Math.min(quotaPercentage(provider), 100) + '%' }"
              />
            </div>
            <div v-else class="wse-quota-bar-spacer" />
            <span class="wse-field-label wse-no-margin">
              {{ provider.quota_used ?? 0 }} /
              {{ provider.quota_limit != null && provider.quota_limit > 0 ? provider.quota_limit : '∞' }}
            </span>
            <button
              v-if="(provider.quota_used ?? 0) > 0"
              type="button"
              class="wse-reset-btn"
              @click="resetUsage(pIdx)"
            >
              {{ t('admin.settings.webSearchEmulation.resetUsage') }}
            </button>
          </div>

          <!-- Proxy + Test -->
          <div class="wse-proxy-test-row">
            <div class="wse-proxy-wrap">
              <label class="wse-field-label">{{ t('admin.settings.webSearchEmulation.proxy') }}</label>
              <ProxySelector v-model="provider.proxy_id" :proxies="proxies" />
            </div>
            <button type="button" class="wse-btn-secondary wse-btn-test" @click="openTestDialog">
              {{ t('admin.settings.webSearchEmulation.test') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Save Button -->
    <div class="wse-footer">
      <button
        type="button"
        class="wse-btn-save"
        :disabled="saving"
        @click="save"
      >
        <svg v-if="saving" class="wse-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
        {{ saving ? t('common.saving') : t('common.save') }}
      </button>
    </div>

    <!-- Test Dialog -->
    <div
      v-if="testDialogOpen"
      class="wse-dialog-overlay"
      @click.self="testDialogOpen = false"
    >
      <div class="wse-dialog">
        <h3 class="wse-dialog-title">{{ t('admin.settings.webSearchEmulation.testResultTitle') }}</h3>
        <div class="wse-dialog-search">
          <input
            v-model="testQuery"
            type="text"
            class="wse-input wse-input--flex"
            :placeholder="t('admin.settings.webSearchEmulation.testDefaultQuery')"
            @keyup.enter="runTest"
          />
          <button
            type="button"
            class="wse-btn-primary"
            :disabled="testLoading"
            @click="runTest"
          >
            {{ testLoading
              ? t('admin.settings.webSearchEmulation.testing')
              : t('admin.settings.webSearchEmulation.test') }}
          </button>
        </div>
        <div v-if="testResult" class="wse-test-results">
          <p class="wse-test-provider">
            {{ t('admin.settings.webSearchEmulation.testResultProvider') }}: {{ testResult.provider }}
          </p>
          <div v-if="testResult.results.length === 0" class="wse-test-empty">
            {{ t('admin.settings.webSearchEmulation.testNoResults') }}
          </div>
          <div
            v-for="(r, rIdx) in testResult.results"
            :key="rIdx"
            class="wse-test-result-item"
          >
            <a :href="r.url" target="_blank" class="wse-test-result-link">{{ r.title }}</a>
            <p class="wse-test-result-snippet">{{ r.snippet }}</p>
          </div>
        </div>
        <div class="wse-dialog-footer">
          <button type="button" class="wse-btn-secondary" @click="testDialogOpen = false">
            {{ t('common.close') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import Select from '@/components/common/Select.vue'
import ProxySelector from '@/components/common/ProxySelector.vue'
import adminAPI from '@/api/admin'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'
import type {
  WebSearchEmulationConfig,
  WebSearchProviderConfig,
  WebSearchTestResult,
} from '@/api/admin/settings'
import type { Proxy } from '@/types'

// Self-contained: ignores parent settings/form props — has own GET/PUT cycle
defineProps<{
  settings?: Record<string, unknown>
  formValues?: Record<string, unknown>
}>()

const { t } = useI18n()
const appStore = useAppStore()

const DEFAULT_QUOTA_LIMIT = 1000

// ── State ──────────────────────────────────────────────────────────────────────

const config = reactive<WebSearchEmulationConfig>({
  enabled: false,
  providers: [],
})

const proxies = ref<Proxy[]>([])
const saving = ref(false)

const expandedProviders = reactive<Record<number, boolean>>({})
const apiKeyVisible = reactive<Record<number, boolean>>({})

const testDialogOpen = ref(false)
const testQuery = ref('')
const testLoading = ref(false)
const testResult = ref<WebSearchTestResult | null>(null)

// ── Lifecycle ──────────────────────────────────────────────────────────────────

onMounted(async () => {
  try {
    const [resp, proxiesResp] = await Promise.all([
      adminAPI.settings.getWebSearchEmulationConfig(),
      adminAPI.proxies.list().catch(() => ({ items: [] as Proxy[] })),
    ])
    if (resp) {
      config.enabled = resp.enabled || false
      config.providers = resp.providers || []
    }
    proxies.value = proxiesResp.items || []
  } catch (err: unknown) {
    const status = (err as { status?: number })?.status
    if (status !== 404 && status !== undefined) {
      appStore.showError(extractApiErrorMessage(err, t('common.error')))
    }
  }
})

// ── Helpers ────────────────────────────────────────────────────────────────────

function formatSubscribedAt(ts: number | null): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  const y = d.getUTCFullYear()
  const m = String(d.getUTCMonth() + 1).padStart(2, '0')
  const day = String(d.getUTCDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function parseSubscribedAt(dateStr: string): number | null {
  if (!dateStr) return null
  return Math.floor(new Date(dateStr + 'T00:00:00Z').getTime() / 1000)
}

function quotaPercentage(provider: WebSearchProviderConfig): number {
  if (!provider.quota_limit || provider.quota_limit <= 0) return 0
  return ((provider.quota_used ?? 0) / provider.quota_limit) * 100
}

// ── Provider CRUD ──────────────────────────────────────────────────────────────

function addProvider() {
  const idx = config.providers.length
  config.providers.push({
    type: 'brave',
    api_key: '',
    api_key_configured: false,
    quota_limit: DEFAULT_QUOTA_LIMIT,
    subscribed_at: null,
    proxy_id: null,
    expires_at: null,
  } as WebSearchProviderConfig)
  expandedProviders[idx] = true
}

function removeProvider(idx: number) {
  config.providers.splice(idx, 1)
  const newExpanded: Record<number, boolean> = {}
  const newVisible: Record<number, boolean> = {}
  for (let i = 0; i < config.providers.length; i++) {
    const oldIdx = i >= idx ? i + 1 : i
    newExpanded[i] = expandedProviders[oldIdx] ?? false
    newVisible[i] = apiKeyVisible[oldIdx] ?? false
  }
  Object.keys(expandedProviders).forEach((k) => delete expandedProviders[Number(k)])
  Object.keys(apiKeyVisible).forEach((k) => delete apiKeyVisible[Number(k)])
  Object.assign(expandedProviders, newExpanded)
  Object.assign(apiKeyVisible, newVisible)
}

function toggleExpand(idx: number) {
  expandedProviders[idx] = !expandedProviders[idx]
}

// ── API Key actions ────────────────────────────────────────────────────────────

async function copyApiKey(idx: number) {
  const key = config.providers[idx]?.api_key
  if (!key) return
  try {
    await navigator.clipboard.writeText(key)
    appStore.showSuccess(t('admin.settings.webSearchEmulation.copied'))
  } catch {
    appStore.showError(t('common.error'))
  }
}

// ── Usage reset ────────────────────────────────────────────────────────────────

async function resetUsage(idx: number) {
  const provider = config.providers[idx]
  if (!provider) return
  if (!confirm(t('admin.settings.webSearchEmulation.resetUsageConfirm'))) return
  try {
    await adminAPI.settings.resetWebSearchUsage({ provider_type: provider.type })
    provider.quota_used = 0
    appStore.showSuccess(t('admin.settings.webSearchEmulation.resetUsageSuccess'))
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  }
}

// ── Save ───────────────────────────────────────────────────────────────────────

async function save() {
  for (const p of config.providers) {
    const raw = p.quota_limit
    if (raw != null && Number(raw) !== 0 && Number(raw) < 1) {
      appStore.showError(t('admin.settings.webSearchEmulation.quotaLimitMustBePositive'))
      return
    }
  }
  saving.value = true
  try {
    const providers = config.providers.map((p: WebSearchProviderConfig) => ({
      ...p,
      quota_limit: Number(p.quota_limit) > 0 ? Number(p.quota_limit) : null,
    }))
    await adminAPI.settings.updateWebSearchEmulationConfig({
      enabled: config.enabled,
      providers,
    })
    appStore.showSuccess(t('common.saved'))
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    saving.value = false
  }
}

// ── Test Dialog ────────────────────────────────────────────────────────────────

function openTestDialog() {
  testResult.value = null
  testDialogOpen.value = true
}

async function runTest() {
  testLoading.value = true
  testResult.value = null
  try {
    const query = testQuery.value.trim() || t('admin.settings.webSearchEmulation.testDefaultQuery')
    testResult.value = await adminAPI.settings.testWebSearchEmulation(query)
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    testLoading.value = false
  }
}
</script>

<style scoped>
.wse-body {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

/* Row: switch */
.wse-row-switch {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.wse-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--ink-0, #E8EBF0);
  display: block;
}

.wse-hint-inline {
  font-size: 11.5px;
  color: var(--ink-2, #5C6470);
  margin: 2px 0 0;
  line-height: 1.45;
}

/* Providers list */
.wse-providers {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.wse-providers-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.wse-empty {
  border: 1px dashed var(--line-1, #2F3540);
  border-radius: 8px;
  padding: 14px;
  text-align: center;
  font-size: 13px;
  color: var(--ink-2, #5C6470);
}

.wse-provider-card {
  border: 1px solid var(--line-1, #2F3540);
  border-radius: 10px;
  overflow: hidden;
}

.wse-provider-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  cursor: pointer;
  user-select: none;
}

.wse-provider-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.wse-chevron {
  width: 16px;
  height: 16px;
  color: var(--ink-2, #5C6470);
  transition: transform 0.15s;
  flex-shrink: 0;
}

.wse-chevron--open {
  transform: rotate(90deg);
}

.wse-quota-summary {
  font-size: 12px;
  color: var(--ink-2, #5C6470);
  font-variant-numeric: tabular-nums;
}

.wse-key-configured {
  font-size: 12px;
  color: #34d399;
}

.wse-remove-btn {
  font-size: 12px;
  color: #f87171;
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
  transition: color 0.12s;
}

.wse-remove-btn:hover {
  color: #fca5a5;
}

/* Provider body */
.wse-provider-body {
  padding: 14px;
  border-top: 1px solid var(--line-1, #2F3540);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* Input wrap with inline actions */
.wse-input-wrap {
  position: relative;
  margin-top: 4px;
}

.wse-input {
  width: 100%;
  padding: 7px 11px;
  border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540);
  background: var(--bg-0, #0C0E12);
  color: var(--ink-0, #E8EBF0);
  font-size: 13px;
  font-family: inherit;
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
  box-sizing: border-box;
}

.wse-input:focus,
.wse-input:focus-visible {
  border-color: var(--azure, #5CA8FF);
  box-shadow: 0 0 0 3px rgba(92, 168, 255, 0.14);
}

.wse-input--with-actions {
  padding-right: 64px;
}

.wse-input--mono {
  font-variant-numeric: tabular-nums;
  font-family: 'JetBrains Mono', 'Fira Mono', monospace;
}

.wse-input--flex {
  flex: 1;
}

.wse-input-actions {
  position: absolute;
  inset-y: 0;
  right: 0;
  display: flex;
  align-items: center;
  padding-right: 4px;
  gap: 2px;
}

.wse-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--ink-2, #5C6470);
  cursor: pointer;
  transition: color 0.12s;
  padding: 0;
}

.wse-icon-btn:hover {
  color: var(--ink-0, #E8EBF0);
}

.wse-icon-btn--disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.wse-icon {
  width: 15px;
  height: 15px;
}

/* Quota + Subscription grid */
.wse-grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.wse-field-label {
  font-size: 11.5px;
  color: var(--ink-2, #5C6470);
  display: block;
  margin-bottom: 4px;
}

.wse-no-margin {
  margin: 0;
}

.wse-field-hint {
  font-size: 11px;
  color: var(--ink-2, #5C6470);
  margin: 3px 0 0;
  line-height: 1.4;
}

/* Usage bar */
.wse-usage-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.wse-quota-bar {
  flex: 1;
  height: 6px;
  border-radius: 999px;
  background: var(--bg-2, #171A20);
  overflow: hidden;
}

.wse-quota-bar-spacer {
  flex: 1;
}

.wse-quota-fill {
  height: 100%;
  border-radius: 999px;
  transition: width 0.3s;
}

.wse-quota-fill--ok {
  background: #22c55e;
}

.wse-quota-fill--warn {
  background: #eab308;
}

.wse-quota-fill--danger {
  background: #ef4444;
}

.wse-reset-btn {
  font-size: 11.5px;
  color: var(--azure, #5CA8FF);
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 4px;
  white-space: nowrap;
  transition: color 0.12s;
}

.wse-reset-btn:hover {
  color: #93c5fd;
}

/* Proxy + Test row */
.wse-proxy-test-row {
  display: flex;
  align-items: flex-end;
  gap: 12px;
}

.wse-proxy-wrap {
  flex: 1;
}

.wse-btn-test {
  white-space: nowrap;
  flex-shrink: 0;
}

/* Buttons */
.wse-btn-secondary {
  padding: 6px 14px;
  border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540);
  background: transparent;
  color: var(--ink-1, #97A0AF);
  font-size: 12.5px;
  font-weight: 500;
  cursor: pointer;
  font-family: inherit;
  transition: border-color 0.15s, color 0.15s, background 0.15s;
}

.wse-btn-secondary:hover {
  border-color: var(--azure, #5CA8FF);
  color: var(--azure, #5CA8FF);
  background: rgba(92, 168, 255, 0.07);
}

.wse-btn-secondary:focus-visible {
  outline: 2px solid var(--azure, #5CA8FF);
  outline-offset: 2px;
}

.wse-btn-primary {
  padding: 7px 16px;
  border-radius: 8px;
  border: none;
  background: var(--azure, #5CA8FF);
  color: #0C0E12;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  font-family: inherit;
  transition: opacity 0.15s;
  white-space: nowrap;
}

.wse-btn-primary:hover:not(:disabled) {
  opacity: 0.88;
}

.wse-btn-primary:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

/* Save button — metal convex QUENCH style */
.wse-footer {
  display: flex;
  justify-content: flex-end;
  padding-top: 4px;
  border-top: 1px solid var(--line-0, #20242C);
}

.wse-btn-save {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 22px;
  border-radius: 9px;
  border: 1px solid rgba(92, 168, 255, 0.35);
  background: linear-gradient(180deg, rgba(92,168,255,0.18) 0%, rgba(92,168,255,0.09) 100%);
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.07), 0 2px 8px rgba(0,0,0,0.3);
  color: var(--azure, #5CA8FF);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  font-family: inherit;
  transition: opacity 0.15s, box-shadow 0.15s;
}

.wse-btn-save:hover:not(:disabled) {
  opacity: 0.88;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.09), 0 2px 12px rgba(92,168,255,0.2);
}

.wse-btn-save:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.wse-btn-save:focus-visible {
  outline: 2px solid var(--azure, #5CA8FF);
  outline-offset: 2px;
}

.wse-spin {
  width: 14px;
  height: 14px;
  animation: wse-spin 1s linear infinite;
}

@keyframes wse-spin {
  to { transform: rotate(360deg); }
}

/* Dialog */
.wse-dialog-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.55);
}

.wse-dialog {
  background: var(--metal, #15181E);
  border: 1px solid var(--line-0, #20242C);
  border-radius: 12px;
  padding: 24px;
  width: 100%;
  max-width: 520px;
  margin: 0 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
}

.wse-dialog-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--ink-0, #E8EBF0);
  margin: 0 0 16px;
}

.wse-dialog-search {
  display: flex;
  gap: 8px;
  align-items: center;
}

.wse-test-results {
  margin-top: 16px;
  max-height: 320px;
  overflow-y: auto;
  border-radius: 8px;
  background: var(--bg-2, #171A20);
  padding: 14px;
}

.wse-test-provider {
  font-size: 13px;
  font-weight: 500;
  color: var(--ink-0, #E8EBF0);
  margin: 0 0 8px;
}

.wse-test-empty {
  font-size: 13px;
  color: var(--ink-2, #5C6470);
}

.wse-test-result-item {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid var(--line-1, #2F3540);
}

.wse-test-result-item:first-child {
  margin-top: 0;
  padding-top: 0;
  border-top: none;
}

.wse-test-result-link {
  font-size: 13px;
  font-weight: 500;
  color: #60a5fa;
  text-decoration: none;
}

.wse-test-result-link:hover {
  text-decoration: underline;
}

.wse-test-result-snippet {
  font-size: 12px;
  color: var(--ink-2, #5C6470);
  margin: 4px 0 0;
  line-height: 1.4;
}

.wse-dialog-footer {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

@media (prefers-reduced-motion: reduce) {
  .wse-spin { animation: none; }
  .wse-chevron { transition: none; }
}
</style>
