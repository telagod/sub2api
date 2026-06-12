/**
 * Users tab — new-user defaults, auth-source grant defaults.
 * Keys sourced from SettingsView.vue activeTab === 'users' block (lines 3058–3671).
 *
 * Flat form. bindings in this tab:
 *   default_balance, default_concurrency, default_user_rpm_limit              ✓ migrated
 *   force_email_on_third_party_signup                                         ✓ migrated (authSourceDefaults card)
 *
 * Backlog — not schema-able:
 *   form.default_subscriptions          — dynamic multi-item list with custom GroupBadge/GroupOptionItem Select widgets
 *   form.default_platform_quotas[p].*  — 4×3 quota matrix (anthropic/openai/gemini/antigravity × daily/weekly/monthly)
 *   authSourceDefaults[source].*       — per-auth-source reactive object with nested subscriptions + quota override matrix
 * All three exceed FieldRenderer capacity and require dedicated section.component components.
 */
import type { SettingsSection } from '../types'

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
 * Auth source defaults — flat gate that controls whether third-party signups
 * must supply an email address.
 *
 * Sources: SettingsView.vue activeTab === 'users', authSourceDefaults card (line 3356).
 * The per-auth-source reactive matrix (grant_on_signup, balance, concurrency,
 * subscriptions, platform_quotas) is non-flat and backlogged for section.component.
 */
const authSourceDefaults: SettingsSection = {
  id: 'users.authSourceDefaults',
  tab: 'users',
  title: 'admin.settings.authSourceDefaults.title',
  description: 'admin.settings.authSourceDefaults.description',
  fields: [
    {
      key: 'force_email_on_third_party_signup',
      label: 'admin.settings.authSourceDefaults.requireEmailLabel',
      type: 'switch',
      help: 'admin.settings.authSourceDefaults.requireEmailHint',
    },
  ],
}

export default [userDefaults, authSourceDefaults] as SettingsSection[]
