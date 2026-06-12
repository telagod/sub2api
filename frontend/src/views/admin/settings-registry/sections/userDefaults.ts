/**
 * Users tab — additional sections that require custom components.
 *
 * Block A-1: default_subscriptions
 *   Dynamic multi-row list. Each row: group_id (GroupBadge Select) + validity_days.
 *   Source: SettingsView.vue ~3128–3264.
 *
 * Block A-2: default_platform_quotas
 *   4 platforms × 3 windows (daily/weekly/monthly) matrix.
 *   Serialized via sanitizePlatformQuotasMap before emit.
 *   Source: SettingsView.vue ~3266–3329.
 *
 * Both blocks share one section.component so they render as a single card
 * (mirrors the original SettingsView layout where they are in the same card).
 *
 * Global form mode: component reads props.settings, emits('update:field') per change,
 * re-syncs via watch(props.settings). No self-contained save button.
 */
import type { SettingsSection } from '../types'
import UserDefaultsSection from '../special/UserDefaultsSection.vue'

const userDefaults: SettingsSection = {
  id: 'users.defaultSubscriptionsAndQuotas',
  tab: 'users',
  // Reuses existing i18n keys — "用户默认设置" card title
  title: 'admin.settings.defaults.defaultSubscriptions',
  description: 'admin.settings.defaults.defaultSubscriptionsHint',
  fields: [],
  component: UserDefaultsSection,
}

export default userDefaults
