/**
 * Users tab — new-user defaults, auth-source grant defaults.
 * Keys sourced from SettingsView.vue activeTab === 'users' block (lines 3058–3671).
 *
 * Flat form. bindings in this tab:
 *   default_balance, default_concurrency, default_user_rpm_limit              ✓ migrated
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

// force_email_on_third_party_signup is managed by security.ts (security.registration section).
// The authSourceDefaults card (grant_on_signup, per-source balance /
// concurrency / subscriptions / platform_quotas) is non-flat reactive state and
// is backlogged for a section.component migration.

export default [userDefaults] as SettingsSection[]
