/**
 * Affiliate custom-users CRUD panel.
 *
 * This section renders entirely through section.component because it requires
 * async search, pagination, and modal state that cannot be expressed as
 * SettingsField rows. It is intentionally placed after features.ts affiliate
 * section so it appears beneath the global-rate fields in the Features tab.
 *
 * The component is self-contained (independent CRUD endpoints — no global save
 * bar involvement) so no emit('update:field') is emitted.
 */
import type { SettingsSection } from '../types'
import AffiliateCustomUsersSection from '../special/AffiliateCustomUsersSection.vue'

const affiliateCustomUsers: SettingsSection = {
  id: 'features.affiliateCustomUsers',
  tab: 'features',
  title: 'admin.settings.features.affiliate.customUsers.title',
  description: 'admin.settings.features.affiliate.customUsers.description',
  fields: [],
  component: AffiliateCustomUsersSection,
}

export default affiliateCustomUsers
