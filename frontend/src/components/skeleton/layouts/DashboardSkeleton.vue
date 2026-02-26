<script setup lang="ts">
/**
 * Skeleton layout mirroring DashboardView.vue
 *
 * Matches the exact grid, spacing, and border-radius values from:
 * - DashboardHeader.vue (sticky header)
 * - TodaysMeal.vue (32px radius, 320px min-height)
 * - WeeklyMenuGrid.vue (7-column day cards, 16px radius)
 * - QuickActions.vue (4 action rows, 20px radius)
 * - HouseholdWidget.vue (3 member rows, 20px radius)
 * - ShoppingListWidget.vue (progress bar, 20px radius)
 */
import SkeletonBlock from '../SkeletonBlock.vue'
import SkeletonCircle from '../SkeletonCircle.vue'
</script>

<template>
  <div class="dashboard-skeleton" aria-hidden="true">
    <!-- Header skeleton (mirrors DashboardHeader.vue) -->
    <header class="header-skeleton">
      <div class="header-content">
        <!-- Logo -->
        <div class="logo-skeleton">
          <SkeletonCircle size="28px" />
          <SkeletonBlock width="90px" height="22px" radius="8px" />
        </div>

        <!-- Household badge -->
        <SkeletonBlock width="140px" height="36px" radius="100px" />

        <!-- User menu -->
        <div class="user-skeleton">
          <SkeletonCircle size="36px" />
          <SkeletonBlock width="72px" height="14px" radius="6px" />
          <SkeletonBlock width="48px" height="18px" radius="100px" />
        </div>
      </div>
    </header>

    <!-- Main content -->
    <main class="content-skeleton">
      <div class="grid-skeleton">
        <!-- Main area (left column) -->
        <div class="main-area">
          <!-- TodaysMeal skeleton -->
          <div class="todays-meal-skeleton">
            <SkeletonBlock width="64px" height="24px" radius="100px" />
            <SkeletonCircle size="80px" />
            <SkeletonBlock width="60%" height="2rem" radius="12px" />
            <SkeletonBlock width="100px" height="14px" radius="8px" />
            <SkeletonBlock width="130px" height="40px" radius="100px" />
          </div>

          <!-- WeeklyMenuGrid skeleton -->
          <div class="weekly-menu-skeleton">
            <div class="menu-header-skeleton">
              <SkeletonBlock width="110px" height="16px" radius="8px" />
              <SkeletonBlock width="56px" height="22px" radius="100px" />
            </div>
            <div class="days-grid-skeleton">
              <div v-for="i in 7" :key="i" class="day-skeleton">
                <SkeletonBlock width="24px" height="10px" radius="4px" />
                <SkeletonCircle size="40px" />
              </div>
            </div>
          </div>
        </div>

        <!-- Sidebar (right column) -->
        <aside class="sidebar-skeleton">
          <!-- QuickActions skeleton -->
          <div class="widget-skeleton quick-actions-skeleton">
            <SkeletonBlock width="110px" height="14px" radius="8px" />
            <div class="actions-list-skeleton">
              <div v-for="i in 4" :key="i" class="action-skeleton">
                <SkeletonCircle size="36px" />
                <div class="action-text-skeleton">
                  <SkeletonBlock width="80%" height="13px" radius="6px" />
                  <SkeletonBlock width="55%" height="10px" radius="4px" />
                </div>
              </div>
            </div>
          </div>

          <!-- HouseholdWidget skeleton -->
          <div class="widget-skeleton household-skeleton">
            <div class="widget-header-skeleton">
              <SkeletonBlock width="75px" height="14px" radius="8px" />
              <SkeletonBlock width="72px" height="22px" radius="100px" />
            </div>
            <div class="members-list-skeleton">
              <div v-for="i in 3" :key="i" class="member-skeleton">
                <SkeletonCircle size="36px" />
                <div class="member-info-skeleton">
                  <SkeletonBlock width="70%" height="13px" radius="6px" />
                  <SkeletonBlock width="50%" height="10px" radius="4px" />
                </div>
              </div>
            </div>
            <SkeletonBlock width="100%" height="38px" radius="12px" />
          </div>

          <!-- ShoppingListWidget skeleton -->
          <div class="widget-skeleton shopping-skeleton">
            <div class="shopping-content-skeleton">
              <SkeletonCircle size="40px" />
              <div class="shopping-info-skeleton">
                <SkeletonBlock width="90px" height="14px" radius="8px" />
                <SkeletonBlock width="120px" height="10px" radius="4px" />
                <SkeletonBlock width="100%" height="6px" radius="100px" />
              </div>
            </div>
          </div>
        </aside>
      </div>
    </main>
  </div>
