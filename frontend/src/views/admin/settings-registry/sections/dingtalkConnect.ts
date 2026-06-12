/**
 * Security tab — DingTalk Connect section.
 *
 * Keys covered (all form bindings from SettingsView.vue dingtalk block):
 *   dingtalk_connect_enabled
 *   dingtalk_connect_client_id
 *   dingtalk_connect_client_secret          (masked / configured sentinel)
 *   dingtalk_connect_client_secret_configured
 *   dingtalk_connect_redirect_url
 *   dingtalk_connect_corp_restriction_policy  (radio: none | internal_only)
 *   dingtalk_connect_internal_corp_id         (form default, not rendered in legacy UI)
 *   dingtalk_connect_bypass_registration      (showWhen: internal_only)
 *   dingtalk_connect_sync_display_name        (showWhen: internal_only)
 *   dingtalk_connect_sync_display_name_attr_key
 *   dingtalk_connect_sync_display_name_attr_name
 *   dingtalk_connect_sync_corp_email          (showWhen: internal_only)
 *   dingtalk_connect_sync_corp_email_attr_key
 *   dingtalk_connect_sync_corp_email_attr_name
 *   dingtalk_connect_sync_dept                (showWhen: internal_only)
 *   dingtalk_connect_sync_dept_attr_key
 *   dingtalk_connect_sync_dept_attr_name
 *
 * The interaction complexity (corp_restriction_policy radio → cascading toggles →
 * conditional attr-key/attr-name pairs) exceeds FieldRenderer capacity.
 * DingtalkConnectSection.vue is the authoritative UI; fields[] is empty
 * so FieldRenderer emits nothing — all state flows through the component's
 * emit('update:field') calls.
 */
import { defineAsyncComponent } from 'vue'
import type { SettingsSection } from '../types'

const dingtalkConnect: SettingsSection = {
  id: 'security.dingtalk',
  tab: 'security',
  title: 'admin.settings.dingtalk.title',
  description: 'admin.settings.dingtalk.description',
  fields: [],
  component: defineAsyncComponent(
    () => import('../special/DingtalkConnectSection.vue'),
  ),
}

export default dingtalkConnect
