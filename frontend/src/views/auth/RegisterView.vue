<template>
  <AuthLayout>
    <div class="rv-body">
      <!-- 标题 -->
      <div class="rv-head">
        <h2 class="rv-title">{{ t('auth.createAccount') }}</h2>
        <p class="rv-sub">{{ t('auth.signUpToStart', { siteName }) }}</p>
      </div>

      <!-- 注册已关闭提示 -->
      <div v-if="!registrationEnabled && settingsLoaded" class="rv-notice rv-notice--warn">
        <Icon name="exclamationCircle" size="md" class="rv-notice-icon rv-notice-icon--warn" />
        <p class="rv-notice-txt rv-notice-txt--warn">{{ t('auth.registrationDisabled') }}</p>
      </div>

      <!-- 注册表单 -->
      <form v-else @submit.prevent="handleRegister" class="rv-form">
        <!-- Email -->
        <div class="rv-field">
          <label for="email" class="rv-label">{{ t('auth.emailLabel') }}</label>
          <div class="rv-inp-wrap" :class="{ 'rv-inp-wrap--error': errors.email }">
            <Icon name="mail" size="md" class="rv-inp-icon" />
            <input
              id="email"
              v-model="formData.email"
              type="email"
              required
              autofocus
              autocomplete="email"
              :disabled="registrationActionDisabled"
              class="rv-inp"
              :placeholder="t('auth.emailPlaceholder')"
            />
          </div>
        </div>

        <!-- Password -->
        <div class="rv-field">
          <label for="password" class="rv-label">{{ t('auth.passwordLabel') }}</label>
          <div class="rv-inp-wrap" :class="{ 'rv-inp-wrap--error': errors.password }">
            <Icon name="lock" size="md" class="rv-inp-icon" />
            <input
              id="password"
              v-model="formData.password"
              :type="showPassword ? 'text' : 'password'"
              required
              autocomplete="new-password"
              :disabled="registrationActionDisabled"
              class="rv-inp"
              :placeholder="t('auth.createPasswordPlaceholder')"
            />
            <button
              type="button"
              :disabled="registrationActionDisabled"
              @click="showPassword = !showPassword"
              class="rv-eye"
              :aria-label="showPassword ? '隐藏密码' : '显示密码'"
            >
              <Icon v-if="showPassword" name="eyeOff" size="md" />
              <Icon v-else name="eye" size="md" />
            </button>
          </div>
          <p class="rv-hint">{{ t('auth.passwordHint') }}</p>
        </div>

        <!-- 邀请码（必填，开启时） -->
        <div v-if="invitationCodeEnabled" class="rv-field">
          <label for="invitation_code" class="rv-label">{{ t('auth.invitationCodeLabel') }}</label>
          <div
            class="rv-inp-wrap"
            :class="{
              'rv-inp-wrap--ok': invitationValidation.valid,
              'rv-inp-wrap--error': invitationValidation.invalid || errors.invitation_code
            }"
          >
            <Icon
              name="key"
              size="md"
              class="rv-inp-icon"
              :class="invitationValidation.valid ? 'rv-inp-icon--ok' : ''"
            />
            <input
              id="invitation_code"
              v-model="formData.invitation_code"
              type="text"
              :disabled="registrationActionDisabled"
              class="rv-inp"
              :placeholder="t('auth.invitationCodePlaceholder')"
              @input="handleInvitationCodeInput"
            />
            <div v-if="invitationValidating" class="rv-inp-indicator">
              <svg class="rv-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
            </div>
            <div v-else-if="invitationValidation.valid" class="rv-inp-indicator">
              <Icon name="checkCircle" size="md" class="rv-icon-ok" />
            </div>
            <div v-else-if="invitationValidation.invalid || errors.invitation_code" class="rv-inp-indicator">
              <Icon name="exclamationCircle" size="md" class="rv-icon-err" />
            </div>
          </div>
          <transition name="rv-fade">
            <div v-if="invitationValidation.valid" class="rv-validation-ok">
              <Icon name="checkCircle" size="sm" class="rv-icon-ok" />
              <span>{{ t('auth.invitationCodeValid') }}</span>
            </div>
          </transition>
        </div>

        <!-- 推广码（选填） -->
        <div v-if="promoCodeEnabled" class="rv-field">
          <label for="promo_code" class="rv-label">
            {{ t('auth.promoCodeLabel') }}
            <span class="rv-optional">({{ t('common.optional') }})</span>
          </label>
          <div
            class="rv-inp-wrap"
            :class="{
              'rv-inp-wrap--ok': promoValidation.valid,
              'rv-inp-wrap--error': promoValidation.invalid
            }"
          >
            <Icon
              name="gift"
              size="md"
              class="rv-inp-icon"
              :class="promoValidation.valid ? 'rv-inp-icon--ok' : ''"
            />
            <input
              id="promo_code"
              v-model="formData.promo_code"
              type="text"
              :disabled="registrationActionDisabled"
              class="rv-inp"
              :placeholder="t('auth.promoCodePlaceholder')"
              @input="handlePromoCodeInput"
            />
            <div v-if="promoValidating" class="rv-inp-indicator">
              <svg class="rv-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
            </div>
            <div v-else-if="promoValidation.valid" class="rv-inp-indicator">
              <Icon name="checkCircle" size="md" class="rv-icon-ok" />
            </div>
            <div v-else-if="promoValidation.invalid" class="rv-inp-indicator">
              <Icon name="exclamationCircle" size="md" class="rv-icon-err" />
            </div>
          </div>
          <transition name="rv-fade">
            <div v-if="promoValidation.valid" class="rv-validation-ok">
              <Icon name="gift" size="sm" class="rv-icon-ok" />
              <span>{{ t('auth.promoCodeValid', { amount: promoValidation.bonusAmount?.toFixed(2) }) }}</span>
            </div>
          </transition>
        </div>

        <!-- Turnstile -->
        <div v-if="turnstileEnabled && turnstileSiteKey">
          <TurnstileWidget
            ref="turnstileRef"
            :site-key="turnstileSiteKey"
            @verify="onTurnstileVerify"
            @expire="onTurnstileExpire"
            @error="onTurnstileError"
          />
        </div>

        <!-- 登录协议 -->
        <LoginAgreementPrompt
          v-if="loginAgreementEnabled"
          :accepted="agreementAccepted"
          :documents="loginAgreementDocuments"
          :mode="loginAgreementMode"
          :updated-at="loginAgreementUpdatedAt"
          :visible="showAgreementModal"
          @accept="acceptLoginAgreement"
          @reject="rejectLoginAgreement"
          @open="showAgreementModal = true"
        />

        <!-- 注册按钮 -->
        <button
          type="submit"
          :disabled="registrationActionDisabled || (turnstileEnabled && !turnstileToken)"
          class="rv-submit"
        >
          <svg v-if="isLoading" class="rv-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          <Icon v-else name="userPlus" size="md" />
          {{
            isLoading
              ? t('auth.processing')
              : emailVerifyEnabled
                ? t('auth.continue')
                : t('auth.createAccount')
          }}
        </button>
      </form>

      <!-- OAuth -->
      <div v-if="showOAuthLogin" class="rv-oauth">
        <div class="rv-divider">
          <span class="rv-divider-line"></span>
          <span class="rv-divider-txt">{{ t('auth.oauthOrContinue') }}</span>
          <span class="rv-divider-line"></span>
        </div>
        <EmailOAuthButtons
          :disabled="registrationActionDisabled"
          :aff-code="formData.aff_code"
          :github-enabled="githubOAuthEnabled"
          :google-enabled="googleOAuthEnabled"
          :show-divider="false"
        />
        <LinuxDoOAuthSection
          v-if="linuxdoOAuthEnabled"
          :disabled="registrationActionDisabled"
          :aff-code="formData.aff_code"
          :show-divider="false"
        />
        <WechatOAuthSection
          v-if="wechatOAuthEnabled"
          :disabled="registrationActionDisabled"
          :aff-code="formData.aff_code"
          :show-divider="false"
        />
        <OidcOAuthSection
          v-if="oidcOAuthEnabled"
          :disabled="registrationActionDisabled"
          :provider-name="oidcOAuthProviderName"
          :aff-code="formData.aff_code"
          :show-divider="false"
        />
      </div>
    </div>

    <!-- 页脚 -->
    <template #footer>
      <p class="rv-footer-txt">
        {{ t('auth.alreadyHaveAccount') }}
        <router-link to="/login" class="rv-footer-link">{{ t('auth.signIn') }}</router-link>
      </p>
    </template>
  </AuthLayout>
