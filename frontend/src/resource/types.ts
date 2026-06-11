/**
 * Resource Descriptor 类型定义
 * 淬钢 QUENCH · 资源描述符系统
 */

import type { ColumnDef } from '@/components/datatable/types'

/** 下拉选项 */
export interface SelectOption {
  label: string
  value: string | number | boolean
}

/** 过滤字段定义 */
export interface FilterDef {
  key: string
  label: string
  type: 'select' | 'text'
  options?: SelectOption[]
  placeholder?: string
}

/** 表单字段定义 */
export interface FieldDef {
  key: string
  label: string
  type: 'text' | 'number' | 'select' | 'password' | 'switch'
  required?: boolean
  options?: SelectOption[]
  placeholder?: string
  /** 字段级别的条件显示，返回 true 才渲染 */
  showWhen?: (formData: Record<string, unknown>) => boolean
}

/** 行操作 */
export interface RowAction<T = Record<string, unknown>> {
  key: string
  label: string
  /** danger 样式 */
  danger?: boolean
  /** 是否对该行隐藏 */
  hidden?: (row: T) => boolean
  handler: (row: T) => void | Promise<void>
}

/** 批量操作 */
export interface BulkAction<T = Record<string, unknown>> {
  key: string
  label: string
  danger?: boolean
  handler: (rows: T[]) => void | Promise<void>
}

/** API 适配器 */
export interface ResourceApi<T = Record<string, unknown>, CreateDto = Partial<T>, UpdateDto = Partial<T>> {
  list: (params: {
    page: number
    pageSize: number
    sort?: string
    order?: 'asc' | 'desc'
    q?: string
    filters?: Record<string, string | string[]>
  }) => Promise<{ items: T[]; total: number }>
  create?: (dto: CreateDto) => Promise<T>
  update?: (id: number | string, dto: UpdateDto) => Promise<T>
  remove?: (id: number | string) => Promise<void>
  removeMany?: (ids: (number | string)[]) => Promise<void>
}

/** 资源描述符 */
export interface ResourceDef<
  T extends Record<string, unknown> = Record<string, unknown>,
  CreateDto = Partial<T>,
  UpdateDto = Partial<T>
> {
  /** 唯一标识（用于 localStorage key 等隔离） */
  key: string
  /** 页面标题 */
  title: string
  /** 行的主键字段名，默认 'id' */
  rowKey?: keyof T & string
  /** API 适配器 */
  api: ResourceApi<T, CreateDto, UpdateDto>
  /** 列定义 */
  columns: ColumnDef<T>[]
  /** 过滤字段 */
  filters?: FilterDef[]
  /** 表单字段（用于新建/编辑抽屉） */
  form?: FieldDef[]
  /** 行操作（每行末尾的操作按钮） */
  rowActions?: RowAction<T>[]
  /** 批量操作（BulkBar 内的按钮） */
  bulkActions?: BulkAction<T>[]
}
