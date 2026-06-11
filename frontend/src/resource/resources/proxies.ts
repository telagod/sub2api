/**
 * 代理管理 资源描述符
 * 淬钢 QUENCH · proxies resource
 */

import { defineResource } from '../defineResource'
import { adminAPI } from '@/api/admin'
import type { Proxy } from '@/types'

// Proxy 接口无索引签名，用交叉类型让其满足 Record<string, unknown> 约束
type ProxyRow = Proxy & Record<string, unknown>

export const proxiesResource = defineResource<ProxyRow>({
  key: 'proxies-v2',
  title: '代理管理',
  rowKey: 'id',

  // ── API 适配器 ──────────────────────────────────────────────────────
  api: {
    async list({ page, pageSize, sort, order, q, filters }) {
      const result = await adminAPI.proxies.list(page, pageSize, {
        protocol: (filters?.protocol as string) || undefined,
        status: (filters?.status as 'active' | 'inactive' | 'expired') || undefined,
        search: q || undefined,
        sort_by: sort || undefined,
        sort_order: order
      })
      return {
        items: result.items as unknown as ProxyRow[],
        total: result.total
      }
    },

    async create(dto) {
      const port = typeof dto.port === 'string' ? parseInt(dto.port as string, 10) : (dto.port as number)
      const result = await adminAPI.proxies.create({
        name: (dto.name as string).trim(),
        protocol: dto.protocol as Proxy['protocol'],
        host: (dto.host as string).trim(),
        port,
        username: (dto.username as string)?.trim() || null,
        password: (dto.password as string)?.trim() || null,
        expires_at: null,
        fallback_mode: (dto.fallback_mode as 'none' | 'proxy' | 'direct') ?? 'none',
      })
      return result as unknown as ProxyRow
    },

    async update(id, dto) {
      const port = dto.port != null
        ? (typeof dto.port === 'string' ? parseInt(dto.port as string, 10) : (dto.port as number))
        : undefined
      const result = await adminAPI.proxies.update(id as number, {
        name: (dto.name as string | undefined)?.trim(),
        protocol: dto.protocol as Proxy['protocol'] | undefined,
        host: (dto.host as string | undefined)?.trim(),
        port,
        username: (dto.username as string | undefined)?.trim() ?? null,
        ...(dto.password !== undefined ? { password: (dto.password as string).trim() || null } : {}),
        status: dto.status as 'active' | 'inactive' | undefined,
        fallback_mode: dto.fallback_mode as 'none' | 'proxy' | 'direct' | undefined,
      })
      return result as unknown as ProxyRow
    },

    async remove(id) {
      await adminAPI.proxies.delete(id as number)
    },

    async removeMany(ids) {
      await adminAPI.proxies.batchDelete(ids as number[])
    }
  },

  // ── 列定义 ──────────────────────────────────────────────────────────
  columns: [
    { key: 'name', title: '名称', sortable: true },
    { key: 'protocol', title: '协议', width: '90px' },
    { key: 'host_port', title: '主机:端口' },
    { key: 'status', title: '状态', width: '80px' },
    { key: 'account_count', title: '账号数', width: '80px', align: 'right', sortable: true },
    { key: 'created_at', title: '创建时间', width: '140px', sortable: true },
    { key: 'actions', title: '操作', width: '120px', align: 'right' }
  ],

  // ── 过滤器 ──────────────────────────────────────────────────────────
  filters: [
    {
      key: 'protocol',
      label: '协议',
      type: 'select',
      placeholder: '全部协议',
      options: [
        { label: 'HTTP', value: 'http' },
        { label: 'HTTPS', value: 'https' },
        { label: 'SOCKS5', value: 'socks5' },
        { label: 'SOCKS5H', value: 'socks5h' }
      ]
    },
    {
      key: 'status',
      label: '状态',
      type: 'select',
      placeholder: '全部状态',
      options: [
        { label: '活跃', value: 'active' },
        { label: '禁用', value: 'inactive' },
        { label: '已过期', value: 'expired' }
      ]
    }
  ],

  // ── 表单字段 ────────────────────────────────────────────────────────
  form: [
    { key: 'name', label: '名称', type: 'text', required: true, placeholder: '代理名称' },
    {
      key: 'protocol',
      label: '协议',
      type: 'select',
      required: true,
      options: [
        { label: 'HTTP', value: 'http' },
        { label: 'HTTPS', value: 'https' },
        { label: 'SOCKS5', value: 'socks5' },
        { label: 'SOCKS5H', value: 'socks5h' }
      ]
    },
    { key: 'host', label: '主机', type: 'text', required: true, placeholder: '127.0.0.1 或域名' },
    { key: 'port', label: '端口', type: 'number', required: true, placeholder: '8080' },
    { key: 'username', label: '用户名（可选）', type: 'text', placeholder: '留空表示无认证' },
    { key: 'password', label: '密码（可选）', type: 'password', placeholder: '留空表示无认证' },
    {
      key: 'status',
      label: '状态',
      type: 'select',
      options: [
        { label: '活跃', value: 'active' },
        { label: '禁用', value: 'inactive' }
      ],
      // 仅编辑时显示（新建时无意义）
      showWhen: (data) => !!(data._isEdit)
    },
    {
      key: 'fallback_mode',
      label: '回退模式',
      type: 'select',
      options: [
        { label: '无回退', value: 'none' },
        { label: '直连回退', value: 'direct' },
        { label: '代理回退', value: 'proxy' }
      ]
    }
  ],

  // ── 批量操作 ─────────────────────────────────────────────────────────
  bulkActions: [
    {
      key: 'bulk-delete',
      label: '批量删除',
      danger: true,
      async handler(rows) {
        if (!confirm(`确认删除选中的 ${rows.length} 个代理？此操作不可撤销。`)) return
        const ids = rows.map((r) => r.id as number)
        await adminAPI.proxies.batchDelete(ids)
      }
    }
  ]
})
