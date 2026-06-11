/**
 * Features tab — channel monitor, available channels, risk control, affiliate.
 * Keys sourced from SettingsView.vue activeTab === 'features' block (lines 5185–5703).
 *
 * Backlog: affiliate customUsers table (CRUD + user-search + pagination,
 * lines 5399–5529) exceeds FieldRenderer capacity; needs section.component.
 */
import type { SettingsSection } from '../types'

const channelMonitor: SettingsSection = {
  id: 'features.channelMonitor',
  tab: 'features',
  title: 'admin.settings.features.channelMonitor.title',
  description: 'admin.settings.features.channelMonitor.description',
  fields: [
    {
      key: 'channel_monitor_enabled',
      label: 'admin.settings.features.channelMonitor.enabled',
      type: 'switch',
      help: 'admin.settings.features.channelMonitor.enabledHint',
    },
    {
      key: 'channel_monitor_default_interval_seconds',
      label: 'admin.settings.features.channelMonitor.defaultInterval',
      type: 'number',
      help: 'admin.settings.features.channelMonitor.defaultIntervalHint',
      showWhen: (v) => !!v['channel_monitor_enabled'],
    },
  ],
}

const availableChannels: SettingsSection = {
  id: 'features.availableChannels',
  tab: 'features',
  title: 'admin.settings.features.availableChannels.title',
  description: 'admin.settings.features.availableChannels.description',
  fields: [
    {
      key: 'available_channels_enabled',
      label: 'admin.settings.features.availableChannels.enabled',
      type: 'switch',
      help: 'admin.settings.features.availableChannels.enabledHint',
    },
  ],
}

const riskControl: SettingsSection = {
  id: 'features.riskControl',
  tab: 'features',
  title: 'admin.settings.features.riskControl.title',
  description: 'admin.settings.features.riskControl.description',
  fields: [
    {
      key: 'risk_control_enabled',
      label: 'admin.settings.features.riskControl.enabled',
      type: 'switch',
      help: 'admin.settings.features.riskControl.enabledHint',
    },
  ],
}

// Affiliate global rate config (flat keys).
// The per-user CRUD table (customUsers, lines 5399–5529) is backlogged for
// a section.component migration — it requires async search, pagination, and
// modal state that cannot be expressed as SettingsField.
const affiliate: SettingsSection = {
  id: 'features.affiliate',
  tab: 'features',
  title: 'admin.settings.features.affiliate.title',
  description: 'admin.settings.features.affiliate.description',
  fields: [
    {
      key: 'affiliate_enabled',
      label: 'admin.settings.features.affiliate.enabled',
      type: 'switch',
      help: 'admin.settings.features.affiliate.enabledHint',
    },
    {
      key: 'affiliate_rebate_rate',
      label: 'admin.settings.features.affiliate.rebateRate',
      type: 'number',
      help: 'admin.settings.features.affiliate.rebateRateHint',
      showWhen: (v) => !!v['affiliate_enabled'],
    },
    {
      key: 'affiliate_rebate_freeze_hours',
      label: 'admin.settings.features.affiliate.freezeHours',
      type: 'number',
      help: 'admin.settings.features.affiliate.freezeHoursDesc',
      showWhen: (v) => !!v['affiliate_enabled'],
    },
    {
      key: 'affiliate_rebate_duration_days',
      label: 'admin.settings.features.affiliate.durationDays',
      type: 'number',
      help: 'admin.settings.features.affiliate.durationDaysDesc',
      showWhen: (v) => !!v['affiliate_enabled'],
    },
    {
      key: 'affiliate_rebate_per_invitee_cap',
      label: 'admin.settings.features.affiliate.perInviteeCap',
      type: 'number',
      help: 'admin.settings.features.affiliate.perInviteeCapDesc',
      showWhen: (v) => !!v['affiliate_enabled'],
    },
  ],
}

export default [channelMonitor, availableChannels, riskControl, affiliate] as SettingsSection[]
