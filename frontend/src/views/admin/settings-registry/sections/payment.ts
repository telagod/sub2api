/**
 * Payment tab — payment system configuration.
 *
 * Keys sourced from SettingsView.vue activeTab === 'payment' region (lines 5707–6196).
 *
 * form. bindings in old view: 20
 *   payment_enabled, payment_product_name_prefix, payment_product_name_suffix,
 *   payment_min_amount, payment_max_amount, payment_daily_limit,
 *   payment_balance_recharge_multiplier, payment_recharge_fee_rate,
 *   payment_order_timeout_minutes, payment_max_pending_orders,
 *   payment_load_balance_strategy, payment_cancel_rate_limit_enabled,
 *   payment_cancel_rate_limit_window_mode, payment_cancel_rate_limit_window,
 *   payment_cancel_rate_limit_unit, payment_cancel_rate_limit_max,
 *   payment_alipay_force_qrcode, payment_enabled_types,
 *   payment_help_image_url, payment_help_text
 *
 * Flat schema fields: 18 (all except payment_enabled_types + provider list)
 * Escape hatch section: PaymentProviderListSection (payment_enabled_types + provider CRUD)
 * Total: 20 / 20 ✓
 */
import { defineAsyncComponent } from 'vue'
import type { SettingsSection } from '../types'

const paymentConfig: SettingsSection = {
  id: 'payment.config',
  tab: 'payment',
  title: 'admin.settings.payment.title',
  description: 'admin.settings.payment.description',
  fields: [
    {
      key: 'payment_enabled',
      label: 'admin.settings.payment.enabled',
      type: 'switch',
      help: 'admin.settings.payment.enabledHint',
    },
    {
      key: 'payment_product_name_prefix',
      label: 'admin.settings.payment.productNamePrefix',
      type: 'text',
      placeholder: 'subme',
      showWhen: (v) => !!v['payment_enabled'],
    },
    {
      key: 'payment_product_name_suffix',
      label: 'admin.settings.payment.productNameSuffix',
      type: 'text',
      placeholder: 'CNY',
      showWhen: (v) => !!v['payment_enabled'],
    },
    {
      key: 'payment_min_amount',
      label: 'admin.settings.payment.minAmount',
      type: 'number',
      placeholder: 'admin.settings.payment.noLimit',
      showWhen: (v) => !!v['payment_enabled'],
    },
    {
      key: 'payment_max_amount',
      label: 'admin.settings.payment.maxAmount',
      type: 'number',
      placeholder: 'admin.settings.payment.noLimit',
      showWhen: (v) => !!v['payment_enabled'],
    },
    {
      key: 'payment_daily_limit',
      label: 'admin.settings.payment.dailyLimit',
      type: 'number',
      placeholder: 'admin.settings.payment.noLimit',
      showWhen: (v) => !!v['payment_enabled'],
    },
    {
      key: 'payment_balance_recharge_multiplier',
      label: 'admin.settings.payment.balanceRechargeMultiplier',
      type: 'number',
      help: 'admin.settings.payment.balanceRechargeMultiplierHint',
      showWhen: (v) => !!v['payment_enabled'],
    },
    {
      key: 'payment_recharge_fee_rate',
      label: 'admin.settings.payment.rechargeFeeRate',
      type: 'number',
      help: 'admin.settings.payment.rechargeFeeRateHint',
      showWhen: (v) => !!v['payment_enabled'],
    },
    {
      key: 'payment_order_timeout_minutes',
      label: 'admin.settings.payment.orderTimeout',
      type: 'number',
      help: 'admin.settings.payment.orderTimeoutHint',
      showWhen: (v) => !!v['payment_enabled'],
    },
    {
      key: 'payment_max_pending_orders',
      label: 'admin.settings.payment.maxPendingOrders',
      type: 'number',
      showWhen: (v) => !!v['payment_enabled'],
    },
    {
      key: 'payment_load_balance_strategy',
      label: 'admin.settings.payment.loadBalanceStrategy',
      type: 'select',
      options: [
        { value: 'random', label: 'Random' },
        { value: 'round_robin', label: 'Round Robin' },
        { value: 'least_conn', label: 'Least Connections' },
      ],
      showWhen: (v) => !!v['payment_enabled'],
    },
    {
      key: 'payment_cancel_rate_limit_enabled',
      label: 'admin.settings.payment.cancelRateLimit',
      type: 'switch',
      showWhen: (v) => !!v['payment_enabled'],
    },
    {
      key: 'payment_cancel_rate_limit_window_mode',
      label: '取消限流窗口模式',
      type: 'select',
      options: [
        { value: 'sliding', label: 'Sliding' },
        { value: 'fixed', label: 'Fixed' },
      ],
      showWhen: (v) => !!v['payment_enabled'] && !!v['payment_cancel_rate_limit_enabled'],
    },
    {
      key: 'payment_cancel_rate_limit_window',
      label: '取消限流窗口大小',
      type: 'number',
      showWhen: (v) => !!v['payment_enabled'] && !!v['payment_cancel_rate_limit_enabled'],
    },
    {
      key: 'payment_cancel_rate_limit_unit',
      label: '取消限流时间单位',
      type: 'select',
      options: [
        { value: 'minute', label: '分钟' },
        { value: 'hour', label: '小时' },
        { value: 'day', label: '天' },
      ],
      showWhen: (v) => !!v['payment_enabled'] && !!v['payment_cancel_rate_limit_enabled'],
    },
    {
      key: 'payment_cancel_rate_limit_max',
      label: '取消限流最大次数',
      type: 'number',
      showWhen: (v) => !!v['payment_enabled'] && !!v['payment_cancel_rate_limit_enabled'],
    },
    {
      key: 'payment_alipay_force_qrcode',
      label: 'admin.settings.payment.alipayForceQRCode',
      type: 'switch',
      help: 'admin.settings.payment.alipayForceQRCodeHint',
      showWhen: (v) => !!v['payment_enabled'],
    },
    {
      key: 'payment_help_image_url',
      label: 'admin.settings.payment.helpImage',
      type: 'image',
      showWhen: (v) => !!v['payment_enabled'],
    },
    {
      key: 'payment_help_text',
      label: 'admin.settings.payment.helpText',
      type: 'textarea',
      placeholder: 'admin.settings.payment.helpTextPlaceholder',
      showWhen: (v) => !!v['payment_enabled'],
    },
  ],
}

/**
 * Payment provider list — dynamic CRUD + payment_enabled_types badge toggler.
 * Delegates to PaymentProviderListSection custom component.
 * Covers: payment_enabled_types + provider CRUD (not a flat key/value field).
 */
const paymentProviders: SettingsSection = {
  id: 'payment.providers',
  tab: 'payment',
  title: 'admin.settings.payment.enabledPaymentTypes',
  description: 'admin.settings.payment.enabledPaymentTypesHint',
  fields: [],
  component: defineAsyncComponent(
    () => import('../special/PaymentProviderListSection.vue'),
  ),
}

export default [paymentConfig, paymentProviders] as SettingsSection[]