</template>

<script setup lang="ts">
import { computed, ref, reactive, onMounted, onUnmounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import LinuxDoOAuthSection from '@/components/auth/LinuxDoOAuthSection.vue'
import OidcOAuthSection from '@/components/auth/OidcOAuthSection.vue'
import WechatOAuthSection from '@/components/auth/WechatOAuthSection.vue'
import EmailOAuthButtons from '@/components/auth/EmailOAuthButtons.vue'
import LoginAgreementPrompt from '@/components/auth/LoginAgreementPrompt.vue'
import Icon from '@/components/icons/Icon.vue'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import { useAuthStore, useAppStore } from '@/stores'
import {
  getPublicSettings,
  isWeChatWebOAuthEnabled,
  validatePromoCode,
  validateInvitationCode
} from '@/api/auth'
import { buildAuthErrorMessage } from '@/utils/authError'
import {
  formatRegistrationEmailSuffixWhitelistForMessage,
  isRegistrationEmailSuffixAllowed,
  normalizeRegistrationEmailSuffixWhitelist
} from '@/utils/registrationEmailPolicy'
import {
  clearAffiliateReferralCode,
  loadAffiliateReferralCode,
  resolveAffiliateReferralCode
} from '@/utils/oauthAffiliate'
import type { LoginAgreementDocument } from '@/types'

const { t, locale } = useI18n()
const LOGIN_AGREEMENT_STORAGE_KEY = 'sub2api_login_agreement_consent'

// ==================== Router & Stores ====================

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const appStore = useAppStore()

// ==================== State ====================

const isLoading = ref<boolean>(false)
const settingsLoaded = ref<boolean>(false)
const errorMessage = ref<string>('')
const showPassword = ref<boolean>(false)

// Public settings
const registrationEnabled = ref<boolean>(true)
const emailVerifyEnabled = ref<boolean>(false)
const promoCodeEnabled = ref<boolean>(true)
const invitationCodeEnabled = ref<boolean>(false)
const turnstileEnabled = ref<boolean>(false)
const turnstileSiteKey = ref<string>('')
const siteName = ref<string>('subme')
const linuxdoOAuthEnabled = ref<boolean>(false)
const wechatOAuthEnabled = ref<boolean>(false)
const oidcOAuthEnabled = ref<boolean>(false)
const oidcOAuthProviderName = ref<string>('OIDC')
const githubOAuthEnabled = ref<boolean>(false)
const googleOAuthEnabled = ref<boolean>(false)
const registrationEmailSuffixWhitelist = ref<string[]>([])
const loginAgreementEnabled = ref<boolean>(false)
const loginAgreementMode = ref<'modal' | 'checkbox' | string>('modal')
const loginAgreementUpdatedAt = ref<string>('')
const loginAgreementRevision = ref<string>('')
const loginAgreementDocuments = ref<LoginAgreementDocument[]>([])
const agreementAccepted = ref<boolean>(false)
const showAgreementModal = ref<boolean>(false)

// Turnstile
const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref<string>('')

// Promo code validation
const promoValidating = ref<boolean>(false)
const promoValidation = reactive({
  valid: false,
  invalid: false,
  bonusAmount: null as number | null,
  message: ''
})
let promoValidateTimeout: ReturnType<typeof setTimeout> | null = null

// Invitation code validation
const invitationValidating = ref<boolean>(false)
const invitationValidation = reactive({
  valid: false,
  invalid: false,
  message: ''
})
let invitationValidateTimeout: ReturnType<typeof setTimeout> | null = null

const formData = reactive({
  email: '',
  password: '',
  promo_code: '',
  invitation_code: '',
  aff_code: ''
})

const errors = reactive({
  email: '',
  password: '',
  turnstile: '',
  invitation_code: ''
})

const validationToastMessage = computed(() =>
  errors.email ||
  errors.password ||
  (invitationValidation.invalid ? invitationValidation.message : '') ||
  errors.invitation_code ||
  (promoValidation.invalid ? promoValidation.message : '') ||
  errors.turnstile ||
  ''
)

const showOAuthLogin = computed(
  () =>
    linuxdoOAuthEnabled.value ||
    wechatOAuthEnabled.value ||
    oidcOAuthEnabled.value ||
    githubOAuthEnabled.value ||
    googleOAuthEnabled.value
)

const agreementGateActive = computed(
  () => loginAgreementEnabled.value && !agreementAccepted.value
)

const registrationActionDisabled = computed(
  () => isLoading.value || !settingsLoaded.value || agreementGateActive.value
)

watch(validationToastMessage, (value, previousValue) => {
  if (value && value !== previousValue) {
    appStore.showError(value)
  }
})

function syncAffiliateReferralCode(): string {
  const code = resolveAffiliateReferralCode(route.query.aff, route.query.aff_code)
  if (code) {
    formData.aff_code = code
  }
  return code
}

// ==================== Lifecycle ====================

onMounted(async () => {
  syncAffiliateReferralCode()

  try {
    const settings = await getPublicSettings()
    registrationEnabled.value = settings.registration_enabled
    emailVerifyEnabled.value = settings.email_verify_enabled
    promoCodeEnabled.value = settings.promo_code_enabled
    invitationCodeEnabled.value = settings.invitation_code_enabled
    turnstileEnabled.value = settings.turnstile_enabled
    turnstileSiteKey.value = settings.turnstile_site_key || ''
    siteName.value = settings.site_name || 'subme'
    linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled
    wechatOAuthEnabled.value = isWeChatWebOAuthEnabled(settings)
    oidcOAuthEnabled.value = settings.oidc_oauth_enabled
    oidcOAuthProviderName.value = settings.oidc_oauth_provider_name || 'OIDC'
    githubOAuthEnabled.value = settings.github_oauth_enabled
    googleOAuthEnabled.value = settings.google_oauth_enabled
    registrationEmailSuffixWhitelist.value = normalizeRegistrationEmailSuffixWhitelist(
      settings.registration_email_suffix_whitelist || []
    )
    applyLoginAgreementSettings(settings)

    // Read promo code from URL parameter only if promo code is enabled
    if (promoCodeEnabled.value) {
      const promoParam = route.query.promo as string
      if (promoParam) {
        formData.promo_code = promoParam
        // Validate the promo code from URL
        await validatePromoCodeDebounced(promoParam)
      }
    }
    syncAffiliateReferralCode()
  } catch (error) {
    console.error('Failed to load public settings:', error)
    loginAgreementEnabled.value = false
    agreementAccepted.value = true
  } finally {
    settingsLoaded.value = true
  }
})

watch(
  () => [route.query.aff, route.query.aff_code],
  () => {
    syncAffiliateReferralCode()
  }
)

onUnmounted(() => {
  if (promoValidateTimeout) {
    clearTimeout(promoValidateTimeout)
  }
  if (invitationValidateTimeout) {
    clearTimeout(invitationValidateTimeout)
  }
})

// ==================== Login Agreement ====================

function applyLoginAgreementSettings(settings: {
  login_agreement_enabled?: boolean
  login_agreement_mode?: string
  login_agreement_updated_at?: string
  login_agreement_revision?: string
  login_agreement_documents?: LoginAgreementDocument[]
}): void {
  const documents = Array.isArray(settings.login_agreement_documents)
    ? settings.login_agreement_documents.filter((doc) => doc.title?.trim())
    : []
  loginAgreementDocuments.value = documents
  loginAgreementEnabled.value = settings.login_agreement_enabled === true && documents.length > 0
  loginAgreementMode.value = settings.login_agreement_mode === 'checkbox' ? 'checkbox' : 'modal'
  loginAgreementUpdatedAt.value = settings.login_agreement_updated_at || ''
  loginAgreementRevision.value =
    settings.login_agreement_revision ||
    `${loginAgreementUpdatedAt.value}:${documents.map((doc) => `${doc.id}:${doc.title}`).join('|')}`

  agreementAccepted.value = !loginAgreementEnabled.value || hasAcceptedLoginAgreement(loginAgreementRevision.value)
  showAgreementModal.value =
    loginAgreementEnabled.value && !agreementAccepted.value && loginAgreementMode.value !== 'checkbox'
}

function hasAcceptedLoginAgreement(revision: string): boolean {
  if (!revision) {
    return false
  }
  try {
    const raw = localStorage.getItem(LOGIN_AGREEMENT_STORAGE_KEY)
    if (!raw) {
      return false
    }
    const parsed = JSON.parse(raw) as { revision?: string }
    return parsed.revision === revision
  } catch {
    return false
  }
}

function acceptLoginAgreement(): void {
  if (loginAgreementRevision.value) {
    localStorage.setItem(
      LOGIN_AGREEMENT_STORAGE_KEY,
      JSON.stringify({
        revision: loginAgreementRevision.value,
        accepted_at: new Date().toISOString()
      })
    )
  }
  agreementAccepted.value = true
  showAgreementModal.value = false
}

function rejectLoginAgreement(): void {
  localStorage.removeItem(LOGIN_AGREEMENT_STORAGE_KEY)
  agreementAccepted.value = false
  showAgreementModal.value = false
  appStore.showWarning('未同意最新条款前，无法注册或使用快捷登录。')
}

// ==================== Promo Code Validation ====================

function handlePromoCodeInput(): void {
  const code = formData.promo_code.trim()

  // Clear previous validation
  promoValidation.valid = false
  promoValidation.invalid = false
  promoValidation.bonusAmount = null
  promoValidation.message = ''

  if (!code) {
    promoValidating.value = false
    return
  }

  // Debounce validation
  if (promoValidateTimeout) {
    clearTimeout(promoValidateTimeout)
  }

  promoValidateTimeout = setTimeout(() => {
    validatePromoCodeDebounced(code)
  }, 500)
}

async function validatePromoCodeDebounced(code: string): Promise<void> {
  if (!code.trim()) return

  promoValidating.value = true

  try {
    const result = await validatePromoCode(code)

    if (result.valid) {
      promoValidation.valid = true
      promoValidation.invalid = false
      promoValidation.bonusAmount = result.bonus_amount || 0
      promoValidation.message = ''
    } else {
      promoValidation.valid = false
      promoValidation.invalid = true
      promoValidation.bonusAmount = null
      // 根据错误码显示对应的翻译
      promoValidation.message = getPromoErrorMessage(result.error_code)
    }
  } catch (error) {
    console.error('Failed to validate promo code:', error)
    promoValidation.valid = false
    promoValidation.invalid = true
    promoValidation.message = t('auth.promoCodeInvalid')
  } finally {
    promoValidating.value = false
  }
}

function getPromoErrorMessage(errorCode?: string): string {
  switch (errorCode) {
    case 'PROMO_CODE_NOT_FOUND':
      return t('auth.promoCodeNotFound')
    case 'PROMO_CODE_EXPIRED':
      return t('auth.promoCodeExpired')
    case 'PROMO_CODE_DISABLED':
      return t('auth.promoCodeDisabled')
    case 'PROMO_CODE_MAX_USED':
      return t('auth.promoCodeMaxUsed')
    case 'PROMO_CODE_ALREADY_USED':
      return t('auth.promoCodeAlreadyUsed')
    default:
      return t('auth.promoCodeInvalid')
  }
}

// ==================== Invitation Code Validation ====================

function handleInvitationCodeInput(): void {
  const code = formData.invitation_code.trim()

  // Clear previous validation
  invitationValidation.valid = false
  invitationValidation.invalid = false
  invitationValidation.message = ''
  errors.invitation_code = ''

  if (!code) {
    return
  }

  // Debounce validation
  if (invitationValidateTimeout) {
    clearTimeout(invitationValidateTimeout)
  }

  invitationValidateTimeout = setTimeout(() => {
    validateInvitationCodeDebounced(code)
  }, 500)
}

async function validateInvitationCodeDebounced(code: string): Promise<void> {
  invitationValidating.value = true

  try {
    const result = await validateInvitationCode(code)

    if (result.valid) {
      invitationValidation.valid = true
      invitationValidation.invalid = false
      invitationValidation.message = ''
    } else {
      invitationValidation.valid = false
      invitationValidation.invalid = true
      invitationValidation.message = getInvitationErrorMessage(result.error_code)
    }
  } catch {
    invitationValidation.valid = false
    invitationValidation.invalid = true
    invitationValidation.message = t('auth.invitationCodeInvalid')
  } finally {
    invitationValidating.value = false
  }
}

function getInvitationErrorMessage(errorCode?: string): string {
  switch (errorCode) {
    case 'INVITATION_CODE_NOT_FOUND':
      return t('auth.invitationCodeInvalid')
    case 'INVITATION_CODE_INVALID':
      return t('auth.invitationCodeInvalid')
    case 'INVITATION_CODE_USED':
      return t('auth.invitationCodeInvalid')
    case 'INVITATION_CODE_DISABLED':
      return t('auth.invitationCodeInvalid')
    default:
      return t('auth.invitationCodeInvalid')
  }
}

// ==================== Turnstile Handlers ====================

function onTurnstileVerify(token: string): void {
  turnstileToken.value = token
  errors.turnstile = ''
}

function onTurnstileExpire(): void {
  turnstileToken.value = ''
  errors.turnstile = t('auth.turnstileExpired')
}

function onTurnstileError(): void {
  turnstileToken.value = ''
  errors.turnstile = t('auth.turnstileFailed')
}

// ==================== Validation ====================

function validateEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

function buildEmailSuffixNotAllowedMessage(): string {
  const normalizedWhitelist = normalizeRegistrationEmailSuffixWhitelist(
    registrationEmailSuffixWhitelist.value
  )
  if (normalizedWhitelist.length === 0) {
    return t('auth.emailSuffixNotAllowed')
  }
  const separator = String(locale.value || '').toLowerCase().startsWith('zh') ? '、' : ', '
  return t('auth.emailSuffixNotAllowedWithAllowed', {
    suffixes: formatRegistrationEmailSuffixWhitelistForMessage(normalizedWhitelist, {
      separator,
      more: (count) => t('auth.emailSuffixAllowedMore', { count })
    })
  })
}

function validateForm(): boolean {
  // Reset errors
  errors.email = ''
  errors.password = ''
  errors.turnstile = ''
  errors.invitation_code = ''

  let isValid = true

  if (agreementGateActive.value) {
    appStore.showWarning('请先阅读并同意最新条款后再注册。')
    if (loginAgreementMode.value !== 'checkbox') {
      showAgreementModal.value = true
    }
    return false
  }

  // Email validation
  if (!formData.email.trim()) {
    errors.email = t('auth.emailRequired')
    isValid = false
  } else if (!validateEmail(formData.email)) {
    errors.email = t('auth.invalidEmail')
    isValid = false
  } else if (
    !isRegistrationEmailSuffixAllowed(formData.email, registrationEmailSuffixWhitelist.value)
  ) {
    errors.email = buildEmailSuffixNotAllowedMessage()
    isValid = false
  }

  // Password validation
  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
    isValid = false
  }

  // Invitation code validation (required when enabled)
  if (invitationCodeEnabled.value) {
    if (!formData.invitation_code.trim()) {
      errors.invitation_code = t('auth.invitationCodeRequired')
      isValid = false
    }
  }

  // Turnstile validation
  if (turnstileEnabled.value && !turnstileToken.value) {
    errors.turnstile = t('auth.completeVerification')
    isValid = false
  }

  return isValid
}

