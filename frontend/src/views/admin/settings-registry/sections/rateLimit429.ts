/**
 * Gateway tab — Rate Limit 429 Cooldown section.
 *
 * Self-contained component: mounts its own GET/PUT lifecycle against
 *   GET  /admin/settings/rate-limit-429-cooldown
 *   PUT  /admin/settings/rate-limit-429-cooldown
 *
 * Fields covered (SettingsView.vue rateLimit429CooldownForm ~340 / ~6882):
 *   enabled          — Toggle
 *   cooldown_seconds — number input (min 1, max 7200), shown when enabled
 */
import { defineAsyncComponent } from 'vue'
import type { SettingsSection } from '../types'

const rateLimit429: SettingsSection = {
  id: 'gateway.rateLimit429',
  tab: 'gateway',
  title: 'admin.settings.rateLimit429Cooldown.title',
  description: 'admin.settings.rateLimit429Cooldown.description',
  fields: [],
  component: defineAsyncComponent(
    () => import('../special/RateLimit429Section.vue'),
  ),
}

export default rateLimit429
