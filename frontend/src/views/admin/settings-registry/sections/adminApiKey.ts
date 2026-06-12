/**
 * Security tab — Admin API Key section.
 *
 * Self-contained component: loads from getAdminApiKey(), creates/regenerates
 * via regenerateAdminApiKey(), deletes via deleteAdminApiKey(). Full CRUD with
 * masked display and one-time new-key reveal. Does NOT participate in the
 * global flat-form save cycle.
 *
 * Source: SettingsView.vue lines 48–200 (adminApiKey template block),
 * lines 6864–6868 (state refs), lines 8540–8599 (CRUD functions).
 */
import type { SettingsSection } from '../types'
import AdminApiKeySection from '../special/AdminApiKeySection.vue'

const adminApiKey: SettingsSection = {
  id: 'security.adminApiKey',
  tab: 'security',
  title: 'admin.settings.adminApiKey.title',
  description: 'admin.settings.adminApiKey.description',
  fields: [],
  component: AdminApiKeySection,
}

export default adminApiKey
