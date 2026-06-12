/**
 * General tab — Custom Menu Items + Custom Endpoints.
 *
 * Keys covered:
 *   custom_menu_items   — sortable list (id, label, icon_svg, url, visibility, sort_order)
 *   custom_endpoints    — list of (name, endpoint, description)
 *
 * Both are object-array values; the UI is too rich for FieldRenderer (reorder,
 * SVG upload, delete), so a dedicated CustomMenuSection component is used.
 *
 * Source of truth: SettingsView.vue, activeTab === 'general', lines 4566–5000
 * and form.custom_menu_items / form.custom_endpoints init + save handlers.
 */
import { defineAsyncComponent } from 'vue'
import type { SettingsSection } from '../types'

const customMenu: SettingsSection = {
  id: 'general.customMenu',
  tab: 'general',
  title: 'admin.settings.customMenu.title',
  fields: [],
  component: defineAsyncComponent(
    () => import('../special/CustomMenuSection.vue'),
  ),
}

export default customMenu
