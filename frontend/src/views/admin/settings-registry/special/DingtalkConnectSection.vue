<template>
  <div class="dtk-body">
    <!-- enable toggle -->
    <div class="dtk-row-between">
      <div>
        <label class="dtk-label">{{ t('admin.settings.dingtalk.enable') }}</label>
        <p class="dtk-hint">{{ t('admin.settings.dingtalk.enableHint') }}</p>
      </div>
      <Toggle :model-value="!!local.dingtalk_connect_enabled" @update:model-value="set('dingtalk_connect_enabled', $event)" />
    </div>

    <!-- expanded fields — only when enabled -->
    <div v-if="local.dingtalk_connect_enabled" class="dtk-expanded">
      <!-- App Key / Client ID -->
      <div class="dtk-field">
        <label class="dtk-field-label">{{ t('admin.settings.dingtalk.clientId') }}</label>
        <input
          :value="local.dingtalk_connect_client_id"
          type="text"
          class="input font-mono text-sm"
          :placeholder="t('admin.settings.dingtalk.clientIdPlaceholder')"
          @input="set('dingtalk_connect_client_id', ($event.target as HTMLInputElement).value)"
        />
        <p class="dtk-field-hint">{{ t('admin.settings.dingtalk.clientIdHint') }}</p>
      </div>

      <!-- App Secret / Client Secret (masked) -->
      <div class="dtk-field">
        <label class="dtk-field-label">{{ t('admin.settings.dingtalk.clientSecret') }}</label>
        <input
          :value="local.dingtalk_connect_client_secret"
          type="password"
          class="input font-mono text-sm"
          :placeholder="local.dingtalk_connect_client_secret_configured
            ? t('admin.settings.dingtalk.clientSecretConfiguredPlaceholder')
            : t('admin.settings.dingtalk.clientSecretPlaceholder')"
          @input="set('dingtalk_connect_client_secret', ($event.target as HTMLInputElement).value)"
        />
        <p class="dtk-field-hint">
          {{ local.dingtalk_connect_client_secret_configured
            ? t('admin.settings.dingtalk.clientSecretConfiguredHint')
            : t('admin.settings.dingtalk.clientSecretHint') }}
        </p>
      </div>

      <!-- Redirect URL -->
      <div class="dtk-field">
        <label class="dtk-field-label">{{ t('admin.settings.dingtalk.redirectUrl') }}</label>
        <input
          :value="local.dingtalk_connect_redirect_url"
          type="url"
          class="input font-mono text-sm"
          :placeholder="t('admin.settings.dingtalk.redirectUrlPlaceholder')"
          @input="set('dingtalk_connect_redirect_url', ($event.target as HTMLInputElement).value)"
        />
        <p class="dtk-field-hint">{{ t('admin.settings.dingtalk.redirectUrlHint') }}</p>
      </div>

      <!-- Corp Restriction Policy -->
      <div class="dtk-field dtk-border-top">
        <label class="dtk-field-label">{{ t('admin.settings.dingtalk.corpPolicy.label') }}</label>
        <p class="dtk-field-hint dtk-field-hint--top">{{ t('admin.settings.dingtalk.corpPolicy.hint') }}</p>
        <div class="dtk-radio-group">
          <label class="dtk-radio-item">
            <input
              type="radio"
              value="none"
              class="h-4 w-4 text-primary-600"
              :checked="local.dingtalk_connect_corp_restriction_policy === 'none'"
              @change="onCorpPolicyChange('none')"
            />
            <span class="text-sm text-foreground/85">{{ t('admin.settings.dingtalk.corpPolicy.none') }}</span>
          </label>
          <label class="dtk-radio-item">
            <input
              type="radio"
              value="internal_only"
              class="h-4 w-4 text-primary-600"
              :checked="local.dingtalk_connect_corp_restriction_policy === 'internal_only'"
              @change="onCorpPolicyChange('internal_only')"
            />
            <span class="text-sm text-foreground/85">{{ t('admin.settings.dingtalk.corpPolicy.internalOnly') }}</span>
          </label>
        </div>
      </div>

      <!-- internal_only-gated fields -->
      <template v-if="local.dingtalk_connect_corp_restriction_policy === 'internal_only'">
        <!-- Bypass Registration -->
        <div class="dtk-row-between dtk-border-top">
          <div>
            <label class="dtk-label">{{ t('admin.settings.dingtalk.bypassRegistration') }}</label>
            <p class="dtk-hint">{{ t('admin.settings.dingtalk.bypassRegistrationHint') }}</p>
          </div>
          <Toggle :model-value="!!local.dingtalk_connect_bypass_registration" @update:model-value="set('dingtalk_connect_bypass_registration', $event)" />
        </div>

        <!-- Sync Display Name -->
        <div class="dtk-border-top dtk-sync-block">
          <div class="dtk-row-between">
            <div>
              <label class="dtk-label">{{ t('admin.settings.dingtalk.syncDisplayName') }}</label>
              <p class="dtk-hint">{{ t('admin.settings.dingtalk.syncDisplayNameHint') }}</p>
            </div>
            <Toggle :model-value="!!local.dingtalk_connect_sync_display_name" @update:model-value="set('dingtalk_connect_sync_display_name', $event)" />
          </div>
          <template v-if="local.dingtalk_connect_sync_display_name">
            <div class="dtk-attr-row">
              <label class="dtk-attr-label">{{ t('admin.settings.dingtalk.syncDisplayNameTarget') }}</label>
              <input
                :value="local.dingtalk_connect_sync_display_name_attr_key"
                type="text"
                placeholder="dingtalk_name"
                class="input text-sm dtk-attr-input"
                @input="set('dingtalk_connect_sync_display_name_attr_key', ($event.target as HTMLInputElement).value)"
              />
            </div>
            <div class="dtk-attr-row">
              <label class="dtk-attr-label">{{ t('admin.settings.dingtalk.syncAttrDisplayName') }}</label>
              <input
                :value="local.dingtalk_connect_sync_display_name_attr_name"
                type="text"
                placeholder="钉钉姓名"
                class="input text-sm dtk-attr-input"
                @input="set('dingtalk_connect_sync_display_name_attr_name', ($event.target as HTMLInputElement).value)"
              />
            </div>
            <p class="dtk-field-hint">{{ t('admin.settings.dingtalk.syncDisplayNameTargetHint') }}</p>
          </template>
        </div>

        <!-- Sync Corp Email -->
        <div class="dtk-border-top dtk-sync-block">
          <div class="dtk-row-between">
            <div>
              <label class="dtk-label">{{ t('admin.settings.dingtalk.syncCorpEmail') }}</label>
              <p class="dtk-hint">{{ t('admin.settings.dingtalk.syncCorpEmailHint') }}</p>
              <p class="dtk-hint dtk-hint--warn">{{ t('admin.settings.dingtalk.syncCorpEmailPermissionHint') }}</p>
            </div>
            <Toggle :model-value="!!local.dingtalk_connect_sync_corp_email" @update:model-value="set('dingtalk_connect_sync_corp_email', $event)" />
          </div>
          <template v-if="local.dingtalk_connect_sync_corp_email">
            <div class="dtk-attr-row">
              <label class="dtk-attr-label">{{ t('admin.settings.dingtalk.syncCorpEmailTarget') }}</label>
              <input
                :value="local.dingtalk_connect_sync_corp_email_attr_key"
                type="text"
                placeholder="dingtalk_email"
                class="input text-sm dtk-attr-input"
                @input="set('dingtalk_connect_sync_corp_email_attr_key', ($event.target as HTMLInputElement).value)"
              />
            </div>
            <div class="dtk-attr-row">
              <label class="dtk-attr-label">{{ t('admin.settings.dingtalk.syncAttrDisplayName') }}</label>
              <input
                :value="local.dingtalk_connect_sync_corp_email_attr_name"
                type="text"
                placeholder="钉钉企业邮箱"
                class="input text-sm dtk-attr-input"
                @input="set('dingtalk_connect_sync_corp_email_attr_name', ($event.target as HTMLInputElement).value)"
              />
            </div>
            <p class="dtk-field-hint">{{ t('admin.settings.dingtalk.syncCorpEmailTargetHint') }}</p>
          </template>
        </div>

        <!-- Sync Department -->
        <div class="dtk-border-top dtk-sync-block">
          <div class="dtk-row-between">
            <div>
              <label class="dtk-label">{{ t('admin.settings.dingtalk.syncDept') }}</label>
              <p class="dtk-hint">{{ t('admin.settings.dingtalk.syncDeptHint') }}</p>
              <p class="dtk-hint dtk-hint--warn">{{ t('admin.settings.dingtalk.syncDeptPermissionHint') }}</p>
            </div>
            <Toggle :model-value="!!local.dingtalk_connect_sync_dept" @update:model-value="set('dingtalk_connect_sync_dept', $event)" />
          </div>
          <template v-if="local.dingtalk_connect_sync_dept">
            <div class="dtk-attr-row">
              <label class="dtk-attr-label">{{ t('admin.settings.dingtalk.syncDeptTarget') }}</label>
              <input
                :value="local.dingtalk_connect_sync_dept_attr_key"
                type="text"
                placeholder="dingtalk_department"
                class="input text-sm dtk-attr-input"
                @input="set('dingtalk_connect_sync_dept_attr_key', ($event.target as HTMLInputElement).value)"
              />
            </div>
            <div class="dtk-attr-row">
              <label class="dtk-attr-label">{{ t('admin.settings.dingtalk.syncAttrDisplayName') }}</label>
              <input
                :value="local.dingtalk_connect_sync_dept_attr_name"
                type="text"
                placeholder="钉钉部门"
                class="input text-sm dtk-attr-input"
                @input="set('dingtalk_connect_sync_dept_attr_name', ($event.target as HTMLInputElement).value)"
              />
            </div>
            <p class="dtk-field-hint">{{ t('admin.settings.dingtalk.syncDeptTargetHint') }}</p>
          </template>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'

