<template>
  <Teleport to="body">
    <Transition name="ufd-slide">
      <div v-if="open" class="ufd-overlay" @click.self="$emit('close')" role="dialog" :aria-label="isEdit ? '编辑用户' : '新建用户'">
        <div class="ufd-panel">
          <!-- 头部 -->
          <div class="ufd-head">
            <div class="ufd-title">{{ isEdit ? '编辑用户' : '新建用户' }}</div>
            <button class="ufd-close" aria-label="关闭" @click="$emit('close')">
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true">
                <path d="M2 2L12 12M12 2L2 12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
            </button>
          </div>

          <!-- 表单主体 -->
          <form class="ufd-body" @submit.prevent="handleSubmit">
            <!-- 邮箱 -->
            <div class="ufd-field">
              <label class="ufd-label">邮箱 <span class="ufd-req">*</span></label>
              <input
                v-model="form.email"
                type="email"
                required
                class="ufd-input"
                placeholder="user@example.com"
                autocomplete="off"
              />
            </div>

            <!-- 密码 -->
            <div class="ufd-field">
              <label class="ufd-label">{{ isEdit ? '新密码（留空不修改）' : '密码 *' }}</label>
              <div class="ufd-row">
                <input
                  v-model="form.password"
                  type="text"
                  class="ufd-input"
                  :required="!isEdit"
                  placeholder="至少 8 位"
                  autocomplete="new-password"
                />
                <button type="button" class="ufd-gen-btn" title="随机生成" @click="generatePassword">
                  <svg width="13" height="13" viewBox="0 0 13 13" fill="none" aria-hidden="true">
                    <path d="M11.5 6.5A5 5 0 1 1 6.5 1.5" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/>
                    <path d="M9 1.5v3h-3" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                </button>
              </div>
            </div>

            <!-- 用户名 -->
            <div class="ufd-field">
              <label class="ufd-label">用户名</label>
              <input v-model="form.username" type="text" class="ufd-input" placeholder="可选" autocomplete="off" />
            </div>

            <!-- 初始余额（仅创建时） -->
            <div v-if="!isEdit" class="ufd-field">
              <label class="ufd-label">初始余额</label>
              <input v-model="form.balance" type="number" step="0.01" min="0" class="ufd-input" placeholder="0.00" />
            </div>

            <!-- 并发上限 -->
            <div class="ufd-field">
              <label class="ufd-label">并发上限</label>
              <input v-model.number="form.concurrency" type="number" min="1" class="ufd-input" />
            </div>

            <!-- RPM 限速 -->
            <div class="ufd-field">
              <label class="ufd-label">RPM 限速 <span class="ufd-hint-inline">（0 = 不限制）</span></label>
              <input v-model.number="form.rpm_limit" type="number" min="0" step="1" class="ufd-input" />
            </div>

            <!-- 备注（仅编辑时） -->
            <div v-if="isEdit" class="ufd-field">
              <label class="ufd-label">备注</label>
              <textarea v-model="form.notes" rows="3" class="ufd-input ufd-textarea" placeholder="管理员内部备注"></textarea>
            </div>

            <!-- 错误提示 -->
            <div v-if="errorMsg" class="ufd-error">{{ errorMsg }}</div>

            <!-- 底部操作 -->
            <div class="ufd-footer">
              <button type="button" class="ufd-btn" @click="$emit('close')">取消</button>
              <button type="submit" class="ufd-btn ufd-btn-primary" :disabled="submitting">
                {{ submitting ? (isEdit ? '保存中…' : '创建中…') : (isEdit ? '保存更改' : '创建用户') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { adminAPI } from '@/api/admin'
import type { AdminUser } from '@/types'
import { useAppStore } from '@/stores/app'

const props = defineProps<{
  open: boolean
  user: AdminUser | null
}>()

const emit = defineEmits<{
  close: []
  success: []
}>()

const appStore = useAppStore()
const submitting = ref(false)
const errorMsg = ref('')
const isEdit = computed(() => !!props.user)

const form = reactive({
  email: '',
  password: '',
  username: '',
  balance: '',
  concurrency: 1,
  rpm_limit: 0,
  notes: '',
})

// 同步 user prop → form
watch(() => props.user, (u) => {
  if (u) {
    form.email = u.email
    form.password = ''
    form.username = u.username || ''
    form.concurrency = u.concurrency
    form.rpm_limit = u.rpm_limit ?? 0
    form.notes = u.notes || ''
    form.balance = ''
  } else {
    form.email = ''
    form.password = ''
    form.username = ''
    form.concurrency = 1
    form.rpm_limit = 0
    form.notes = ''
    form.balance = ''
  }
  errorMsg.value = ''
}, { immediate: true })

// 每次面板打开时清错误
watch(() => props.open, (v) => {
  if (v) errorMsg.value = ''
})

function generatePassword() {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789!@#$%^&*'
  let p = ''
  for (let i = 0; i < 16; i++) p += chars.charAt(Math.floor(Math.random() * chars.length))
  form.password = p
}

async function handleSubmit() {
  errorMsg.value = ''
  if (!form.email.trim()) { errorMsg.value = '邮箱不能为空'; return }
  if (!isEdit.value && !form.password.trim()) { errorMsg.value = '密码不能为空'; return }
  if (form.concurrency < 1) { errorMsg.value = '并发上限至少为 1'; return }

  submitting.value = true
  try {
    if (isEdit.value && props.user) {
      const data: Record<string, unknown> = {
        email: form.email,
        username: form.username,
        notes: form.notes,
        concurrency: form.concurrency,
        rpm_limit: form.rpm_limit,
      }
      if (form.password.trim()) data.password = form.password.trim()
      await adminAPI.users.update(props.user.id, data as any)
      appStore.showSuccess('用户已更新')
    } else {
      const balanceStr = String(form.balance).trim()
      const payload: Parameters<typeof adminAPI.users.create>[0] = {
        email: form.email,
        password: form.password,
        username: form.username || undefined,
        concurrency: form.concurrency,
        rpm_limit: form.rpm_limit,
      }
      if (balanceStr !== '') payload.balance = Number(balanceStr)
      await adminAPI.users.create(payload)
      appStore.showSuccess('用户已创建')
    }
    emit('success')
  } catch (e: any) {
    errorMsg.value = e?.response?.data?.detail || e?.message || '操作失败'
  } finally {
    submitting.value = false
  }
}
</script>

<style src="./user-form-drawer.css"></style>
