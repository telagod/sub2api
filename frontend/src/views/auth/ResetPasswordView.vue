<template>
  <AuthLayout>
    <div class="rpv-body">
      <!-- 标题 -->
      <div class="rpv-head">
        <h2 class="rpv-title">{{ t('auth.resetPasswordTitle') }}</h2>
        <p class="rpv-sub">{{ t('auth.resetPasswordHint') }}</p>
      </div>

      <!-- 无效链接 -->
      <div v-if="isInvalidLink" class="rpv-state rpv-state--warn">
        <div class="rpv-state-icon-wrap rpv-state-icon-wrap--warn">
          <Icon name="exclamationCircle" size="lg" class="rpv-icon-warn" />
        </div>
        <div>
          <h3 class="rpv-state-title">{{ t('auth.invalidResetLink') }}</h3>
          <p class="rpv-state-sub">{{ t('auth.invalidResetLinkHint') }}</p>
        </div>
        <router-link to="/forgot-password" class="rpv-action-link">
          {{ t('auth.requestNewResetLink') }}
        </router-link>
      </div>

      <!-- 成功状态 -->
      <div v-else-if="isSuccess" class="rpv-state rpv-state--ok">
        <div class="rpv-state-icon-wrap rpv-state-icon-wrap--ok">
          <Icon name="checkCircle" size="lg" class="rpv-icon-ok" />
        </div>
        <div>
          <h3 class="rpv-state-title">{{ t('auth.passwordResetSuccess') }}</h3>
          <p class="rpv-state-sub">{{ t('auth.passwordResetSuccessHint') }}</p>
        </div>
        <router-link to="/login" class="rpv-submit rpv-submit--inline">
          <Icon name="login" size="md" />
          {{ t('auth.signIn') }}
        </router-link>
      </div>

      <!-- 表单 -->
      <form v-else @submit.prevent="handleSubmit" class="rpv-form">
        <!-- Email（只读） -->
        <div class="rpv-field">
          <label for="email" class="rpv-label">{{ t('auth.emailLabel') }}</label>
          <div class="rpv-inp-wrap rpv-inp-wrap--readonly">
            <Icon name="mail" size="md" class="rpv-inp-icon" />
            <input id="email" :value="email" type="email" readonly disabled class="rpv-inp" />
          </div>
        </div>

        <!-- 新密码 -->
        <div class="rpv-field">
          <label for="password" class="rpv-label">{{ t('auth.newPassword') }}</label>
          <div class="rpv-inp-wrap" :class="{ 'rpv-inp-wrap--error': errors.password }">
            <Icon name="lock" size="md" class="rpv-inp-icon" />
            <input
              id="password"
              v-model="formData.password"
              :type="showPassword ? 'text' : 'password'"
              required
              autocomplete="new-password"
              :disabled="isLoading"
              class="rpv-inp"
              :placeholder="t('auth.newPasswordPlaceholder')"
            />
            <button
              type="button"
              @click="showPassword = !showPassword"
              class="rpv-eye"
              :aria-label="showPassword ? '隐藏密码' : '显示密码'"
            >
              <Icon v-if="showPassword" name="eyeOff" size="md" />
              <Icon v-else name="eye" size="md" />
            </button>
          </div>
        </div>

        <!-- 确认密码 -->
        <div class="rpv-field">
          <label for="confirmPassword" class="rpv-label">{{ t('auth.confirmPassword') }}</label>
          <div class="rpv-inp-wrap" :class="{ 'rpv-inp-wrap--error': errors.confirmPassword }">
            <Icon name="lock" size="md" class="rpv-inp-icon" />
            <input
              id="confirmPassword"
              v-model="formData.confirmPassword"
              :type="showConfirmPassword ? 'text' : 'password'"
              required
              autocomplete="new-password"
              :disabled="isLoading"
              class="rpv-inp"
              :placeholder="t('auth.confirmPasswordPlaceholder')"
            />
            <button
              type="button"
              @click="showConfirmPassword = !showConfirmPassword"
              class="rpv-eye"
              :aria-label="showConfirmPassword ? '隐藏密码' : '显示密码'"
            >
              <Icon v-if="showConfirmPassword" name="eyeOff" size="md" />
              <Icon v-else name="eye" size="md" />
            </button>
          </div>
        </div>

        <button type="submit" :disabled="isLoading" class="rpv-submit">
          <svg v-if="isLoading" class="rpv-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          <Icon v-else name="checkCircle" size="md" />
          {{ isLoading ? t('auth.resettingPassword') : t('auth.resetPassword') }}
        </button>
      </form>
    </div>

    <template #footer>
      <p class="rpv-footer-txt">
        {{ t('auth.rememberedPassword') }}
        <router-link to="/login" class="rpv-footer-link">{{ t('auth.signIn') }}</router-link>
      </p>
    </template>
  </AuthLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { resetPassword } from '@/api/auth'

