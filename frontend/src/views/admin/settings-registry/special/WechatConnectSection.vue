<template>
  <div class="wc-body">
    <!-- Master switch -->
    <div class="wc-row wc-master-row">
      <div class="wc-label-group">
        <label class="wc-label">{{ t('admin.settings.wechatConnect.enabledLabel') }}</label>
        <p class="wc-hint">{{ t('admin.settings.wechatConnect.enabledHint') }}</p>
      </div>
      <Toggle
        :model-value="local.wechat_connect_enabled"
        data-testid="wechat-connect-enabled"
        @update:model-value="setField('wechat_connect_enabled', $event)"
      />
    </div>

    <template v-if="local.wechat_connect_enabled">
      <!-- ── Mode panels ─────────────────────────────────────────────── -->
      <div class="wc-panels">

        <!-- PC App (open platform) -->
        <div class="wc-panel">
          <div class="wc-panel-head">
            <div>
              <h3 class="wc-panel-title">{{ t('admin.settings.wechatConnect.open.title') }}</h3>
              <p class="wc-panel-desc">{{ t('admin.settings.wechatConnect.open.description') }}</p>
            </div>
            <Toggle
              :model-value="local.wechat_connect_open_enabled"
              data-testid="wechat-connect-open-enabled"
              @update:model-value="handleOpenEnabledChange"
            />
          </div>
          <div v-if="local.wechat_connect_open_enabled" class="wc-cred-grid">
            <div>
              <label class="wc-field-label">{{ t('admin.settings.wechatConnect.open.appId') }}</label>
              <input
                :value="local.wechat_connect_open_app_id"
                data-testid="wechat-connect-open-app-id"
                type="text"
                class="input font-mono text-sm"
                :placeholder="t('admin.settings.wechatConnect.open.appIdPlaceholder')"
                @input="setField('wechat_connect_open_app_id', ($event.target as HTMLInputElement).value)"
              />
            </div>
            <div>
              <label class="wc-field-label">{{ t('admin.settings.wechatConnect.open.appSecret') }}</label>
              <input
                :value="local.wechat_connect_open_app_secret"
                data-testid="wechat-connect-open-app-secret"
                type="password"
                class="input font-mono text-sm"
                :placeholder="local.wechat_connect_open_app_secret_configured
                  ? t('admin.settings.wechatConnect.appSecretConfiguredPlaceholder')
                  : t('admin.settings.wechatConnect.open.appSecretPlaceholder')"
                @input="setField('wechat_connect_open_app_secret', ($event.target as HTMLInputElement).value)"
              />
            </div>
          </div>
        </div>

        <!-- Official Account (mp) -->
        <div class="wc-panel">
          <div class="wc-panel-head">
            <div>
              <h3 class="wc-panel-title">{{ t('admin.settings.wechatConnect.mp.title') }}</h3>
              <p class="wc-panel-desc">{{ t('admin.settings.wechatConnect.mp.description') }}</p>
            </div>
            <Toggle
              :model-value="local.wechat_connect_mp_enabled"
              data-testid="wechat-connect-mp-enabled"
              @update:model-value="handleMPEnabledChange"
            />
          </div>
          <div v-if="local.wechat_connect_mp_enabled" class="wc-cred-grid">
            <div>
              <label class="wc-field-label">{{ t('admin.settings.wechatConnect.mp.appId') }}</label>
              <input
                :value="local.wechat_connect_mp_app_id"
                data-testid="wechat-connect-mp-app-id"
                type="text"
                class="input font-mono text-sm"
                :placeholder="t('admin.settings.wechatConnect.mp.appIdPlaceholder')"
                @input="setField('wechat_connect_mp_app_id', ($event.target as HTMLInputElement).value)"
              />
            </div>
            <div>
              <label class="wc-field-label">{{ t('admin.settings.wechatConnect.mp.appSecret') }}</label>
              <input
                :value="local.wechat_connect_mp_app_secret"
                data-testid="wechat-connect-mp-app-secret"
                type="password"
                class="input font-mono text-sm"
                :placeholder="local.wechat_connect_mp_app_secret_configured
                  ? t('admin.settings.wechatConnect.appSecretConfiguredPlaceholder')
                  : t('admin.settings.wechatConnect.mp.appSecretPlaceholder')"
                @input="setField('wechat_connect_mp_app_secret', ($event.target as HTMLInputElement).value)"
              />
            </div>
          </div>
        </div>

        <!-- Mobile App -->
        <div class="wc-panel">
          <div class="wc-panel-head">
            <div>
              <h3 class="wc-panel-title">{{ t('admin.settings.wechatConnect.mobile.title') }}</h3>
              <p class="wc-panel-desc">{{ t('admin.settings.wechatConnect.mobile.description') }}</p>
            </div>
            <Toggle
              :model-value="local.wechat_connect_mobile_enabled"
              data-testid="wechat-connect-mobile-enabled"
              @update:model-value="handleMobileEnabledChange"
            />
          </div>
          <div v-if="local.wechat_connect_mobile_enabled" class="wc-cred-grid">
            <div>
              <label class="wc-field-label">{{ t('admin.settings.wechatConnect.mobile.appId') }}</label>
              <input
                :value="local.wechat_connect_mobile_app_id"
                data-testid="wechat-connect-mobile-app-id"
                type="text"
                class="input font-mono text-sm"
                :placeholder="t('admin.settings.wechatConnect.mobile.appIdPlaceholder')"
                @input="setField('wechat_connect_mobile_app_id', ($event.target as HTMLInputElement).value)"
              />
            </div>
            <div>
              <label class="wc-field-label">{{ t('admin.settings.wechatConnect.mobile.appSecret') }}</label>
              <input
                :value="local.wechat_connect_mobile_app_secret"
                data-testid="wechat-connect-mobile-app-secret"
                type="password"
                class="input font-mono text-sm"
                :placeholder="local.wechat_connect_mobile_app_secret_configured
                  ? t('admin.settings.wechatConnect.appSecretConfiguredPlaceholder')
                  : t('admin.settings.wechatConnect.mobile.appSecretPlaceholder')"
                @input="setField('wechat_connect_mobile_app_secret', ($event.target as HTMLInputElement).value)"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- UnionID warning: open + (mp OR mobile) -->
      <div
        v-if="local.wechat_connect_open_enabled && (local.wechat_connect_mp_enabled || local.wechat_connect_mobile_enabled)"
        class="wc-unionid-warn"
      >
        {{ t('admin.settings.wechatConnect.unionIdWarning') }}
      </div>

      <!-- Redirect URL -->
      <div class="wc-field-wrap">
        <label class="wc-field-label">{{ t('admin.settings.wechatConnect.redirectUrlLabel') }}</label>
        <input
          :value="local.wechat_connect_redirect_url"
          data-testid="wechat-connect-redirect-url"
          type="url"
          class="input font-mono text-sm"
          :placeholder="t('admin.settings.wechatConnect.redirectUrlPlaceholder')"
          @input="setField('wechat_connect_redirect_url', ($event.target as HTMLInputElement).value)"
        />
        <p class="wc-help">{{ t('admin.settings.wechatConnect.redirectUrlHint') }}</p>
        <div class="wc-redirect-actions">
          <button
            type="button"
            class="btn btn-secondary btn-sm w-fit"
            @click="generateAndCopyRedirectUrl"
          >
            {{ t('admin.settings.wechatConnect.generateAndCopy') }}
          </button>
          <code
            v-if="redirectUrlSuggestion"
            class="wc-suggestion-code"
          >{{ redirectUrlSuggestion }}</code>
        </div>
      </div>

      <!-- Frontend Redirect URL -->
      <div class="wc-field-wrap">
        <label class="wc-field-label">{{ t('admin.settings.wechatConnect.frontendRedirectUrlLabel') }}</label>
        <input
          :value="local.wechat_connect_frontend_redirect_url"
          data-testid="wechat-connect-frontend-redirect-url"
          type="text"
          class="input font-mono text-sm"
          :placeholder="t('admin.settings.wechatConnect.frontendRedirectUrlPlaceholder')"
          @input="setField('wechat_connect_frontend_redirect_url', ($event.target as HTMLInputElement).value)"
        />
        <p class="wc-help">{{ t('admin.settings.wechatConnect.frontendRedirectUrlHint') }}</p>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import {
  resolveWeChatConnectModeCapabilities,
  deriveWeChatConnectStoredMode,
  defaultWeChatConnectScopesForMode,
} from '@/api/admin/settings'

