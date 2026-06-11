/**
 * defineResource — 带类型推导的资源描述符工厂函数
 * 淬钢 QUENCH · 资源描述符系统
 */

import type { ResourceDef } from './types'

/**
 * 带完整类型推导的恒等函数
 * 用法：export const myResource = defineResource<MyType>({ ... })
 */
export function defineResource<
  T extends Record<string, unknown>,
  CreateDto = Partial<T>,
  UpdateDto = Partial<T>
>(def: ResourceDef<T, CreateDto, UpdateDto>): ResourceDef<T, CreateDto, UpdateDto> {
  return def
}
