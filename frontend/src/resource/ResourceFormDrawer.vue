<template>
  <!-- 遮罩 -->
  <Transition name="q-drawer-mask">
    <div
      v-if="modelValue"
      class="q-drawer-mask"
      aria-hidden="true"
      @click="handleClose"
    />
  </Transition>

  <!-- 抽屉面板 -->
  <Transition name="q-drawer">
    <aside
      v-if="modelValue"
      class="q-drawer"
      role="dialog"
      :aria-label="title"
      aria-modal="true"
    >
      <!-- 头部 -->
      <div class="q-drawer-header">
        <span class="q-drawer-title">{{ title }}</span>
        <button class="q-drawer-close" :aria-label="'关闭'" @click="handleClose">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true">
            <path d="M2 2L12 12M12 2L2 12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
          </svg>
        </button>
      </div>

      <!-- 表单区 -->
      <div class="q-drawer-body">
        <form id="resource-form" class="q-drawer-form" @submit.prevent="handleSubmit">
          <template v-for="field in visibleFields" :key="field.key">
            <div class="q-field">
              <label class="q-field-label">
                {{ field.label }}
                <span v-if="field.required" class="q-field-req" aria-hidden="true">*</span>
              </label>

              <!-- text / password -->
              <template v-if="field.type === 'text' || field.type === 'password'">
                <div class="q-input-wrap">
                  <input
                    v-model="formData[field.key]"
                    :type="passwordVisible[field.key] ? 'text' : field.type"
                    :required="field.required"
                    :placeholder="field.placeholder ?? ''"
                    class="q-input"
                  />
                  <button
                    v-if="field.type === 'password'"
                    type="button"
                    class="q-pw-toggle"
                    :aria-label="passwordVisible[field.key] ? '隐藏' : '显示'"
                    @click="passwordVisible[field.key] = !passwordVisible[field.key]"
                  >
                    <svg v-if="passwordVisible[field.key]" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
                      <path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19m-6.72-1.07a3 3 0 11-4.24-4.24"/>
                      <line x1="1" y1="1" x2="23" y2="23"/>
                    </svg>
                    <svg v-else width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
                      <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                      <circle cx="12" cy="12" r="3"/>
                    </svg>
                  </button>
                </div>
              </template>

              <!-- number -->
              <template v-else-if="field.type === 'number'">
                <input
                  v-model.number="formData[field.key]"
                  type="number"
                  :required="field.required"
                  :placeholder="field.placeholder ?? ''"
                  class="q-input"
                />
              </template>

              <!-- select -->
              <template v-else-if="field.type === 'select'">
                <select v-model="formData[field.key]" :required="field.required" class="q-select">
                  <option v-if="field.placeholder" value="" disabled>{{ field.placeholder }}</option>
                  <option
                    v-for="opt in (field.options ?? [])"
                    :key="String(opt.value)"
                    :value="opt.value"
                  >{{ opt.label }}</option>
                </select>
              </template>

              <!-- switch -->
              <template v-else-if="field.type === 'switch'">
                <label class="q-switch">
                  <input
                    v-model="formData[field.key]"
                    type="checkbox"
                    class="q-switch-input"
                  />
                  <span class="q-switch-track"></span>
                </label>
              </template>
            </div>
          </template>
        </form>
      </div>

      <!-- 底部操作 -->
      <div class="q-drawer-footer">
        <button type="button" class="q-btn q-btn-ghost" @click="handleClose">取消</button>
        <button
          type="submit"
          form="resource-form"
          class="q-btn q-btn-primary"
          :disabled="submitting"
        >
          <svg
            v-if="submitting"
            class="q-spin"
            width="14" height="14" viewBox="0 0 24 24" fill="none"
            aria-hidden="true"
          >
            <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" opacity=".25"/>
            <path d="M12 2a10 10 0 010 20" stroke="currentColor" stroke-width="3" stroke-linecap="round"/>
          </svg>
          {{ submitting ? '保存中…' : (isEdit ? '保存' : '创建') }}
        </button>
      </div>
    </aside>
  </Transition>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import type { FieldDef } from './types'