const props = defineProps<{
  settings: Record<string, unknown>
  formValues?: Record<string, unknown>
}>()

const emit = defineEmits<{
  'update:field': [key: string, value: unknown]
}>()

const { t } = useI18n()

// ── Local state (mirrors parent form + settings) ────────────────────────────

type WechatLocal = {
  wechat_connect_enabled: boolean
  wechat_connect_open_enabled: boolean
  wechat_connect_mp_enabled: boolean
  wechat_connect_mobile_enabled: boolean
  wechat_connect_open_app_id: string
  wechat_connect_open_app_secret: string
  wechat_connect_open_app_secret_configured: boolean
  wechat_connect_mp_app_id: string
  wechat_connect_mp_app_secret: string
  wechat_connect_mp_app_secret_configured: boolean
  wechat_connect_mobile_app_id: string
  wechat_connect_mobile_app_secret: string
  wechat_connect_mobile_app_secret_configured: boolean
  wechat_connect_mode: string
  wechat_connect_scopes: string
  wechat_connect_redirect_url: string
  wechat_connect_frontend_redirect_url: string
  // legacy / passthrough
  wechat_connect_app_id: string
  wechat_connect_app_secret: string
  wechat_connect_app_secret_configured: boolean
}

function buildLocal(src: Record<string, unknown>): WechatLocal {
  const bool = (key: string, fallback = false): boolean =>
    typeof src[key] === 'boolean' ? (src[key] as boolean) : fallback
  const str = (key: string, fallback = ''): string =>
    typeof src[key] === 'string' ? (src[key] as string) : fallback

  const capabilities = resolveWeChatConnectModeCapabilities(
    src['wechat_connect_open_enabled'],
    src['wechat_connect_mp_enabled'],
    src['wechat_connect_mobile_enabled'],
    src['wechat_connect_mode'],
  )

  const mode = deriveWeChatConnectStoredMode(
    capabilities.openEnabled,
    capabilities.mpEnabled,
    capabilities.mobileEnabled,
    src['wechat_connect_mode'],
  )

  return {
    wechat_connect_enabled: bool('wechat_connect_enabled'),
    wechat_connect_open_enabled: capabilities.openEnabled,
    wechat_connect_mp_enabled: capabilities.mpEnabled,
    wechat_connect_mobile_enabled: capabilities.mobileEnabled,
    wechat_connect_open_app_id: str('wechat_connect_open_app_id'),
    wechat_connect_open_app_secret: str('wechat_connect_open_app_secret'),
    wechat_connect_open_app_secret_configured: bool('wechat_connect_open_app_secret_configured'),
    wechat_connect_mp_app_id: str('wechat_connect_mp_app_id'),
    wechat_connect_mp_app_secret: str('wechat_connect_mp_app_secret'),
    wechat_connect_mp_app_secret_configured: bool('wechat_connect_mp_app_secret_configured'),
    wechat_connect_mobile_app_id: str('wechat_connect_mobile_app_id'),
    wechat_connect_mobile_app_secret: str('wechat_connect_mobile_app_secret'),
    wechat_connect_mobile_app_secret_configured: bool('wechat_connect_mobile_app_secret_configured'),
    wechat_connect_mode: mode,
    wechat_connect_scopes: str('wechat_connect_scopes') || defaultWeChatConnectScopesForMode(mode),
    wechat_connect_redirect_url: str('wechat_connect_redirect_url'),
    wechat_connect_frontend_redirect_url: str('wechat_connect_frontend_redirect_url', '/auth/wechat/callback'),
    wechat_connect_app_id: str('wechat_connect_app_id'),
    wechat_connect_app_secret: str('wechat_connect_app_secret'),
    wechat_connect_app_secret_configured: bool('wechat_connect_app_secret_configured'),
  }
}

