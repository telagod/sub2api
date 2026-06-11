/**
 * Settings Registry — auto-discovers all sections/*.ts files via import.meta.glob.
 * Adding a new section = dropping a new file in sections/, zero central changes.
 */
import type { SettingsSection, TabId } from './types'

// Eagerly import every section module
const sectionModules = import.meta.glob<{ default: SettingsSection | SettingsSection[] }>(
  './sections/*.ts',
  { eager: true },
)

function collectSections(): SettingsSection[] {
  const all: SettingsSection[] = []
  for (const mod of Object.values(sectionModules)) {
    const exported = mod.default
    if (Array.isArray(exported)) {
      all.push(...exported)
    } else if (exported) {
      all.push(exported)
    }
  }
  return all
}

/** All sections in discovery order */
export const allSections: SettingsSection[] = collectSections()

/** Sections grouped by tab */
export function getSectionsByTab(): Map<TabId, SettingsSection[]> {
  const map = new Map<TabId, SettingsSection[]>()
  for (const section of allSections) {
    const existing = map.get(section.tab) ?? []
    existing.push(section)
    map.set(section.tab, existing)
  }
  return map
}

/** All tab ids that have at least one section */
export function getActiveTabs(): TabId[] {
  return [...getSectionsByTab().keys()]
}