const { t } = useI18n()

// ==================== Router & Stores ====================

const route = useRoute()
const appStore = useAppStore()

// ==================== State ====================

const isLoading = ref<boolean>(false)
const isSuccess = ref<boolean>(false)
const errorMessage = ref<string>('')
const showPassword = ref<boolean>(false)
const showConfirmPassword = ref<boolean>(false)

// URL parameters
const email = ref<string>('')
const token = ref<string>('')

const formData = reactive({
  password: '',
  confirmPassword: ''
})

const errors = reactive({
  password: '',
  confirmPassword: ''
})

const validationToastMessage = computed(
  () => errors.password || errors.confirmPassword || ''
)

watch(validationToastMessage, (value, previousValue) => {
  if (value && value !== previousValue) {
    appStore.showError(value)
  }
})

// Check if the reset link is valid (has email and token)
const isInvalidLink = computed(() => !email.value || !token.value)

// ==================== Lifecycle ====================

onMounted(() => {
  // Get email and token from URL query parameters
  email.value = (route.query.email as string) || ''
  token.value = (route.query.token as string) || ''

  if (!email.value || !token.value) {
    appStore.showError(t('auth.invalidResetLink'))
  }
})

// ==================== Validation ====================

function validateForm(): boolean {
  errors.password = ''
  errors.confirmPassword = ''

  let isValid = true

  // Password validation
  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
    isValid = false
  }

  // Confirm password validation
  if (!formData.confirmPassword) {
    errors.confirmPassword = t('auth.confirmPasswordRequired')
    isValid = false
  } else if (formData.password !== formData.confirmPassword) {
    errors.confirmPassword = t('auth.passwordsDoNotMatch')
    isValid = false
  }

  return isValid
}

// ==================== Form Handlers ====================

async function handleSubmit(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    await resetPassword({
      email: email.value,
      token: token.value,
      new_password: formData.password
    })

    isSuccess.value = true
    appStore.showSuccess(t('auth.passwordResetSuccess'))
  } catch (error: unknown) {
    const err = error as { message?: string; response?: { data?: { detail?: string; code?: string } } }

    // Check for invalid/expired token error
    if (err.response?.data?.code === 'INVALID_RESET_TOKEN') {
      errorMessage.value = t('auth.invalidOrExpiredToken')
    } else if (err.response?.data?.detail) {
      errorMessage.value = err.response.data.detail
    } else if (err.message) {
      errorMessage.value = err.message
    } else {
      errorMessage.value = t('auth.resetPasswordFailed')
    }

    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}
</script>

<style scoped>
.rpv-body { display: flex; flex-direction: column; gap: 0; }
.rpv-head { margin-bottom: 24px; text-align: center; }
.rpv-title { font-size: 17px; font-weight: 700; letter-spacing: 0.04em; color: var(--ink-0); margin-bottom: 4px; }
.rpv-sub { font-size: 12px; color: var(--ink-2); }

/* 状态盒 */
.rpv-state {
  display: flex; flex-direction: column; align-items: center; gap: 14px;
  border-radius: 12px; padding: 28px 20px; text-align: center;
  margin-bottom: 8px;
}
.rpv-state--ok { background: rgba(70,201,140,.07); border: 1px solid rgba(70,201,140,.25); }
.rpv-state--warn { background: rgba(224,179,78,.07); border: 1px solid rgba(224,179,78,.25); }