const { t } = useI18n()

const props = defineProps<{
  settings: Record<string, unknown>
  formValues?: Record<string, unknown>
}>()

const emit = defineEmits<{
  'update:field': [key: string, value: unknown]
}>()

// Prefer live form dirty-state over persisted settings
const activeSource = () => props.formValues ?? props.settings

// All DingTalk keys mirrored locally for reactive rendering
interface DingtalkLocal {
  dingtalk_connect_enabled: boolean
  dingtalk_connect_client_id: string
  dingtalk_connect_client_secret: string
  dingtalk_connect_client_secret_configured: boolean
  dingtalk_connect_redirect_url: string
  dingtalk_connect_corp_restriction_policy: string
  dingtalk_connect_internal_corp_id: string
  dingtalk_connect_bypass_registration: boolean
  dingtalk_connect_sync_display_name: boolean
  dingtalk_connect_sync_display_name_attr_key: string
  dingtalk_connect_sync_display_name_attr_name: string
  dingtalk_connect_sync_corp_email: boolean
  dingtalk_connect_sync_corp_email_attr_key: string
  dingtalk_connect_sync_corp_email_attr_name: string
  dingtalk_connect_sync_dept: boolean
  dingtalk_connect_sync_dept_attr_key: string
  dingtalk_connect_sync_dept_attr_name: string
}