// ── Props ──────────────────────────────────────────────────────────────
const props = defineProps<{
  modelValue: boolean
  title: string
  fields: FieldDef[]
  /** 编辑时传入初始值，新建时为 undefined */
  initialData?: Record<string, unknown>
  submitting?: boolean
}>()

// ── Emits ──────────────────────────────────────────────────────────────
const emit = defineEmits<{
  'update:modelValue': [v: boolean]
  'submit': [data: Record<string, unknown>]
}>()

// ── 状态 ───────────────────────────────────────────────────────────────
const isEdit = computed(() => !!props.initialData)
const formData = reactive<Record<string, unknown>>({})
const passwordVisible = reactive<Record<string, boolean>>({})

// 初始数据同步
watch(
  () => [props.modelValue, props.initialData] as const,
  ([open, init]) => {
    if (!open) return
    // 重置 formData
    for (const key of Object.keys(formData)) {
      delete formData[key]
    }
    for (const key of Object.keys(passwordVisible)) {
      delete passwordVisible[key]
    }
    // 填入初始值或字段默认
    for (const field of props.fields) {
      if (init && field.key in init) {
        formData[field.key] = init[field.key]
      } else {
        formData[field.key] = field.type === 'switch' ? false
          : field.type === 'number' ? 0
          : field.type === 'select' && field.options?.[0] ? field.options[0].value
          : ''
      }
    }
  },
  { immediate: true }
)

const visibleFields = computed(() =>
  props.fields.filter(f => !f.showWhen || f.showWhen(formData))
)

function handleClose() {
  emit('update:modelValue', false)
}

function handleSubmit() {
  emit('submit', { ...formData })
}
</script>

<style scoped>
/* ── 淬钢 QUENCH · ResourceFormDrawer ── */

/* 遮罩 */
.q-drawer-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  z-index: 49;
  backdrop-filter: blur(1px);
}