.rpv-state-icon-wrap {
  width: 48px; height: 48px; border-radius: 50%; display: grid; place-items: center;
}
.rpv-state-icon-wrap--ok { background: rgba(70,201,140,.12); border: 1px solid rgba(70,201,140,.3); }
.rpv-state-icon-wrap--warn { background: rgba(224,179,78,.12); border: 1px solid rgba(224,179,78,.3); }
.rpv-icon-ok { color: #46C98C; }
.rpv-icon-warn { color: #E0B34E; }

.rpv-state-title { font-size: 15px; font-weight: 600; color: var(--ink-0); margin-bottom: 5px; }
.rpv-state-sub { font-size: 12px; color: var(--ink-2); }

.rpv-action-link {
  font-size: 13px; font-weight: 500; color: var(--ink-1); text-decoration: none;
  transition: color 0.15s ease;
}
.rpv-action-link:hover { color: var(--azure); }

/* 表单 */
.rpv-form { display: flex; flex-direction: column; gap: 0; }
.rpv-field { margin-bottom: 18px; }
.rpv-label { display: block; font-size: 12px; color: var(--ink-1); margin-bottom: 7px; }

.rpv-inp-wrap {
  display: flex; align-items: center;
  background: #0a0c0f; border: 1px solid var(--line-1);
  border-radius: 12px; padding: 0 14px; height: 46px;
  transition: box-shadow 0.25s ease, border-color 0.25s ease; gap: 10px;
}
.rpv-inp-wrap:focus-within { border-color: rgba(92,168,255,.75); box-shadow: var(--glow-focus); }
.rpv-inp-wrap--error { border-color: rgba(242,92,105,.6); }
.rpv-inp-wrap--error:focus-within { border-color: rgba(92,168,255,.75); box-shadow: var(--glow-focus); }
.rpv-inp-wrap--readonly { opacity: 0.55; pointer-events: none; }

.rpv-inp-icon { flex: none; color: var(--ink-2); }
.rpv-inp {
  flex: 1; min-width: 0; background: none; border: none; outline: none;
  color: var(--ink-0); font: inherit; font-size: 13.5px;
}
.rpv-inp::placeholder { color: var(--ink-2); }
.rpv-inp:disabled { cursor: not-allowed; }

.rpv-eye {
  flex: none; background: none; border: none; cursor: pointer;
  color: var(--ink-2); display: flex; align-items: center; padding: 0;
  transition: color 0.15s ease;
}
.rpv-eye:hover { color: var(--ink-0); }
.rpv-eye:focus-visible { outline: 1.5px solid var(--azure); outline-offset: 2px; border-radius: 4px; }

/* 按钮 */
.rpv-submit {
  width: 100%; height: 46px; border-radius: 12px;
  border: 1px solid #3a4250; background: var(--metal-raised);
  color: var(--ink-0); font: inherit; font-size: 14px; font-weight: 600;
  letter-spacing: 0.2em; cursor: pointer;
  display: inline-flex; align-items: center; justify-content: center; gap: 8px;
  box-shadow: var(--edge-hi), 0 2px 10px rgba(0,0,0,.4);
  transition: border-color 0.18s ease, box-shadow 0.18s ease;
  text-decoration: none;
}
.rpv-submit--inline { width: auto; padding: 0 24px; }
.rpv-submit:hover:not(:disabled) {
  border-color: rgba(92,168,255,.55);
  box-shadow: var(--edge-hi), 0 0 16px rgba(92,168,255,.22), 0 2px 10px rgba(0,0,0,.4);
}
.rpv-submit:focus-visible { outline: none; border-color: rgba(92,168,255,.75); box-shadow: var(--glow-focus), 0 2px 10px rgba(0,0,0,.4); }
.rpv-submit:disabled { opacity: 0.45; cursor: not-allowed; }
.rpv-submit:active:not(:disabled) { transform: scale(0.985); }

.rpv-spin { width: 16px; height: 16px; animation: rpv-spin 0.8s linear infinite; flex: none; }
@keyframes rpv-spin { to { transform: rotate(360deg); } }

.rpv-footer-txt { color: var(--ink-2); font-size: 13px; }
.rpv-footer-link { color: var(--ink-1); text-decoration: none; margin-left: 4px; transition: color 0.15s ease; }
.rpv-footer-link:hover { color: var(--azure); }

@media (prefers-reduced-motion: reduce) {
  .rpv-spin { animation: none; }
  .rpv-inp-wrap, .rpv-submit { transition: none; }
}
</style>
