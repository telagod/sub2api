<template>
  <Card>
    <CardContent>
    <div class="border-b border-border px-6 py-4">
      <h2 class="text-lg font-semibold text-foreground">{{ t('dashboard.quickActions') }}</h2>
    </div>
    <div class="space-y-3 p-4">
      <button @click="router.push('/keys')" class="group flex w-full items-center gap-4 rounded-lg bg-accent p-4 text-left transition-all duration-200 hover:bg-muted">
        <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-md border border-border bg-metal-raised shadow-metal-edge transition-transform group-hover:scale-105">
          <Icon name="key" size="lg" class="text-primary-200" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-foreground">{{ t('dashboard.createApiKey') }}</p>
          <p class="text-xs text-muted-foreground">{{ t('dashboard.generateNewKey') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-muted-foreground transition-colors group-hover:text-foreground"
        />
      </button>

      <button @click="router.push('/usage')" class="group flex w-full items-center gap-4 rounded-lg bg-accent p-4 text-left transition-all duration-200 hover:bg-muted">
        <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-md border border-border bg-metal-raised shadow-metal-edge transition-transform group-hover:scale-105">
          <Icon name="chart" size="lg" class="text-primary-200" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-foreground">{{ t('dashboard.viewUsage') }}</p>
          <p class="text-xs text-muted-foreground">{{ t('dashboard.checkDetailedLogs') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-muted-foreground transition-colors group-hover:text-foreground"
        />
      </button>

      <button @click="router.push('/redeem')" class="group flex w-full items-center gap-4 rounded-lg bg-accent p-4 text-left transition-all duration-200 hover:bg-muted">
        <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-md border border-border bg-metal-raised shadow-metal-edge transition-transform group-hover:scale-105">
          <Icon name="gift" size="lg" class="text-primary-200" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-foreground">{{ t('dashboard.redeemCode') }}</p>
          <p class="text-xs text-muted-foreground">{{ t('dashboard.addBalanceWithCode') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-muted-foreground transition-colors group-hover:text-foreground"
        />
      </button>
    </div>
  </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { Card, CardContent } from '@/components/ui/card'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
const router = useRouter()
const { t } = useI18n()
</script>
