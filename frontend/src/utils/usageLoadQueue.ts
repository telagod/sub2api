/**
 * Usage request scheduler.
 *
 * 后端使用 passive sampling，上游 429 已不是顾虑；但浏览器侧仍有连接上限
 * （HTTP/1.1 单域 ~6 连接，HTTP/2 受 max_concurrent_streams 限制）。
 * 大数据量列表渲染时，每行 cell 各自发起 getUsage，瞬时并发会淹没连接池、
 * 让 getBatchTodayStats 等关键请求排在队尾、整页 loading 被拖长。
 *
 * 因此这里用一个轻量信号量限制并发：未占满名额立即执行，占满则排队，
 * 前一个完成后按 FIFO 唤醒下一个。仅削平瞬时尖峰，不影响最终全部完成。
 */

import type { Account } from '@/types'

/** 最大并发 usage 请求数。留出余量给关键请求，避免占满浏览器连接池。 */
const MAX_CONCURRENT = 6

let active = 0
const waiters: Array<() => void> = []

function acquire(): Promise<void> {
  if (active < MAX_CONCURRENT) {
    active++
    return Promise.resolve()
  }
  return new Promise<void>((resolve) => {
    waiters.push(resolve)
  })
}

function release(): void {
  const next = waiters.shift()
  if (next) {
    // 名额直接转交给下一个等待者，active 保持不变
    next()
  } else {
    active--
  }
}

/**
 * Schedule a usage fetch. Runs immediately if a concurrency slot is free,
 * otherwise queues until one frees up.
 */
export function enqueueUsageRequest<T>(
  _account: Account,
  fn: () => Promise<T>
): Promise<T> {
  return acquire().then(() => fn().finally(release))
}