const activeSource = computed(() => props.formValues ?? props.settings)
const local = ref<WechatLocal>(buildLocal(activeSource.value))

// Re-sync when parent resets (e.g., after save)
watch(
  () => activeSource.value,
  (incoming) => {
    local.value = buildLocal(incoming)
  },
  { deep: true },
)

// ── Helpers ─────────────────────────────────────────────────────────────────

function setField(key: keyof WechatLocal, value: unknown) {
  ;(local.value as Record<string, unknown>)[key] = value
  emit('update:field', key, value)
}

function syncMode(preferredMode?: 'open' | 'mp' | 'mobile') {
  // mp and mobile are mutually exclusive
  if (local.value.wechat_connect_mp_enabled && local.value.wechat_connect_mobile_enabled) {
    if (preferredMode === 'mobile') {
      local.value.wechat_connect_mp_enabled = false
      emit('update:field', 'wechat_connect_mp_enabled', false)
    } else {
      local.value.wechat_connect_mobile_enabled = false
      emit('update:field', 'wechat_connect_mobile_enabled', false)
    }
  }

  const capabilities = resolveWeChatConnectModeCapabilities(
    local.value.wechat_connect_open_enabled,
    local.value.wechat_connect_mp_enabled,
    local.value.wechat_connect_mobile_enabled,
    local.value.wechat_connect_mode,
  )

  local.value.wechat_connect_open_enabled = capabilities.openEnabled
  local.value.wechat_connect_mp_enabled = capabilities.mpEnabled
  local.value.wechat_connect_mobile_enabled = capabilities.mobileEnabled

  const newMode = deriveWeChatConnectStoredMode(
    capabilities.openEnabled,
    capabilities.mpEnabled,
    capabilities.mobileEnabled,
    local.value.wechat_connect_mode,
  )
  local.value.wechat_connect_mode = newMode
  local.value.wechat_connect_scopes = defaultWeChatConnectScopesForMode(newMode)

  emit('update:field', 'wechat_connect_open_enabled', capabilities.openEnabled)
  emit('update:field', 'wechat_connect_mp_enabled', capabilities.mpEnabled)
  emit('update:field', 'wechat_connect_mobile_enabled', capabilities.mobileEnabled)
  emit('update:field', 'wechat_connect_mode', newMode)
  emit('update:field', 'wechat_connect_scopes', local.value.wechat_connect_scopes)
}

