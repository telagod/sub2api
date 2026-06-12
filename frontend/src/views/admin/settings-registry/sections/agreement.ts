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
  title: 'admin.settings.agreement.configTitle',
  description: 'admin.settings.agreement.configDescription',
  fields: [
    {
      key: 'login_agreement_enabled',
      label: 'admin.settings.agreement.enabledLabel',
      type: 'switch',
      help: 'admin.settings.agreement.enabledHint',
    },
    {
      key: 'login_agreement_mode',
      label: 'admin.settings.agreement.modeLabel',
      type: 'select',
      options: [
        { value: 'modal', label: 'admin.settings.agreement.modeModal' },
        { value: 'checkbox', label: 'admin.settings.agreement.modeCheckbox' },
      ],
      help: 'admin.settings.agreement.modeHint',
      showWhen: (v) => !!v['login_agreement_enabled'],
    },
    {
      key: 'login_agreement_updated_at',
      label: 'admin.settings.agreement.updatedAtLabel',
      type: 'text',
      placeholder: 'YYYY-MM-DD',
      help: 'admin.settings.agreement.updatedAtHint',
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
  title: 'admin.settings.agreement.documentsTitle',
  description: 'admin.settings.agreement.documentsDescription',
  fields: [],
  component: defineAsyncComponent(
    () => import('../special/AgreementDocumentsSection.vue'),
  ),
}

export default [agreementConfig, agreementDocuments] as SettingsSection[]
