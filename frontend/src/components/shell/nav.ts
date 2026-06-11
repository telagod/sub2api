/**
 * Quench Shell · 导航模型
 * 按业务域分组，图标来自 lucide-vue-next
 */

import {
  LayoutDashboard,
  Users,
  Gift,
  Package,
  ShoppingCart,
  CreditCard,
  Ticket,
  Tag,
  Server,
  Layers,
  DollarSign,
  Network,
  Activity,
  Radio,
  Shield,
  ClipboardList,
  Settings,
  Megaphone,
} from 'lucide-vue-next'
import type { Component } from 'vue'

export interface NavItem {
  key: string
  labelKey: string
  path: string
  icon: Component
}

export interface NavGroup {
  key: string
  labelKey: string
  items: NavItem[]
}

export const navGroups: NavGroup[] = [
  {
    key: 'cockpit',
    labelKey: 'nav.quench.group.cockpit',
    items: [
      {
        key: 'dashboard',
        labelKey: 'nav.quench.dashboard',
        path: '/admin/dashboard',
        icon: LayoutDashboard,
      },
    ],
  },
  {
    key: 'customers',
    labelKey: 'nav.quench.group.customers',
    items: [
      {
        key: 'users',
        labelKey: 'nav.quench.users',
        path: '/admin/users',
        icon: Users,
      },
      {
        key: 'affiliates',
        labelKey: 'nav.quench.affiliates',
        path: '/admin/affiliates/invites',
        icon: Gift,
      },
    ],
  },
  {
    key: 'monetization',
    labelKey: 'nav.quench.group.monetization',
    items: [
      {
        key: 'subscriptions',
        labelKey: 'nav.quench.subscriptions',
        path: '/admin/subscriptions',
        icon: Package,
      },
      {
        key: 'orders',
        labelKey: 'nav.quench.orders',
        path: '/admin/orders',
        icon: ShoppingCart,
      },
      {
        key: 'paymentDashboard',
        labelKey: 'nav.quench.paymentDashboard',
        path: '/admin/orders/dashboard',
        icon: CreditCard,
      },
      {
        key: 'redeem',
        labelKey: 'nav.quench.redeem',
        path: '/admin/redeem',
        icon: Ticket,
      },
      {
        key: 'promoCodes',
        labelKey: 'nav.quench.promoCodes',
        path: '/admin/promo-codes',
        icon: Tag,
      },
    ],
  },
  {
    key: 'supply',
    labelKey: 'nav.quench.group.supply',
    items: [
      {
        key: 'accounts',
        labelKey: 'nav.quench.accounts',
        path: '/admin/accounts',
        icon: Server,
      },
      {
        key: 'groups',
        labelKey: 'nav.quench.groups',
        path: '/admin/groups',
        icon: Layers,
      },
      {
        key: 'channelPricing',
        labelKey: 'nav.quench.channelPricing',
        path: '/admin/channels/pricing',
        icon: DollarSign,
      },
      {
        key: 'proxies',
        labelKey: 'nav.quench.proxies',
        path: '/admin/proxies',
        icon: Network,
      },
    ],
  },
  {
    key: 'reliability',
    labelKey: 'nav.quench.group.reliability',
    items: [
      {
        key: 'ops',
        labelKey: 'nav.quench.ops',
        path: '/admin/ops',
        icon: Activity,
      },
      {
        key: 'channelMonitor',
        labelKey: 'nav.quench.channelMonitor',
        path: '/admin/channels/monitor',
        icon: Radio,
      },
      {
        key: 'riskControl',
        labelKey: 'nav.quench.riskControl',
        path: '/admin/risk-control',
        icon: Shield,
      },
      {
        key: 'usage',
        labelKey: 'nav.quench.usage',
        path: '/admin/usage',
        icon: ClipboardList,
      },
    ],
  },
  {
    key: 'platform',
    labelKey: 'nav.quench.group.platform',
    items: [
      {
        key: 'settings',
        labelKey: 'nav.quench.settings',
        path: '/admin/settings',
        icon: Settings,
      },
      {
        key: 'announcements',
        labelKey: 'nav.quench.announcements',
        path: '/admin/announcements',
        icon: Megaphone,
      },
    ],
  },
]

/**
 * 从当前 route path 反查所在的 group + item
 */
export function resolveNavItem(path: string): { group: NavGroup; item: NavItem } | null {
  for (const group of navGroups) {
    for (const item of group.items) {
      if (path === item.path || path.startsWith(item.path + '/')) {
        return { group, item }
      }
    }
  }
  return null
}

/**
 * 展平所有导航项，供 ⌘K 搜索使用
 */
export function flatNavItems(): (NavItem & { groupKey: string; groupLabelKey: string })[] {
  return navGroups.flatMap((g) =>
    g.items.map((item) => ({
      ...item,
      groupKey: g.key,
      groupLabelKey: g.labelKey,
    }))
  )
}