function handleOpenEnabledChange(value: boolean) {
  local.value.wechat_connect_open_enabled = value
  syncMode(value ? 'open' : undefined)
}

function handleMPEnabledChange(value: boolean) {
  local.value.wechat_connect_mp_enabled = value
  if (value) {
    local.value.wechat_connect_mobile_enabled = false
  }
  syncMode(value ? 'mp' : undefined)
}

function handleMobileEnabledChange(value: boolean) {
  local.value.wechat_connect_mobile_enabled = value
  if (value) {
    local.value.wechat_connect_mp_enabled = false
  }
  syncMode(value ? 'mobile' : undefined)
}

// ── Redirect URL suggestion ──────────────────────────────────────────────────

const redirectUrlSuggestion = computed(() => {
  if (typeof window === 'undefined') return ''
  const origin = window.location.origin || `${window.location.protocol}//${window.location.host}`
  return `${origin}/api/v1/auth/oauth/wechat/callback`
})

async function generateAndCopyRedirectUrl() {
  const url = redirectUrlSuggestion.value
  if (!url) return
  setField('wechat_connect_redirect_url', url)
  try {
    await navigator.clipboard.writeText(url)
  } catch {
    // clipboard not available — still set the field
  }
}
</script>

<style scoped>
.wc-body {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* master switch row */
.wc-master-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.wc-row {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.wc-label-group {
  flex: 1;
}

.wc-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--ink-0, #E8EBF0);
}

.wc-hint {
  font-size: 12px;
  color: var(--ink-2, #5C6470);
  margin: 2px 0 0;
  line-height: 1.5;
}

/* mode panels */
.wc-panels {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.wc-panel {
  border: 1px solid var(--line-1, #2F3540);
  border-radius: 8px;
  padding: 16px;
}

.wc-panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.wc-panel-title {
  font-size: 13.5px;
  font-weight: 500;
  color: var(--ink-0, #E8EBF0);
  margin: 0 0 4px;
}

.wc-panel-desc {
  font-size: 12px;
  color: var(--ink-2, #5C6470);
  margin: 0;
  line-height: 1.5;
}

.wc-cred-grid {
  margin-top: 16px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
}

.wc-field-label {
  display: block;
  font-size: 12.5px;
  font-weight: 500;
  color: rgba(232, 235, 240, 0.85);
  margin-bottom: 6px;
}

/* UnionID warning banner */
.wc-unionid-warn {
  border: 1px solid rgba(245, 158, 11, 0.3);
  background: rgba(245, 158, 11, 0.08);
  border-radius: 6px;
  padding: 10px 14px;
  font-size: 12.5px;
  color: rgb(251, 191, 36);
  line-height: 1.55;
}

/* redirect URL area */
.wc-field-wrap {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.wc-help {
  font-size: 11.5px;
  color: var(--ink-2, #5C6470);
  margin: 0;
  line-height: 1.5;
}

.wc-redirect-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  margin-top: 4px;
}

.wc-suggestion-code {
  font-size: 11.5px;
  font-family: ui-monospace, monospace;
  color: var(--ink-2, #5C6470);
  background: var(--bg-2, #171A20);
  border-radius: 4px;
  padding: 3px 8px;
  word-break: break-all;
  user-select: all;
}
</style>
