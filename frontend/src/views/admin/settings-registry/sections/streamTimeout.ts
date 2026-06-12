/**
 * Gateway tab — Stream Timeout section.
 *
 * Self-contained component: mounts its own GET/PUT lifecycle against
 *   GET  /admin/settings/stream-timeout
 *   PUT  /admin/settings/stream-timeout
 *
 * Fields covered (SettingsView.vue streamTimeoutForm ~449 / ~6890):
 *   enabled                  — Toggle
 *   action                   — select: temp_unsched | error | none
 *   temp_unsched_minutes     — number input (min 1, max 60), shown when action === 'temp_unsched'
 *   threshold_count          — number input (min 1, max 10)
 *   threshold_window_minutes — number input (min 1, max 60)
 *
 * Conditional rendering: temp_unsched_minutes only shown when action === 'temp_unsched',
 * matching the v-if in SettingsView.vue line ~486.
 */
import { defineAsyncComponent } from 'vue'
import type { SettingsSection } from '../types'

const streamTimeout: SettingsSection = {
  id: 'gateway.streamTimeout',
  tab: 'gateway',
  title: 'admin.settings.streamTimeout.title',
  description: 'admin.settings.streamTimeout.description',
  fields: [],
  component: defineAsyncComponent(
    () => import('../special/StreamTimeoutSection.vue'),
  ),
}

export default streamTimeout
