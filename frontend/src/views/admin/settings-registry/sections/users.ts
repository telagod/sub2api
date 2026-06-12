/**
 * Users tab — new-user defaults, auth-source grant defaults.
 * Keys sourced from SettingsView.vue activeTab === 'users' block (lines 3058–3671).
 *
 * Flat form. bindings in this tab:
 *   default_balance, default_concurrency, default_user_rpm_limit              ✓ migrated
 *   force_email_on_third_party_signup                                         ✓ migrated (authSourceDefaults card)
 *
 * Migrated via section.component:
 *   form.default_subscriptions          — UserDefaultsSection.vue
 *   form.default_platform_quotas[p].*  — UserDefaultsSection.vue
 *   authSourceDefaults[source].*       — AuthSourceDefaultsSection.vue
 *     (grant_on_signup, balance, concurrency, grant_on_first_bind,
 *      subscriptions[], platform_quotas × 7 auth sources)
 */
import type { SettingsSection } from '../types'
import AuthSourceDefaultsSection from '../special/AuthSourceDefaultsSection.vue'

const userDefaults: SettingsSection = {
  id: 'users.defaults',
  tab: 'users',
  title: 'admin.settings.defaults.title',
  description: 'admin.settings.defaults.description',
  fields: [
    {
      key: 'default_balance',
      label: 'admin.settings.defaults.defaultBalance',
      type: 'number',
      placeholder: '0.00',
      help: 'admin.settings.defaults.defaultBalanceHint',
    },
    {
      key: 'default_concurrency',
      label: 'admin.settings.defaults.defaultConcurrency',
      type: 'number',
      placeholder: '1',
      help: 'admin.settings.defaults.defaultConcurrencyHint',
    },
    {
      key: 'default_user_rpm_limit',
      label: 'admin.settings.defaults.defaultUserRpmLimit',
      type: 'number',
      placeholder: '0',
      help: 'admin.settings.defaults.defaultUserRpmLimitHint',
    },
  ],
}

/**
 * Auth source defaults — full per-auth-source matrix:
 *   require email gate (force_email_on_third_party_signup)  → flat field below
 *   per-source: grant_on_signup, balance, concurrency, grant_on_first_bind,
 *               subscriptions[], platform_quotas             → AuthSourceDefaultsSection.vue
 *
 * Sources: SettingsView.vue activeTab === 'users', authSourceDefaults card (lines 3338–3669).
 * Serialization: appendAuthSourceDefaultsToUpdateRequest + sanitizePlatformQuotasMap
 *   (imported inside AuthSourceDefaultsSection.vue, not reimplemented here).
 */
const authSourceDefaults: SettingsSection = {
  id: 'users.authSourceDefaults',
  tab: 'users',
  title: 'admin.settings.authSourceDefaults.title',
  description: 'admin.settings.authSourceDefaults.description',
  // force_email_on_third_party_signup and all per-source fields are rendered
  // inside AuthSourceDefaultsSection.vue via emit('update:field', ...)
  fields: [],
  component: AuthSourceDefaultsSection,
}

export default [userDefaults, authSourceDefaults] as SettingsSection[]
