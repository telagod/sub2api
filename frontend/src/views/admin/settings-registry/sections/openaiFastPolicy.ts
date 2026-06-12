/**
 * OpenAI Fast/Flex Policy section — gateway tab.
 * Global form mode: reads from props.settings['openai_fast_policy_settings'] (JSON object),
 * emits update:field('openai_fast_policy_settings', JSON.stringify({rules:[...]})) —
 * saved via the global settings save bar, matching the pattern of WechatConnectSection.
 *
 * Fields per rule: service_tier, action, scope, error_message,
 *                  model_whitelist[], fallback_action, fallback_error_message
 *
 * Source: SettingsView.vue lines 1073–1351 (openaiFastPolicyForm)
 */
import type { SettingsSection } from '../types'
import OpenaiFastPolicySection from '../special/OpenaiFastPolicySection.vue'

const openaiFastPolicy: SettingsSection = {
  id: 'gateway.openaiFastPolicy',
  tab: 'gateway',
  title: 'admin.settings.openaiFastPolicy.title',
  description: 'admin.settings.openaiFastPolicy.description',
  fields: [],
  component: OpenaiFastPolicySection,
}

export default openaiFastPolicy
