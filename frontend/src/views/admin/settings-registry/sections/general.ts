/**
 * General tab — site branding, OEM config, display preferences.
 * Keys sourced from SettingsView.vue activeTab === 'general' blocks.
 */
import type { SettingsSection } from '../types'

const siteBranding: SettingsSection = {
  id: 'general.site',
  tab: 'general',
  title: 'admin.settings.site.title',
  description: 'admin.settings.site.description',
  fields: [
    {
      key: 'backend_mode_enabled',
      label: 'admin.settings.site.backendMode',
      type: 'switch',
      help: 'admin.settings.site.backendModeDescription',
    },
    {
      key: 'site_name',
      label: 'admin.settings.site.siteName',
      type: 'text',
      placeholder: 'admin.settings.site.siteNamePlaceholder',
      help: 'admin.settings.site.siteNameHint',
    },
    {
      key: 'site_subtitle',
      label: 'admin.settings.site.siteSubtitle',
      type: 'text',
      placeholder: 'admin.settings.site.siteSubtitlePlaceholder',
      help: 'admin.settings.site.siteSubtitleHint',
    },
    {
      key: 'api_base_url',
      label: 'admin.settings.site.apiBaseUrl',
      type: 'text',
      placeholder: 'admin.settings.site.apiBaseUrlPlaceholder',
      help: 'admin.settings.site.apiBaseUrlHint',
    },
    {
      key: 'frontend_url',
      label: 'admin.settings.registration.frontendUrl',
      type: 'text',
      placeholder: 'admin.settings.registration.frontendUrlPlaceholder',
      help: 'admin.settings.registration.frontendUrlHint',
    },
    {
      key: 'contact_info',
      label: 'admin.settings.site.contactInfo',
      type: 'text',
      placeholder: 'admin.settings.site.contactInfoPlaceholder',
      help: 'admin.settings.site.contactInfoHint',
    },
    {
      key: 'doc_url',
      label: 'admin.settings.site.docUrl',
      type: 'text',
      placeholder: 'admin.settings.site.docUrlPlaceholder',
      help: 'admin.settings.site.docUrlHint',
    },
    {
      key: 'site_logo',
      label: 'admin.settings.site.siteLogo',
      type: 'image',
      help: 'admin.settings.site.logoHint',
    },
    {
      key: 'home_content',
      label: 'admin.settings.site.homeContent',
      type: 'textarea',
      placeholder: 'admin.settings.site.homeContentPlaceholder',
      help: 'admin.settings.site.homeContentHint',
    },
    {
      key: 'hide_ccs_import_button',
      label: 'admin.settings.site.hideCcsImportButton',
      type: 'switch',
      help: 'admin.settings.site.hideCcsImportButtonHint',
    },
  ],
}

const tablePreferences: SettingsSection = {
  id: 'general.table',
  tab: 'general',
  title: 'admin.settings.site.tablePreferencesTitle',
  description: 'admin.settings.site.tablePreferencesDescription',
  fields: [
    {
      key: 'table_default_page_size',
      label: 'admin.settings.site.tableDefaultPageSize',
      type: 'number',
      help: 'admin.settings.site.tableDefaultPageSizeHint',
    },
    {
      // Stored as number[], but rendered as comma-separated text via FieldRenderer special handling
      key: 'table_page_size_options',
      label: 'admin.settings.site.tablePageSizeOptions',
      type: 'text',
      placeholder: 'admin.settings.site.tablePageSizeOptionsPlaceholder',
      help: 'admin.settings.site.tablePageSizeOptionsHint',
    },
  ],
}

// custom_menu_items and custom_endpoints are handled by general.customMenu
// (CustomMenuSection.vue) which is registered in sections/customMenu.ts.

export default [siteBranding, tablePreferences] as SettingsSection[]
