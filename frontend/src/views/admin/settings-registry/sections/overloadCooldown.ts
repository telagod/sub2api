/**
 * Gateway tab — Overload Cooldown (529) section.
 *
 * Self-contained component: mounts its own GET/PUT lifecycle against
 *   GET  /admin/settings/overload-cooldown
 *   PUT  /admin/settings/overload-cooldown
 *
 * Fields covered (SettingsView.vue overloadCooldownForm ~239 / ~6874):
 *   enabled          — Toggle
 *   cooldown_minutes — number input (min 1, max 120), shown when enabled
 */
import { defineAsyncComponent } from 'vue'
import type { SettingsSection } from '../types'

const overloadCooldown: SettingsSection = {
  id: 'gateway.overloadCooldown',
  tab: 'gateway',
  title: 'admin.settings.overloadCooldown.title',
  description: 'admin.settings.overloadCooldown.description',
  fields: [],
  component: defineAsyncComponent(
    () => import('../special/OverloadCooldownSection.vue'),
  ),
}

export default overloadCooldown
