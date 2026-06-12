<template>
  <div class="oidc-body">
    <!-- enable toggle -->
    <div class="oidc-row-between">
      <div>
        <label class="oidc-label">{{ t('admin.settings.oidc.enable') }}</label>
        <p class="oidc-hint">{{ t('admin.settings.oidc.enableHint') }}</p>
      </div>
      <Toggle :model-value="!!local.oidc_connect_enabled" @update:model-value="set('oidc_connect_enabled', $event)" />
    </div>

    <!-- expanded fields — only when enabled -->
    <div v-if="local.oidc_connect_enabled" class="oidc-expanded">
      <!-- Row 1: Provider Name / Client ID / Client Secret -->
      <div class="oidc-grid-3">
        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.providerName') }}</label>
          <input
            :value="local.oidc_connect_provider_name"
            type="text"
            class="input"
            :placeholder="t('admin.settings.oidc.providerNamePlaceholder')"
            @input="set('oidc_connect_provider_name', ($event.target as HTMLInputElement).value)"
          />
        </div>

        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.clientId') }}</label>
          <input
            :value="local.oidc_connect_client_id"
            type="text"
            class="input font-mono text-sm"
            :placeholder="t('admin.settings.oidc.clientIdPlaceholder')"
            @input="set('oidc_connect_client_id', ($event.target as HTMLInputElement).value)"
          />
        </div>

        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.clientSecret') }}</label>
          <input
            :value="local.oidc_connect_client_secret"
            type="password"
            class="input font-mono text-sm"
            :placeholder="local.oidc_connect_client_secret_configured
              ? t('admin.settings.oidc.clientSecretConfiguredPlaceholder')
              : t('admin.settings.oidc.clientSecretPlaceholder')"
            @input="set('oidc_connect_client_secret', ($event.target as HTMLInputElement).value)"
          />
          <p class="oidc-field-hint">
            {{ local.oidc_connect_client_secret_configured
              ? t('admin.settings.oidc.clientSecretConfiguredHint')
              : t('admin.settings.oidc.clientSecretHint') }}
          </p>
        </div>
      </div>

      <!-- Row 2: Issuer / Discovery / Authorize / Token / Userinfo / JWKS -->
      <div class="oidc-grid-2">
        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.issuerUrl') }}</label>
          <input
            :value="local.oidc_connect_issuer_url"
            type="url"
            class="input font-mono text-sm"
            :placeholder="t('admin.settings.oidc.issuerUrlPlaceholder')"
            @input="set('oidc_connect_issuer_url', ($event.target as HTMLInputElement).value)"
          />
        </div>

        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.discoveryUrl') }}</label>
          <input
            :value="local.oidc_connect_discovery_url"
            type="url"
            class="input font-mono text-sm"
            :placeholder="t('admin.settings.oidc.discoveryUrlPlaceholder')"
            @input="set('oidc_connect_discovery_url', ($event.target as HTMLInputElement).value)"
          />
        </div>

        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.authorizeUrl') }}</label>
          <input
            :value="local.oidc_connect_authorize_url"
            type="url"
            class="input font-mono text-sm"
            :placeholder="t('admin.settings.oidc.authorizeUrlPlaceholder')"
            @input="set('oidc_connect_authorize_url', ($event.target as HTMLInputElement).value)"
          />
        </div>

        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.tokenUrl') }}</label>
          <input
            :value="local.oidc_connect_token_url"
            type="url"
            class="input font-mono text-sm"
            :placeholder="t('admin.settings.oidc.tokenUrlPlaceholder')"
            @input="set('oidc_connect_token_url', ($event.target as HTMLInputElement).value)"
          />
        </div>

        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.userinfoUrl') }}</label>
          <input
            :value="local.oidc_connect_userinfo_url"
            type="url"
            class="input font-mono text-sm"
            :placeholder="t('admin.settings.oidc.userinfoUrlPlaceholder')"
            @input="set('oidc_connect_userinfo_url', ($event.target as HTMLInputElement).value)"
          />
        </div>

        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.jwksUrl') }}</label>
          <input
            :value="local.oidc_connect_jwks_url"
            type="url"
            class="input font-mono text-sm"
            :placeholder="t('admin.settings.oidc.jwksUrlPlaceholder')"
            @input="set('oidc_connect_jwks_url', ($event.target as HTMLInputElement).value)"
          />
        </div>
      </div>

      <!-- Row 3: Scopes / Redirect URL / Frontend Redirect URL -->
      <div class="oidc-grid-2">
        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.scopes') }}</label>
          <input
            :value="local.oidc_connect_scopes"
            type="text"
            class="input font-mono text-sm"
            :placeholder="t('admin.settings.oidc.scopesPlaceholder')"
            @input="set('oidc_connect_scopes', ($event.target as HTMLInputElement).value)"
          />
          <p class="oidc-field-hint">{{ t('admin.settings.oidc.scopesHint') }}</p>
        </div>

        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.redirectUrl') }}</label>
          <input
            :value="local.oidc_connect_redirect_url"
            type="url"
            class="input font-mono text-sm"
            :placeholder="t('admin.settings.oidc.redirectUrlPlaceholder')"
            @input="set('oidc_connect_redirect_url', ($event.target as HTMLInputElement).value)"
          />
          <!-- Quick-set / copy suggestion -->
          <div class="oidc-redirect-actions">
            <button type="button" class="btn btn-secondary btn-sm w-fit" @click="quickSetRedirectUrl">
              {{ t('admin.settings.oidc.quickSetCopy') }}
            </button>
            <code v-if="redirectUrlSuggestion" class="oidc-suggestion">
              {{ redirectUrlSuggestion }}
            </code>
          </div>
          <p class="oidc-field-hint">{{ t('admin.settings.oidc.redirectUrlHint') }}</p>
        </div>

        <div class="oidc-field oidc-col-span-2">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.frontendRedirectUrl') }}</label>
          <input
            :value="local.oidc_connect_frontend_redirect_url"
            type="text"
            class="input font-mono text-sm"
            :placeholder="t('admin.settings.oidc.frontendRedirectUrlPlaceholder')"
            @input="set('oidc_connect_frontend_redirect_url', ($event.target as HTMLInputElement).value)"
          />
          <p class="oidc-field-hint">{{ t('admin.settings.oidc.frontendRedirectUrlHint') }}</p>
        </div>
      </div>

      <!-- Row 4: Token Auth Method / Clock Skew / Signing Algs -->
      <div class="oidc-grid-3">
        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.tokenAuthMethod') }}</label>
          <select
            :value="local.oidc_connect_token_auth_method"
            class="input font-mono text-sm"
            @change="set('oidc_connect_token_auth_method', ($event.target as HTMLSelectElement).value)"
          >
            <option value="client_secret_post">client_secret_post</option>
            <option value="client_secret_basic">client_secret_basic</option>
            <option value="none">none</option>
          </select>
        </div>

        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.clockSkewSeconds') }}</label>
          <input
            :value="local.oidc_connect_clock_skew_seconds"
            type="number"
            min="0"
            max="600"
            class="input"
            @input="set('oidc_connect_clock_skew_seconds', Number(($event.target as HTMLInputElement).value))"
          />
        </div>

        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.allowedSigningAlgs') }}</label>
          <input
            :value="local.oidc_connect_allowed_signing_algs"
            type="text"
            class="input font-mono text-sm"
            :placeholder="t('admin.settings.oidc.allowedSigningAlgsPlaceholder')"
            @input="set('oidc_connect_allowed_signing_algs', ($event.target as HTMLInputElement).value)"
          />
        </div>
      </div>

      <!-- Row 5: PKCE / Validate ID Token / Require Email Verified (toggle cards) -->
      <div class="oidc-grid-3">
        <div class="oidc-toggle-card">
          <label class="oidc-label">{{ t('admin.settings.oidc.usePkce') }}</label>
          <Toggle
            :model-value="!!local.oidc_connect_use_pkce"
            data-testid="oidc-connect-use-pkce"
            @update:model-value="set('oidc_connect_use_pkce', $event)"
          />
        </div>

        <div class="oidc-toggle-card">
          <label class="oidc-label">{{ t('admin.settings.oidc.validateIdToken') }}</label>
          <Toggle
            :model-value="!!local.oidc_connect_validate_id_token"
            data-testid="oidc-connect-validate-id-token"
            @update:model-value="set('oidc_connect_validate_id_token', $event)"
          />
        </div>

        <div class="oidc-toggle-card">
          <label class="oidc-label">{{ t('admin.settings.oidc.requireEmailVerified') }}</label>
          <Toggle
            :model-value="!!local.oidc_connect_require_email_verified"
            @update:model-value="set('oidc_connect_require_email_verified', $event)"
          />
        </div>
      </div>

      <!-- Row 6: Userinfo path overrides -->
      <div class="oidc-grid-3">
        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.userinfoEmailPath') }}</label>
          <input
            :value="local.oidc_connect_userinfo_email_path"
            type="text"
            class="input font-mono text-sm"
            :placeholder="t('admin.settings.oidc.userinfoEmailPathPlaceholder')"
            @input="set('oidc_connect_userinfo_email_path', ($event.target as HTMLInputElement).value)"
          />
        </div>

        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.userinfoIdPath') }}</label>
          <input
            :value="local.oidc_connect_userinfo_id_path"
            type="text"
            class="input font-mono text-sm"
            :placeholder="t('admin.settings.oidc.userinfoIdPathPlaceholder')"
            @input="set('oidc_connect_userinfo_id_path', ($event.target as HTMLInputElement).value)"
          />
        </div>

        <div class="oidc-field">
          <label class="oidc-field-label">{{ t('admin.settings.oidc.userinfoUsernamePath') }}</label>
          <input
            :value="local.oidc_connect_userinfo_username_path"
            type="text"
            class="input font-mono text-sm"
            :placeholder="t('admin.settings.oidc.userinfoUsernamePathPlaceholder')"
            @input="set('oidc_connect_userinfo_username_path', ($event.target as HTMLInputElement).value)"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import { useClipboard } from '@/composables/useClipboard'