// ==================== Form Handlers ====================

async function handleRegister(): Promise<void> {
  // Clear previous error
  errorMessage.value = ''

  // Validate form
  if (!validateForm()) {
    return
  }

  // Check promo code validation status
  if (formData.promo_code.trim()) {
    // If promo code is being validated, wait
    if (promoValidating.value) {
      errorMessage.value = t('auth.promoCodeValidating')
      return
    }
    // If promo code is invalid, block submission
    if (promoValidation.invalid) {
      errorMessage.value = t('auth.promoCodeInvalidCannotRegister')
      return
    }
  }

  // Check invitation code validation status (if enabled and code provided)
  if (invitationCodeEnabled.value) {
    // If still validating, wait
    if (invitationValidating.value) {
      errorMessage.value = t('auth.invitationCodeValidating')
      return
    }
    // If invitation code is invalid, block submission
    if (invitationValidation.invalid) {
      errorMessage.value = t('auth.invitationCodeInvalidCannotRegister')
      return
    }
    // If invitation code is required but not validated yet
    if (formData.invitation_code.trim() && !invitationValidation.valid) {
      errorMessage.value = t('auth.invitationCodeValidating')
      // Trigger validation
      await validateInvitationCodeDebounced(formData.invitation_code.trim())
      if (!invitationValidation.valid) {
        errorMessage.value = t('auth.invitationCodeInvalidCannotRegister')
        return
      }
    }
  }

  isLoading.value = true

  try {
    const affCode = formData.aff_code.trim() || loadAffiliateReferralCode()
    if (affCode) {
      formData.aff_code = affCode
    }

    // If email verification is enabled, redirect to verification page
    if (emailVerifyEnabled.value) {
      // Store registration data in sessionStorage
      sessionStorage.setItem(
        'register_data',
        JSON.stringify({
          email: formData.email,
          password: formData.password,
          turnstile_token: turnstileToken.value,
          promo_code: formData.promo_code || undefined,
          invitation_code: formData.invitation_code || undefined,
          ...(affCode ? { aff_code: affCode } : {})
        })
      )

      // Navigate to email verification page
      await router.push('/email-verify')
      return
    }

    // Otherwise, directly register
    await authStore.register({
      email: formData.email,
      password: formData.password,
      turnstile_token: turnstileEnabled.value ? turnstileToken.value : undefined,
      promo_code: formData.promo_code || undefined,
      invitation_code: formData.invitation_code || undefined,
      ...(affCode ? { aff_code: affCode } : {})
    })
    clearAffiliateReferralCode()

    // Show success toast
    appStore.showSuccess(t('auth.accountCreatedSuccess', { siteName: siteName.value }))

    // Redirect to dashboard
    await router.push('/dashboard')
  } catch (error: unknown) {
    // Reset Turnstile on error
    if (turnstileRef.value) {
      turnstileRef.value.reset()
      turnstileToken.value = ''
    }

    // Handle registration error
    errorMessage.value = buildAuthErrorMessage(error, {
      fallback: t('auth.registrationFailed')
    })

    // Also show error toast
    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}
</script>

<style scoped>
.rv-body { display: flex; flex-direction: column; gap: 0; }

.rv-head { margin-bottom: 24px; text-align: center; }
.rv-title { font-size: 17px; font-weight: 700; letter-spacing: 0.04em; color: var(--ink-0); margin-bottom: 4px; }
.rv-sub { font-size: 12px; color: var(--ink-2); }

/* 注册关闭提示 */
.rv-notice {
  display: flex; align-items: flex-start; gap: 10px;
  border-radius: 10px; padding: 12px 14px; margin-bottom: 18px;
}
.rv-notice--warn { background: rgba(224,179,78,.1); border: 1px solid rgba(224,179,78,.3); }
.rv-notice-icon--warn { color: #E0B34E; flex: none; }
.rv-notice-txt--warn { font-size: 13px; color: #E0B34E; }

.rv-form { display: flex; flex-direction: column; gap: 0; }
.rv-field { margin-bottom: 18px; }
.rv-label { display: block; font-size: 12px; color: var(--ink-1); margin-bottom: 7px; }
.rv-optional { font-size: 11px; color: var(--ink-2); margin-left: 4px; }
.rv-hint { font-size: 11px; color: var(--ink-2); margin-top: 5px; }

/* 输入框包裹 */
.rv-inp-wrap {
  display: flex; align-items: center;
  background: #0a0c0f; border: 1px solid var(--line-1);
  border-radius: 12px; padding: 0 14px; height: 46px;
  transition: box-shadow 0.25s ease, border-color 0.25s ease;
  gap: 10px;
}
.rv-inp-wrap:focus-within {
  border-color: rgba(92,168,255,.75);
  box-shadow: var(--glow-focus);
}
.rv-inp-wrap--ok { border-color: rgba(70,201,140,.55); }
.rv-inp-wrap--ok:focus-within { border-color: rgba(92,168,255,.75); box-shadow: var(--glow-focus); }
.rv-inp-wrap--error { border-color: rgba(242,92,105,.6); }
.rv-inp-wrap--error:focus-within { border-color: rgba(92,168,255,.75); box-shadow: var(--glow-focus); }

.rv-inp-icon { flex: none; color: var(--ink-2); }
.rv-inp-icon--ok { color: #46C98C; }

.rv-inp {
  flex: 1; min-width: 0;
  background: none; border: none; outline: none;
  color: var(--ink-0); font: inherit; font-size: 13.5px;
}
.rv-inp::placeholder { color: var(--ink-2); }
.rv-inp:disabled { opacity: 0.5; cursor: not-allowed; }

.rv-eye {
  flex: none; background: none; border: none; cursor: pointer;
  color: var(--ink-2); display: flex; align-items: center; padding: 0;
  transition: color 0.15s ease;
}
.rv-eye:hover { color: var(--ink-0); }
.rv-eye:focus-visible { outline: 1.5px solid var(--azure); outline-offset: 2px; border-radius: 4px; }
.rv-eye:disabled { opacity: 0.4; cursor: not-allowed; }

.rv-inp-indicator { flex: none; display: flex; align-items: center; }
.rv-icon-ok { color: #46C98C; }
.rv-icon-err { color: #F25C69; }

/* 校验成功 badge */
.rv-validation-ok {
  display: flex; align-items: center; gap: 7px;
  border-radius: 8px; border: 1px solid rgba(70,201,140,.3);
  background: rgba(70,201,140,.08); padding: 7px 12px; margin-top: 8px;
  font-size: 12px; color: #46C98C;
}

/* 主提交按钮 */
.rv-submit {
  width: 100%; height: 46px; border-radius: 12px;
  border: 1px solid #3a4250; background: var(--metal-raised);
  color: var(--ink-0); font: inherit; font-size: 14px; font-weight: 600;
  letter-spacing: 0.2em; cursor: pointer;
  display: inline-flex; align-items: center; justify-content: center; gap: 8px;
  box-shadow: var(--edge-hi), 0 2px 10px rgba(0,0,0,.4);
  transition: border-color 0.18s ease, box-shadow 0.18s ease;
  margin-bottom: 16px;
}
.rv-submit:hover:not(:disabled) {
  border-color: rgba(92,168,255,.55);
  box-shadow: var(--edge-hi), 0 0 16px rgba(92,168,255,.22), 0 2px 10px rgba(0,0,0,.4);
}
.rv-submit:focus-visible {
  outline: none; border-color: rgba(92,168,255,.75);
  box-shadow: var(--glow-focus), 0 2px 10px rgba(0,0,0,.4);
}
.rv-submit:disabled { opacity: 0.45; cursor: not-allowed; }
.rv-submit:active:not(:disabled) { transform: scale(0.985); }

/* OAuth 区 */
.rv-oauth { display: flex; flex-direction: column; gap: 8px; }
.rv-divider { display: flex; align-items: center; gap: 10px; margin-bottom: 4px; }
.rv-divider-line { flex: 1; height: 1px; background: var(--line-0); }
.rv-divider-txt { font-size: 11px; color: var(--ink-2); white-space: nowrap; }

/* 页脚 */
.rv-footer-txt { color: var(--ink-2); font-size: 13px; }
.rv-footer-link { color: var(--ink-1); text-decoration: none; margin-left: 4px; transition: color 0.15s ease; }
.rv-footer-link:hover { color: var(--azure); }

/* 转圈 */
.rv-spin { width: 16px; height: 16px; animation: rv-spin 0.8s linear infinite; flex: none; }
@keyframes rv-spin { to { transform: rotate(360deg); } }

/* 淡入 transition */
.rv-fade-enter-active, .rv-fade-leave-active { transition: opacity 0.2s ease, transform 0.2s ease; }
.rv-fade-enter-from, .rv-fade-leave-to { opacity: 0; transform: translateY(-6px); }

@media (prefers-reduced-motion: reduce) {
  .rv-spin { animation: none; }
  .rv-inp-wrap, .rv-submit { transition: none; }
  .rv-fade-enter-active, .rv-fade-leave-active { transition: none; }
}
</style>