function pick(src: Record<string, unknown>): DingtalkLocal {
  return {
    dingtalk_connect_enabled: !!(src['dingtalk_connect_enabled'] ?? false),
    dingtalk_connect_client_id: (src['dingtalk_connect_client_id'] as string) ?? '',
    dingtalk_connect_client_secret: (src['dingtalk_connect_client_secret'] as string) ?? '',
    dingtalk_connect_client_secret_configured: !!(src['dingtalk_connect_client_secret_configured'] ?? false),
    dingtalk_connect_redirect_url: (src['dingtalk_connect_redirect_url'] as string) ?? '',
    dingtalk_connect_corp_restriction_policy: (src['dingtalk_connect_corp_restriction_policy'] as string) ?? 'none',
    dingtalk_connect_internal_corp_id: (src['dingtalk_connect_internal_corp_id'] as string) ?? '',
    dingtalk_connect_bypass_registration: !!(src['dingtalk_connect_bypass_registration'] ?? false),
    dingtalk_connect_sync_display_name: !!(src['dingtalk_connect_sync_display_name'] ?? false),
    dingtalk_connect_sync_display_name_attr_key: (src['dingtalk_connect_sync_display_name_attr_key'] as string) ?? 'dingtalk_name',
    dingtalk_connect_sync_display_name_attr_name: (src['dingtalk_connect_sync_display_name_attr_name'] as string) ?? '钉钉姓名',
    dingtalk_connect_sync_corp_email: !!(src['dingtalk_connect_sync_corp_email'] ?? false),
    dingtalk_connect_sync_corp_email_attr_key: (src['dingtalk_connect_sync_corp_email_attr_key'] as string) ?? 'dingtalk_email',
    dingtalk_connect_sync_corp_email_attr_name: (src['dingtalk_connect_sync_corp_email_attr_name'] as string) ?? '钉钉企业邮箱',
    dingtalk_connect_sync_dept: !!(src['dingtalk_connect_sync_dept'] ?? false),
    dingtalk_connect_sync_dept_attr_key: (src['dingtalk_connect_sync_dept_attr_key'] as string) ?? 'dingtalk_department',
    dingtalk_connect_sync_dept_attr_name: (src['dingtalk_connect_sync_dept_attr_name'] as string) ?? '钉钉部门',
  }
}

