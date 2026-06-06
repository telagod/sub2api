import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

/**
 * 合并 Tailwind class：clsx 处理条件/数组，twMerge 消解冲突类（后者胜出）。
 * shadcn-vue 组件的标准 class 合并工具。
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
