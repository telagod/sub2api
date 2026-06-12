/**
 * WeChat Connect section — security tab.
 *
 * UI is too complex for FieldRenderer (three mode sub-panels, masked secrets,
 * mutual-exclusion logic between mp/mobile, UnionID warning) so it delegates
 * to a custom component.
 *
 * Keys covered (full parity with SettingsView.vue):
 *   wechat_connect_enabled
 *   wechat_connect_open_enabled
 *   wechat_connect_mp_enabled
 *   wechat_connect_mobile_enabled
 *   wechat_connect_open_app_id
 *   wechat_connect_open_app_secret          (masked, secret)
 *   wechat_connect_open_app_secret_configured
 *   wechat_connect_mp_app_id
 *   wechat_connect_mp_app_secret            (masked, secret)
 *   wechat_connect_mp_app_secret_configured
 *   wechat_connect_mobile_app_id
 *   wechat_connect_mobile_app_secret        (masked, secret)
 *   wechat_connect_mobile_app_secret_configured
 *   wechat_connect_mode
 *   wechat_connect_scopes
 *   wechat_connect_redirect_url
 *   wechat_connect_frontend_redirect_url
 *   wechat_connect_app_id                   (legacy passthrough)
 *   wechat_connect_app_secret               (legacy passthrough)
 *   wechat_connect_app_secret_configured    (legacy passthrough)
 */
import { defineAsyncComponent } from 'vue'
import type { SettingsSection } from '../types'

const wechatConnect: SettingsSection = {
  id: 'security.wechat_connect',
  tab: 'security',
  title: 'admin.settings.wechatConnect.title',
  description: 'admin.settings.wechatConnect.description',
  fields: [],
  component: defineAsyncComponent(
    () => import('../special/WechatConnectSection.vue'),
  ),
}

export default wechatConnect