/* 抽屉面板 */
.q-drawer {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 480px;
  max-width: 100vw;
  z-index: 50;
  display: flex;
  flex-direction: column;
  background: var(--metal-base, linear-gradient(180deg, #1A1E26 0%, #13161C 100%));
  border-left: 1px solid var(--line-1, #2F3540);
  box-shadow:
    inset 1px 0 0 rgba(255, 255, 255, 0.04),
    -20px 0 60px rgba(0, 0, 0, 0.5),
    0 0 40px rgba(92, 168, 255, 0.04);
  font-family: var(--font-ui, "Archivo", "PingFang SC", sans-serif);
  font-size: 13px;
  color: var(--ink-0, #E8EBF0);
}

/* 头部 */
.q-drawer-header {
  display: flex;
  align-items: center;
  padding: 18px 22px;
  border-bottom: 1px solid var(--line-0, #20242C);
  flex-shrink: 0;
}

.q-drawer-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--ink-0, #E8EBF0);
  flex: 1;
  letter-spacing: 0.01em;
}

.q-drawer-close {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  border-radius: 7px;
  color: var(--ink-2, #5C6470);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.q-drawer-close:hover {
  background: var(--bg-2, #171A20);
  color: var(--ink-0, #E8EBF0);
}

/* 表单区 */
.q-drawer-body {
  flex: 1;
  overflow-y: auto;
  padding: 22px;
}

.q-drawer-form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

/* 字段 */
.q-field {
  display: flex;
  flex-direction: column;
  gap: 7px;
}

.q-field-label {
  font-size: 11.5px;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--ink-2, #5C6470);
}

.q-field-req {
  color: var(--bad, #F25C69);
  margin-left: 2px;
}

/* input */
.q-input-wrap {
  position: relative;
}

.q-input {
  width: 100%;
  background: var(--bg-2, #171A20);
  border: 1px solid var(--line-1, #2F3540);
  border-radius: 8px;
  padding: 9px 12px;
  font-size: 13px;
  color: var(--ink-0, #E8EBF0);
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
  box-sizing: border-box;
  font-family: inherit;
}

.q-input-wrap .q-input {
  padding-right: 38px;
}

.q-input:focus {
  border-color: rgba(92, 168, 255, 0.55);
  box-shadow: 0 0 0 2px rgba(92, 168, 255, 0.15);
}

.q-input::placeholder {
  color: var(--ink-2, #5C6470);
}

/* select */
.q-select {
  width: 100%;
  background: var(--bg-2, #171A20);
  border: 1px solid var(--line-1, #2F3540);
  border-radius: 8px;
  padding: 9px 12px;
  font-size: 13px;
  color: var(--ink-0, #E8EBF0);
  outline: none;
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
  font-family: inherit;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' viewBox='0 0 10 10'%3E%3Cpath d='M2 3.5L5 6.5L8 3.5' stroke='%235C6470' stroke-width='1.5' stroke-linecap='round' fill='none'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 32px;
}

.q-select:focus {
  border-color: rgba(92, 168, 255, 0.55);
  box-shadow: 0 0 0 2px rgba(92, 168, 255, 0.15);
}

/* password toggle */
.q-pw-toggle {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: var(--ink-2, #5C6470);
  cursor: pointer;
  display: grid;
  place-items: center;
  padding: 2px;
  border-radius: 4px;
  transition: color 0.15s;
}

.q-pw-toggle:hover {
  color: var(--ink-0, #E8EBF0);
}

/* switch */
.q-switch {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
}

.q-switch-input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.q-switch-track {
  display: inline-block;
  width: 36px;
  height: 20px;
  border-radius: 10px;
  background: var(--bg-3, #1F232B);
  border: 1px solid var(--line-1, #2F3540);
  position: relative;
  transition: background 0.2s, border-color 0.2s;
}

.q-switch-track::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: var(--ink-2, #5C6470);
  transition: transform 0.2s, background 0.2s;
}

.q-switch-input:checked + .q-switch-track {
  background: var(--azure-dim, rgba(92, 168, 255, 0.25));
  border-color: rgba(92, 168, 255, 0.5);
}

.q-switch-input:checked + .q-switch-track::after {
  transform: translateX(16px);
  background: var(--azure, #5CA8FF);
}

/* 底部 */
.q-drawer-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 22px;
  border-top: 1px solid var(--line-0, #20242C);
  flex-shrink: 0;
}

/* 按钮 */
.q-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
  border: 1px solid transparent;
}

.q-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.q-btn-ghost {
  background: transparent;
  border-color: var(--line-1, #2F3540);
  color: var(--ink-1, #97A0AF);
}

.q-btn-ghost:hover:not(:disabled) {
  background: var(--bg-2, #171A20);
  color: var(--ink-0, #E8EBF0);
}

.q-btn-primary {
  background: var(--azure-dim, rgba(92, 168, 255, 0.18));
  border-color: rgba(92, 168, 255, 0.45);
  color: var(--azure-hi, #8CC4FF);
}

.q-btn-primary:hover:not(:disabled) {
  background: rgba(92, 168, 255, 0.28);
  border-color: rgba(92, 168, 255, 0.65);
  box-shadow: 0 0 16px rgba(92, 168, 255, 0.2);
}

/* 旋转动画 */
.q-spin {
  animation: q-spin 0.8s linear infinite;
}

@keyframes q-spin {
  to { transform: rotate(360deg); }
}

/* 入场过渡 */
.q-drawer-enter-active,
.q-drawer-leave-active {
  transition: transform 0.28s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.q-drawer-enter-from,
.q-drawer-leave-to {
  transform: translateX(100%);
}

.q-drawer-mask-enter-active,
.q-drawer-mask-leave-active {
  transition: opacity 0.25s;
}

.q-drawer-mask-enter-from,
.q-drawer-mask-leave-to {
  opacity: 0;
}

@media (prefers-reduced-motion: reduce) {
  .q-drawer-enter-active,
  .q-drawer-leave-active,
  .q-drawer-mask-enter-active,
  .q-drawer-mask-leave-active { transition: none; }
}
</style>
