import type { Component } from 'vue'

export type FieldType =
  | 'text'
  | 'number'
  | 'switch'
  | 'select'
  | 'textarea'
  | 'password'
  | 'image'
  | 'json'

export interface SelectOption {
  value: string | number | boolean
  label: string
}

export interface SettingsField {
  /** Flat key in SystemSettings / UpdateSettingsRequest */
  key: string
  label: string
  type: FieldType
  options?: SelectOption[]
  placeholder?: string
  help?: string
  /** Show this field only when condition holds (receives the current form values) */
  showWhen?: (values: Record<string, unknown>) => boolean
  /** Mask display value (password-style) */
  sensitive?: boolean
}

export type TabId =
  | 'general'
  | 'agreement'
  | 'features'
  | 'security'
  | 'users'
  | 'gateway'
  | 'payment'
  | 'email'
  | 'backup'

export interface SettingsSection {
  id: string
  tab: TabId
  title: string
  description?: string
  fields: SettingsField[]
  /** Custom component escape hatch — overrides field rendering for this section */
  component?: Component
}