const local = reactive<DingtalkLocal>(pick(activeSource()))

// Re-sync when parent resets (discard/initial load)
watch(
  () => activeSource(),
  (src) => {
    const fresh = pick(src)
    for (const k of Object.keys(fresh) as (keyof DingtalkLocal)[]) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (local as any)[k] = (fresh as any)[k]
    }
  },
  { deep: true },
)

function set(key: keyof DingtalkLocal, value: unknown) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ;(local as any)[key] = value
  emit('update:field', key, value)
}

/**
 * When corp_restriction_policy changes away from internal_only, reset
 * dependent toggles to false (matching SettingsView.vue lines 9650-9652).
 */
function onCorpPolicyChange(policy: string) {
  set('dingtalk_connect_corp_restriction_policy', policy)
  if (policy !== 'internal_only') {
    if (local.dingtalk_connect_sync_corp_email) set('dingtalk_connect_sync_corp_email', false)
    if (local.dingtalk_connect_sync_display_name) set('dingtalk_connect_sync_display_name', false)
    if (local.dingtalk_connect_sync_dept) set('dingtalk_connect_sync_dept', false)
  }
}
</script>

<style scoped>
.dtk-body { padding: 16px 20px; display: flex; flex-direction: column; gap: 16px; }

.dtk-row-between {
  display: flex; align-items: center; justify-content: space-between; gap: 16px;
}

.dtk-label { font-size: 13px; font-weight: 500; color: var(--ink-0, #E8EBF0); }
.dtk-hint { font-size: 11.5px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 2px 0 0; }
.dtk-hint--warn { color: var(--amber, #F5A623); }

.dtk-expanded { display: flex; flex-direction: column; gap: 16px; border-top: 1px solid var(--line-1, #2F3540); padding-top: 16px; }

.dtk-field { display: flex; flex-direction: column; gap: 4px; }
.dtk-field-label { font-size: 12px; font-weight: 500; color: var(--ink-1, #97A0AF); }
.dtk-field-hint { font-size: 11px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 0; }
.dtk-field-hint--top { margin-bottom: 8px; }

.dtk-border-top { border-top: 1px solid var(--line-1, #2F3540); padding-top: 16px; }

.dtk-radio-group { display: flex; flex-direction: column; gap: 8px; }
.dtk-radio-item { display: flex; align-items: center; gap: 10px; cursor: pointer; }

.dtk-sync-block { display: flex; flex-direction: column; gap: 10px; }

.dtk-attr-row { display: flex; align-items: center; gap: 8px; }
.dtk-attr-label {
  font-size: 12px; color: var(--ink-2, #5C6470);
  white-space: nowrap; min-width: 5rem;
}
.dtk-attr-input { flex: 1; max-width: 20rem; }
</style>
