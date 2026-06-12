<template>
  <div class="aak-body">
    <!-- Security Warning -->
    <div class="aak-warning">
      <svg class="aak-warning-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
      </svg>
      <p class="aak-warning-text">{{ t('admin.settings.adminApiKey.securityWarning') }}</p>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="aak-loading">
      <div class="aak-spinner" />
      {{ t('common.loading') }}
    </div>

    <!-- No Key -->
    <div v-else-if="!keyExists" class="aak-no-key">
      <span class="aak-muted">{{ t('admin.settings.adminApiKey.notConfigured') }}</span>
      <button
        type="button"
        class="aak-btn-primary"
        :disabled="operating"
        @click="create"
      >
        <svg v-if="operating" class="aak-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
        {{ operating ? t('admin.settings.adminApiKey.creating') : t('admin.settings.adminApiKey.create') }}
      </button>
    </div>

    <!-- Key Exists -->
    <div v-else class="aak-key-section">
      <div class="aak-current-key-row">
        <div>
          <label class="aak-field-label">{{ t('admin.settings.adminApiKey.currentKey') }}</label>
          <code class="aak-masked-key">{{ maskedKey }}</code>
        </div>
        <div class="aak-actions">
          <button
            type="button"
            class="aak-btn-secondary"
            :disabled="operating"
            @click="regenerate"
          >
            {{ operating ? t('admin.settings.adminApiKey.regenerating') : t('admin.settings.adminApiKey.regenerate') }}
          </button>
          <button
            type="button"
            class="aak-btn-danger"
            :disabled="operating"
            @click="remove"
          >
            {{ t('admin.settings.adminApiKey.delete') }}
          </button>
        </div>
      </div>

      <!-- Newly Generated Key (one-time reveal) -->
      <div v-if="newKey" class="aak-new-key-box">
        <p class="aak-new-key-warning">{{ t('admin.settings.adminApiKey.keyWarning') }}</p>
        <div class="aak-new-key-row">
          <code class="aak-new-key-value">{{ newKey }}</code>
          <button type="button" class="aak-btn-primary aak-btn-copy" @click="copyNewKey">
            {{ t('admin.settings.adminApiKey.copyKey') }}
          </button>
        </div>
        <p class="aak-usage-hint">{{ t('admin.settings.adminApiKey.usage') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import adminAPI from '@/api/admin'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'

// Self-contained: ignores parent settings/form props — has own CRUD cycle
defineProps<{
  settings?: Record<string, unknown>
  formValues?: Record<string, unknown>
}>()

const { t } = useI18n()
const appStore = useAppStore()

// ── State ──────────────────────────────────────────────────────────────────────

const loading = ref(true)
const keyExists = ref(false)
const maskedKey = ref('')
const operating = ref(false)
const newKey = ref('')

// ── Lifecycle ──────────────────────────────────────────────────────────────────

onMounted(async () => {
  loading.value = true
  try {
    const status = await adminAPI.settings.getAdminApiKey()
    keyExists.value = status.exists
    maskedKey.value = status.masked_key
  } catch {
    // Silent fail — status is non-critical
  } finally {
    loading.value = false
  }
})

// ── CRUD ───────────────────────────────────────────────────────────────────────

async function create() {
  operating.value = true
  try {
    const result = await adminAPI.settings.regenerateAdminApiKey()
    newKey.value = result.key
    keyExists.value = true
    maskedKey.value = result.key.substring(0, 10) + '...' + result.key.slice(-4)
    appStore.showSuccess(t('admin.settings.adminApiKey.keyGenerated'))
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    operating.value = false
  }
}

async function regenerate() {
  if (!confirm(t('admin.settings.adminApiKey.regenerateConfirm'))) return
  await create()
}

async function remove() {
  if (!confirm(t('admin.settings.adminApiKey.deleteConfirm'))) return
  operating.value = true
  try {
    await adminAPI.settings.deleteAdminApiKey()
    keyExists.value = false
    maskedKey.value = ''
    newKey.value = ''
    appStore.showSuccess(t('admin.settings.adminApiKey.keyDeleted'))
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    operating.value = false
  }
}

function copyNewKey() {
  navigator.clipboard
    .writeText(newKey.value)
    .then(() => { appStore.showSuccess(t('admin.settings.adminApiKey.keyCopied')) })
    .catch(() => { appStore.showError(t('common.copyFailed')) })
}
</script>

<style scoped>
.aak-body {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* Warning banner */
.aak-warning {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: 8px;
  background: rgba(245, 158, 11, 0.08);
  padding: 12px 14px;
}

.aak-warning-icon {
  width: 18px;
  height: 18px;
  color: #fbbf24;
  flex-shrink: 0;
  margin-top: 1px;
}

.aak-warning-text {
  font-size: 12.5px;
  color: #fbbf24;
  margin: 0;
  line-height: 1.5;
}

/* Loading */
.aak-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--ink-2, #5C6470);
}

.aak-spinner {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 2px solid var(--line-1, #2F3540);
  border-top-color: var(--azure, #5CA8FF);
  animation: aak-spin 1s linear infinite;
}

/* No-key state */
.aak-no-key {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.aak-muted {
  font-size: 13px;
  color: var(--ink-2, #5C6470);
}

/* Key-exists section */
.aak-key-section {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.aak-current-key-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.aak-field-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--ink-1, #97A0AF);
  display: block;
  margin-bottom: 6px;
}

.aak-masked-key {
  display: inline-block;
  font-family: 'JetBrains Mono', 'Fira Mono', monospace;
  font-size: 13px;
  color: var(--ink-0, #E8EBF0);
  background: var(--bg-2, #171A20);
  border-radius: 6px;
  padding: 5px 10px;
  letter-spacing: 0.03em;
}

.aak-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

/* New key reveal box */
.aak-new-key-box {
  border: 1px solid rgba(52, 211, 153, 0.3);
  border-radius: 10px;
  background: rgba(52, 211, 153, 0.07);
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.aak-new-key-warning {
  font-size: 13px;
  font-weight: 500;
  color: #34d399;
  margin: 0;
}

.aak-new-key-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.aak-new-key-value {
  flex: 1;
  font-family: 'JetBrains Mono', 'Fira Mono', monospace;
  font-size: 12.5px;
  color: var(--ink-0, #E8EBF0);
  background: var(--bg-0, #0C0E12);
  border: 1px solid rgba(52, 211, 153, 0.25);
  border-radius: 7px;
  padding: 8px 12px;
  word-break: break-all;
  user-select: all;
}

.aak-usage-hint {
  font-size: 11.5px;
  color: #34d399;
  margin: 0;
  line-height: 1.45;
}

/* Buttons */
.aak-btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 18px;
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

.aak-btn-primary:hover:not(:disabled) {
  opacity: 0.88;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.09), 0 2px 12px rgba(92,168,255,0.2);
}

.aak-btn-primary:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.aak-btn-primary:focus-visible {
  outline: 2px solid var(--azure, #5CA8FF);
  outline-offset: 2px;
}

.aak-btn-copy {
  flex-shrink: 0;
}

.aak-btn-secondary {
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

.aak-btn-secondary:hover:not(:disabled) {
  border-color: var(--azure, #5CA8FF);
  color: var(--azure, #5CA8FF);
  background: rgba(92, 168, 255, 0.07);
}

.aak-btn-secondary:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.aak-btn-secondary:focus-visible {
  outline: 2px solid var(--azure, #5CA8FF);
  outline-offset: 2px;
}

.aak-btn-danger {
  padding: 6px 14px;
  border-radius: 8px;
  border: 1px solid rgba(239, 68, 68, 0.3);
  background: transparent;
  color: #f87171;
  font-size: 12.5px;
  font-weight: 500;
  cursor: pointer;
  font-family: inherit;
  transition: border-color 0.15s, color 0.15s, background 0.15s;
}

.aak-btn-danger:hover:not(:disabled) {
  border-color: rgba(239, 68, 68, 0.5);
  color: #fca5a5;
  background: rgba(239, 68, 68, 0.08);
}

.aak-btn-danger:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.aak-btn-danger:focus-visible {
  outline: 2px solid #ef4444;
  outline-offset: 2px;
}

@keyframes aak-spin {
  to { transform: rotate(360deg); }
}

.aak-spin {
  width: 14px;
  height: 14px;
  animation: aak-spin 1s linear infinite;
}

@media (prefers-reduced-motion: reduce) {
  .aak-spinner, .aak-spin { animation: none; }
}
</style>
