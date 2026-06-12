/**
 * Gateway tab — Web Search Emulation section.
 *
 * Self-contained component: loads from getWebSearchEmulationConfig(),
 * saves via updateWebSearchEmulationConfig(), and calls testWebSearchEmulation()
 * internally. Does NOT participate in the global flat-form save cycle.
 *
 * Source: SettingsView.vue lines 3966–4401 (webSearchConfig reactive object,
 * loadWebSearchConfig / saveWebSearchConfig, webSearch* helpers).
 */
import type { SettingsSection } from '../types'
import WebSearchEmulationSection from '../special/WebSearchEmulationSection.vue'

const webSearchEmulation: SettingsSection = {
  id: 'gateway.webSearchEmulation',
  tab: 'gateway',
  title: 'admin.settings.webSearchEmulation.title',
  description: 'admin.settings.webSearchEmulation.description',
  fields: [],
  component: WebSearchEmulationSection,
}

export default webSearchEmulation
