<script setup lang="ts">
import { onMounted } from 'vue'
import { LogOut, User } from 'lucide-vue-next'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { useUserStore } from '@/stores/user'
import { useDashboardStore } from '@/stores/dashboard'

const userStore = useUserStore()
const dashboardStore = useDashboardStore()

onMounted(async () => {
  if (!dashboardStore.householdName || dashboardStore.householdName === 'Mitt hushåll') {
    await dashboardStore.fetchDashboard()
  }
})
</script>

<template>
  <div class="profile-view">
    <div class="profile-container">
      <h1 class="profile-heading">Min profil</h1>

      <BaseCard class="profile-card">
        <div class="avatar-section">
          <div class="avatar-circle" aria-hidden="true">
            <span class="avatar-initial">{{ userStore.userInitial }}</span>
          </div>
          <div class="user-info">
            <p class="user-name">{{ userStore.userName }}</p>
            <p class="user-email">{{ userStore.currentUser?.email }}</p>
          </div>
        </div>
      </BaseCard>

      <BaseCard class="household-card">
        <div class="household-info">
          <div class="info-row">
            <User class="info-icon" :size="18" />
            <div class="info-content">
              <span class="info-label">Hushåll</span>
              <span class="info-value">{{ dashboardStore.householdName }}</span>
            </div>
          </div>
          <div class="info-row">
            <span class="info-icon count-icon">{{ dashboardStore.householdMembers.length }}</span>
            <div class="info-content">
              <span class="info-label">Medlemmar</span>
              <span class="info-value">{{ dashboardStore.householdMembers.length }} st</span>
            </div>
          </div>
        </div>
      </BaseCard>

      <div class="logout-section">
        <BaseButton variant="primary" class="logout-btn" @click="userStore.logout()">
          <LogOut :size="18" />
          Logga ut
        </BaseButton>
      </div>
    </div>
  </div>
</template>

<style scoped>
.profile-view {
  min-height: 100vh;
  background: var(--bg-primary);
  padding: var(--space-lg) var(--space-md);
}

.profile-container {
  max-width: 480px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.profile-heading {
  font-family: 'Fraunces', serif;
  font-size: 1.75rem;
  font-weight: 800;
  color: var(--text-primary);
  margin: 0 0 var(--space-sm);
}

.profile-card,
.household-card {
  width: 100%;
}

.avatar-section {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.avatar-circle {
  width: 64px;
  height: 64px;
  border-radius: var(--radius-full);
  background: var(--coral, #ff6b5b);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.avatar-initial {
  font-family: 'Fraunces', serif;
  font-size: 1.75rem;
  font-weight: 800;
  color: #fff;
  line-height: 1;
}

.user-info {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.user-name {
  font-family: 'Nunito', sans-serif;
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.user-email {
  font-family: 'Nunito', sans-serif;
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin: 0;
}

.household-info {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.info-row {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.info-icon {
  color: var(--text-secondary);
  flex-shrink: 0;
}

.count-icon {
  width: 18px;
  height: 18px;
  font-family: 'Nunito', sans-serif;
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.info-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.info-label {
  font-family: 'Nunito', sans-serif;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.info-value {
  font-family: 'Nunito', sans-serif;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.logout-section {
  display: flex;
  justify-content: center;
  padding-top: var(--space-sm);
}

.logout-btn {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  width: 100%;
  justify-content: center;
}
</style>
