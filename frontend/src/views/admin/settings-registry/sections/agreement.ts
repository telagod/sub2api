/**
 * Agreement tab — login agreement gate, display mode, updated-at date,
 * and the Markdown document list.
 *
 * Keys sourced from SettingsView.vue activeTab === 'agreement' region (lines 4983–5181).
 *
 * form. bindings in old view: 4
 *   login_agreement_enabled, login_agreement_mode, login_agreement_updated_at,
 *   login_agreement_documents
 *
 * Covered by schema fields: 3 flat keys
 *   (login_agreement_documents carries complex per-item editor logic —
 *   handled via AgreementDocumentsSection component escape hatch; see special/)
 *
 * Fields covered here: login_agreement_enabled, login_agreement_mode,
 *   login_agreement_updated_at = 3 flat keys
 * Escape hatch section covers: login_agreement_documents = 1 complex key
 * Total: 4 / 4 ✓
 */
import { defineAsyncComponent } from 'vue'
import type { SettingsSection } from '../types'

const agreementConfig: SettingsSection = {
  id: 'agreement.config',
  tab: 'agreement',
  title: '登录条款确认',
  description:
    '控制登录页是否要求用户先阅读并同意服务条款、隐私政策或其他 Markdown 文档。',
  fields: [
    {
      key: 'login_agreement_enabled',
      label: '启用登录条款确认',
      type: 'switch',
      help: '开启后，登录页会在用户操作前显示条款确认。',
    },
    {
      key: 'login_agreement_mode',
      label: '展示形式',
      type: 'select',
      options: [
        { value: 'modal', label: '弹窗' },
        { value: 'checkbox', label: '复选框' },
      ],
      help: '弹窗：打开后用户拒绝则所有登录入口禁用；复选框：显示在登录按钮下方，未勾选前禁用。',
      showWhen: (v) => !!v['login_agreement_enabled'],
    },
    {
      key: 'login_agreement_updated_at',
      label: '条款更新日期',
      type: 'text',
      placeholder: 'YYYY-MM-DD',
      help: '日期或文档内容变化后，用户需重新同意。',
      showWhen: (v) => !!v['login_agreement_enabled'],
    },
  ],
}

/**
 * Agreement documents — complex interactive list (per-item title / slug / Markdown editor).
 * Delegates to AgreementDocumentsSection custom component.
 * Covers: login_agreement_documents
 */
const agreementDocuments: SettingsSection = {
  id: 'agreement.documents',
  tab: 'agreement',
  title: '协议文档',
  description:
    '文档名称可自定义，内容按 Markdown 保存。可参考：服务条款、使用政策、支持的国家和地区、服务特定条款。',
  fields: [],
  component: defineAsyncComponent(
    () => import('../special/AgreementDocumentsSection.vue'),
  ),
}

export default [agreementConfig, agreementDocuments] as SettingsSection[]