const { t } = useI18n()
const { copyToClipboard } = useClipboard()

const props = defineProps<{
  settings: Record<string, unknown>
  formValues?: Record<string, unknown>
}>()

const emit = defineEmits<{
  'update:field': [key: string, value: unknown]
}>()

// Prefer live form dirty-state over persisted settings
const activeSource = () => props.formValues ?? props.settings

// All OIDC keys mirrored locally for reactive rendering
interface OidcLocal {
  oidc_connect_enabled: boolean
  oidc_connect_provider_name: string
  oidc_connect_client_id: string
  oidc_connect_client_secret: string
  oidc_connect_client_secret_configured: boolean
  oidc_connect_issuer_url: string
  oidc_connect_discovery_url: string
  oidc_connect_authorize_url: string
  oidc_connect_token_url: string
  oidc_connect_userinfo_url: string
  oidc_connect_jwks_url: string
  oidc_connect_scopes: string
  oidc_connect_redirect_url: string
  oidc_connect_frontend_redirect_url: string
  oidc_connect_token_auth_method: string
  oidc_connect_use_pkce: boolean
  oidc_connect_validate_id_token: boolean
  oidc_connect_allowed_signing_algs: string
  oidc_connect_clock_skew_seconds: number
  oidc_connect_require_email_verified: boolean
  oidc_connect_userinfo_email_path: string
  oidc_connect_userinfo_id_path: string
  oidc_connect_userinfo_username_path: string
}

