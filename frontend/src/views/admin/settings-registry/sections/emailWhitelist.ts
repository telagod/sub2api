/**
 * Security tab — Email Suffix Whitelist.
 *
 * Keys covered:
 *   registration_email_suffix_whitelist  — string[] of canonical suffixes
 *                                          (e.g. "@example.com", "*.edu.cn")
 *
 * The tag-chip input (Enter/comma/space submit, paste bulk-parse, Backspace
 * to pop last chip) cannot be expressed as a FieldRenderer field; a dedicated
 * EmailSuffixWhitelistSection component is used instead.
 *
 * Wire-format note: the component stores domains in display form
 * (normalised, no "@" prefix) internally, and emits the @-prefixed canonical
 * form on every change to match the save transform in SettingsView.vue (line 8160).
 *
 * Source of truth: SettingsView.vue, activeTab === 'security',
 * lines 1399–1465 (template) + lines 7477–7548 (handlers).
 */
import { defineAsyncComponent } from 'vue'
import type { SettingsSection } from '../types'

const emailWhitelist: SettingsSection = {
  id: 'security.emailWhitelist',
  tab: 'security',
  title: 'admin.settings.registration.emailSuffixWhitelist',
  description: 'admin.settings.registration.emailSuffixWhitelistHint',
  fields: [],
  component: defineAsyncComponent(
    () => import('../special/EmailSuffixWhitelistSection.vue'),
  ),
}

export default emailWhitelist
