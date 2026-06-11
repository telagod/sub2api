/**
 * Backup tab — delegates entirely to the existing BackupView component.
 *
 * Keys sourced from SettingsView.vue activeTab === 'backup' region (lines 6620–6622).
 *
 * form. bindings in old view: 0
 *   The backup tab contains only <BackupSettings /> (BackupView.vue).
 *   No flat key/value form fields in this tab — BackupView manages its own state.
 *
 * Fields covered here: 0 schema fields
 * Escape hatch section: backup.view (BackupView component)
 * Total: 0 / 0 flat keys ✓ (component fully handled by BackupView)
 */
import { defineAsyncComponent } from 'vue'
import type { SettingsSection } from '../types'

const backupView: SettingsSection = {
  id: 'backup.view',
  tab: 'backup',
  title: '数据备份与恢复',
  description: '管理系统数据的备份与还原操作。',
  fields: [],
  component: defineAsyncComponent(
    () => import('@/views/admin/BackupView.vue'),
  ),
}

export default [backupView] as SettingsSection[]