</template>

<style scoped>
.dashboard-skeleton {
  min-height: 100vh;
  background: var(--bg-secondary);
}

/* Header — mirrors DashboardHeader.vue */
.header-skeleton {
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border-color);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 1rem 2rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.logo-skeleton {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.user-skeleton {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

/* Content — mirrors DashboardView.vue .dashboard-content */
.content-skeleton {
  max-width: 1400px;
  margin: 0 auto;
  padding: 2rem;
}

.grid-skeleton {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 2rem;
  align-items: start;
}

.main-area {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

/* TodaysMeal — mirrors TodaysMeal.vue */
.todays-meal-skeleton {
  background: var(--bg-primary);
  border-radius: 32px;
  padding: 2.5rem;
  min-height: 320px;
  box-shadow: var(--shadow-md);
  border: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
}

/* WeeklyMenuGrid — mirrors WeeklyMenuGrid.vue */
.weekly-menu-skeleton {
  background: var(--bg-primary);
  border-radius: 24px;
  padding: 1.5rem;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
}

.menu-header-skeleton {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.25rem;
}

.days-grid-skeleton {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 0.5rem;
}

.day-skeleton {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 0.5rem;
  background: var(--bg-card);
  border: 2px solid transparent;
  border-radius: 16px;
}

/* Sidebar — mirrors DashboardView.vue .sidebar */
.sidebar-skeleton {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  position: sticky;
  top: calc(var(--header-height) + 2rem);
}

/* Shared widget styling — mirrors QuickActions/Household/Shopping */
.widget-skeleton {
  background: var(--bg-primary);
  border-radius: 20px;
  padding: 1.25rem;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
}

/* QuickActions skeleton */
.actions-list-skeleton {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-top: 1rem;
}

.action-skeleton {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 14px;
}

.action-text-skeleton {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

/* HouseholdWidget skeleton */
.widget-header-skeleton {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.members-list-skeleton {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.member-skeleton {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  background: var(--bg-card);
  border-radius: 12px;
}

.member-info-skeleton {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

/* ShoppingListWidget skeleton */
.shopping-content-skeleton {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.shopping-info-skeleton {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

/* Responsive — mirrors DashboardView.vue breakpoints */
@media (max-width: 1024px) {
  .grid-skeleton {
    grid-template-columns: 1fr;
  }

  .sidebar-skeleton {
    position: static;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 1rem;
  }
}

@media (max-width: 768px) {
  .header-content {
    padding: 0.75rem 1rem;
  }

  .content-skeleton {
    padding: 1rem;
  }

  .todays-meal-skeleton {
    padding: 2rem 1.5rem;
    min-height: 280px;
    border-radius: 24px;
  }

  .days-grid-skeleton {
    gap: 0.35rem;
  }

  .day-skeleton {
    padding: 0.5rem 0.25rem;
    border-radius: 12px;
  }

  .sidebar-skeleton {
    grid-template-columns: 1fr;
  }

  /* Hide extra header elements on mobile like DashboardHeader does */
  .logo-skeleton :deep(.skeleton-block:last-child) {
    display: none;
  }

  .user-skeleton :deep(.skeleton-block) {
    display: none;
  }
}
</style>
