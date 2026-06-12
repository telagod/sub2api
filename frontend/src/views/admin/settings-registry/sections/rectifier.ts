/**
 * Rectifier section — gateway tab.
 * Self-contained: mounts its own GET/PUT via getRectifierSettings / updateRectifierSettings.
 * Fields: enabled, thinking_signature_enabled, thinking_budget_enabled,
 *         apikey_signature_enabled, apikey_signature_patterns[]
 *
 * Source: SettingsView.vue lines 595–792 (rectifierForm)
 */
import type { SettingsSection } from '../types'
import RectifierSection from '../special/RectifierSection.vue'

const rectifier: SettingsSection = {
  id: 'gateway.rectifier',
  tab: 'gateway',
  title: 'admin.settings.rectifier.title',
  description: 'admin.settings.rectifier.description',
  fields: [],
  component: RectifierSection,
}

export default rectifier