function pick(src: Record<string, unknown>): OidcLocal {
  return {
    oidc_connect_enabled: !!(src['oidc_connect_enabled'] ?? false),
    oidc_connect_provider_name: (src['oidc_connect_provider_name'] as string) ?? 'OIDC',
    oidc_connect_client_id: (src['oidc_connect_client_id'] as string) ?? '',
    oidc_connect_client_secret: (src['oidc_connect_client_secret'] as string) ?? '',
    oidc_connect_client_secret_configured: !!(src['oidc_connect_client_secret_configured'] ?? false),
    oidc_connect_issuer_url: (src['oidc_connect_issuer_url'] as string) ?? '',
    oidc_connect_discovery_url: (src['oidc_connect_discovery_url'] as string) ?? '',
    oidc_connect_authorize_url: (src['oidc_connect_authorize_url'] as string) ?? '',
    oidc_connect_token_url: (src['oidc_connect_token_url'] as string) ?? '',
    oidc_connect_userinfo_url: (src['oidc_connect_userinfo_url'] as string) ?? '',
    oidc_connect_jwks_url: (src['oidc_connect_jwks_url'] as string) ?? '',
    oidc_connect_scopes: (src['oidc_connect_scopes'] as string) ?? 'openid email profile',
    oidc_connect_redirect_url: (src['oidc_connect_redirect_url'] as string) ?? '',
    oidc_connect_frontend_redirect_url: (src['oidc_connect_frontend_redirect_url'] as string) ?? '/auth/oidc/callback',
    oidc_connect_token_auth_method: (src['oidc_connect_token_auth_method'] as string) ?? 'client_secret_post',
    oidc_connect_use_pkce: !!(src['oidc_connect_use_pkce'] ?? false),
    oidc_connect_validate_id_token: !!(src['oidc_connect_validate_id_token'] ?? false),
    oidc_connect_allowed_signing_algs: (src['oidc_connect_allowed_signing_algs'] as string) ?? 'RS256,ES256,PS256',
    oidc_connect_clock_skew_seconds: (src['oidc_connect_clock_skew_seconds'] as number) ?? 120,
    oidc_connect_require_email_verified: !!(src['oidc_connect_require_email_verified'] ?? false),
    oidc_connect_userinfo_email_path: (src['oidc_connect_userinfo_email_path'] as string) ?? '',
    oidc_connect_userinfo_id_path: (src['oidc_connect_userinfo_id_path'] as string) ?? '',
    oidc_connect_userinfo_username_path: (src['oidc_connect_userinfo_username_path'] as string) ?? '',
  }
}

