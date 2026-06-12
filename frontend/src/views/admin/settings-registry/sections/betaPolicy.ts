/**
 * Beta Policy section — gateway tab.
 * Self-contained: mounts its own GET/PUT via getBetaPolicySettings / updateBetaPolicySettings.
 * Fields: rules[].beta_token, .action, .scope, .error_message,
 *         .model_whitelist[], .fallback_action, .fallback_error_message
 *
 * Source: SettingsView.vue lines 794–1071 (betaPolicyForm)
 */
import type { SettingsSection } from '../types'
import BetaPolicySection from '../special/BetaPolicySection.vue'

const betaPolicy: SettingsSection = {
  id: 'gateway.betaPolicy',
  tab: 'gateway',
  title: 'admin.settings.betaPolicy.title',
  description: 'admin.settings.betaPolicy.description',
  fields: [],
  component: BetaPolicySection,
}

export default betaPolicy
