<template>
  <AppLayout>
    <ResourcePage ref="pageRef" :resource="enrichedResource">
      <!-- 协议 badge -->
      <template #cell-protocol="{ value }">
        <span
          v-if="value"
          :class="['rp2-badge', String(value).startsWith('socks5') ? 'rp2-badge-azure' : 'rp2-badge-gray']"
        >
          {{ String(value).toUpperCase() }}
        </span>
        <span v-else class="rp2-muted">-</span>
      </template>

      <!-- 主机:端口 mono -->
      <template #cell-host_port="{ row }">
        <code class="rp2-mono">{{ row.host }}:{{ row.port }}</code>
      </template>

      <!-- 状态点 + 文字 -->
      <template #cell-status="{ value }">
        <span class="rp2-status">
          <span
            class="rp2-dot"
            :class="{
              'rp2-dot-green': value === 'active',
              'rp2-dot-red': value === 'inactive' || value === 'expired'
            }"
          ></span>
          <span class="rp2-status-text">
            {{ value === 'active' ? '活跃' : value === 'inactive' ? '禁用' : '已过期' }}
          </span>
        </span>
      </template>

      <!-- 账号数 -->
      <template #cell-account_count="{ value }">
        <span class="rp2-mono rp2-muted">{{ value ?? 0 }}</span>
      </template>

      <!-- 创建时间 -->
      <template #cell-created_at="{ value }">
        <span class="rp2-muted rp2-xs">{{ formatDate(value as string) }}</span>
      </template>
    </ResourcePage>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { ResourcePage } from '@/resource'
import { proxiesResource } from '@/resource/resources/proxies'
import type { ResourceDef, RowAction } from '@/resource/types'
import type { Proxy } from '@/types'
import { adminAPI } from '@/api/admin'

const pageRef = ref<InstanceType<typeof ResourcePage> | null>(null)

function formatDate(iso: string) {
  if (!iso) return '-'
  return new Date(iso).toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit'
  })
}

// 注入行操作（需要访问 pageRef 打开抽屉）
const enrichedResource = computed<ResourceDef>(() => ({
  ...(proxiesResource as unknown as ResourceDef),
  rowActions: [
    {
      key: 'edit',
      label: '编辑',
      handler(row) {
        // 注入 _isEdit 标记，让 status 字段 showWhen 可感知
        pageRef.value?.openEditDrawer({ ...row, _isEdit: true })
      }
    },
    {
      key: 'delete',
      label: '删除',
      danger: true,
      async handler(row) {
        const proxy = row as unknown as Proxy
        if (!confirm(`确认删除代理「${proxy.name}」？此操作不可撤销。`)) return
        await adminAPI.proxies.delete(proxy.id)
        pageRef.value?.reload()
      }
    }
  ] as RowAction<Record<string, unknown>>[]
}))
</script>

<style scoped>
/* 协议 badge */
.rp2-badge {
  display: inline-block;
  padding: 2px 7px;
  border-radius: 5px;
  font-size: 10.5px;
  font-weight: 700;
  letter-spacing: 0.06em;
  font-family: var(--font-mono, "IBM Plex Mono", monospace);
}

.rp2-badge-azure {
  background: var(--azure-dim, rgba(92, 168, 255, 0.14));
  color: var(--azure-hi, #8CC4FF);
  border: 1px solid rgba(92, 168, 255, 0.3);
}

.rp2-badge-gray {
  background: var(--bg-3, #1F232B);
  color: var(--ink-1, #97A0AF);
  border: 1px solid var(--line-1, #2F3540);
}

/* 状态指示器 */
.rp2-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.rp2-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.rp2-dot-green {
  background: #3DD68C;
  box-shadow: 0 0 6px rgba(61, 214, 140, 0.5);
}

.rp2-dot-red {
  background: var(--bad, #F25C69);
}

.rp2-status-text {
  font-size: 12px;
  color: var(--ink-0, #E8EBF0);
}

/* mono code */
.rp2-mono {
  font-family: var(--font-mono, "IBM Plex Mono", monospace);
  font-size: 11.5px;
  color: var(--ink-0, #E8EBF0);
}

.rp2-muted {
  color: var(--ink-2, #5C6470);
}

.rp2-xs {
  font-size: 11.5px;
}
</style>