const local = reactive<OidcLocal>(pick(activeSource()))

// Re-sync when parent resets (discard / initial load)
watch(
  () => activeSource(),
  (src) => {
    const fresh = pick(src)
    for (const k of Object.keys(fresh) as (keyof OidcLocal)[]) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      ;(local as any)[k] = (fresh as any)[k]
    }
  },
  { deep: true },
)

function set(key: keyof OidcLocal, value: unknown) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ;(local as any)[key] = value
  emit('update:field', key, value)
}

// Redirect URL suggestion (mirrors SettingsView.vue oidcRedirectUrlSuggestion computed)
const redirectUrlSuggestion = computed(() => {
  if (typeof window === 'undefined') return ''
  const origin = window.location.origin || `${window.location.protocol}//${window.location.host}`
  return `${origin}/api/v1/auth/oauth/oidc/callback`
})

async function quickSetRedirectUrl() {
  const url = redirectUrlSuggestion.value
  if (!url) return
  set('oidc_connect_redirect_url', url)
  await copyToClipboard(url, t('admin.settings.oidc.redirectUrlSetAndCopied'))
}
</script>

<style scoped>
.oidc-body { padding: 16px 20px; display: flex; flex-direction: column; gap: 16px; }

.oidc-row-between {
  display: flex; align-items: center; justify-content: space-between; gap: 16px;
}

.oidc-label { font-size: 13px; font-weight: 500; color: var(--ink-0, #E8EBF0); }
.oidc-hint  { font-size: 11.5px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 2px 0 0; }

.oidc-expanded { display: flex; flex-direction: column; gap: 20px; border-top: 1px solid var(--line-1, #2F3540); padding-top: 16px; }

.oidc-field { display: flex; flex-direction: column; gap: 4px; }
.oidc-field-label { font-size: 12px; font-weight: 500; color: var(--ink-1, #97A0AF); }
.oidc-field-hint  { font-size: 11px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 0; }

/* grid layouts matching SettingsView.vue lg:grid-cols-2 / 3 */
.oidc-grid-2 { display: grid; grid-template-columns: 1fr; gap: 16px; }
.oidc-grid-3 { display: grid; grid-template-columns: 1fr; gap: 16px; }
@media (min-width: 768px) {
  .oidc-grid-2 { grid-template-columns: 1fr 1fr; }
  .oidc-grid-3 { grid-template-columns: 1fr 1fr 1fr; }
}
.oidc-col-span-2 { grid-column: 1 / -1; }

/* redirect URL quick-set */
.oidc-redirect-actions {
  display: flex; flex-direction: column; gap: 6px; margin-top: 6px;
}
@media (min-width: 480px) {
  .oidc-redirect-actions { flex-direction: row; align-items: center; gap: 10px; }
}
.oidc-suggestion {
  font-family: var(--font-mono, "IBM Plex Mono", monospace);
  font-size: 11px; padding: 2px 8px; border-radius: 4px;
  background: var(--bg-2, #171A20); color: var(--ink-2, #5C6470);
  word-break: break-all; user-select: all;
}

/* toggle card (PKCE / validate id_token / require email) */
.oidc-toggle-card {
  display: flex; align-items: center; justify-content: space-between; gap: 12px;
  padding: 12px 16px; border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540);
}
</style>
