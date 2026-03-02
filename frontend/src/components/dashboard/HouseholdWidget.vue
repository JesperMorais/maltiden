<script setup lang="ts">
import { ref, computed } from 'vue'
import type { HouseholdMember } from '@/api/types/dashboard.types'
import { useUserStore } from '@/stores/user'
import { useDashboardStore } from '@/stores/dashboard'
import { updateMemberStatus } from '@/api/household.api'
import { useToast } from '@/composables/useToast'
import { Settings, AlertTriangle, Link, Package, UserPlus } from 'lucide-vue-next'

interface Props {
  members: HouseholdMember[]
  inviteCode: string
  horizontal?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  horizontal: false,
})

const emit = defineEmits<{
  'show-invite': []
  'remove-member': [memberId: string]
}>()

const userStore = useUserStore()
const dashboardStore = useDashboardStore()
const toast = useToast()

const showSettings = ref(false)
const confirmRemove = ref<HouseholdMember | null>(null)

// Count members wanting lunch box
const lunchBoxCount = computed(() =>
  props.members.filter(m => m.wantsLunchBox && m.isEatingToday).length
)

function toggleSettings() {
  showSettings.value = !showSettings.value
}

/**
 * Optimistically toggle a member's eating status.
 * Updates local state immediately, calls API in background, reverts on error.
 */
function toggleEating(member: HouseholdMember) {
  const newValue = !member.isEatingToday
  dashboardStore.updateMemberLocally(member.id, { isEatingToday: newValue })

  updateMemberStatus(member.id, { isEatingToday: newValue }).catch(() => {
    dashboardStore.updateMemberLocally(member.id, { isEatingToday: !newValue })
    toast.error('Kunde inte uppdatera status. Försök igen.')
  })
}

/**
 * Optimistically toggle a member's lunch box preference.
 */
function toggleLunchBox(member: HouseholdMember, event: Event) {
  event.stopPropagation()
  const newValue = !member.wantsLunchBox
  dashboardStore.updateMemberLocally(member.id, { wantsLunchBox: newValue })

  updateMemberStatus(member.id, { wantsLunchBox: newValue }).catch(() => {
    dashboardStore.updateMemberLocally(member.id, { wantsLunchBox: !newValue })
    toast.error('Kunde inte uppdatera matlådestatus. Försök igen.')
  })
}

function handleRemoveClick(member: HouseholdMember) {
  confirmRemove.value = member
  showSettings.value = false
}

function confirmRemoveMember() {
  if (confirmRemove.value) {
    emit('remove-member', confirmRemove.value.id)
    confirmRemove.value = null
  }
}

function cancelRemove() {
  confirmRemove.value = null
}
</script>

<template>
  <section class="household-widget" :class="{ horizontal: props.horizontal }">
    <header class="widget-header">
      <div class="header-left">
        <h3 class="widget-title">Hushållet</h3>
        <div v-if="horizontal && lunchBoxCount > 0" class="lunchbox-pill">
          <Package :size="14" />
          <span>{{ lunchBoxCount }} {{ lunchBoxCount > 1 ? 'matlådor' : 'matlåda' }}</span>
        </div>
      </div>
      <div class="header-actions">
        <span class="member-count">{{ members.length }} personer</span>
        <button
          v-if="userStore.isMember"
          class="settings-btn"
          :class="{ active: showSettings }"
          :aria-expanded="showSettings"
          aria-haspopup="true"
          @click="toggleSettings"
          aria-label="Inställningar"
        >
          <Settings :size="16" />
        </button>
      </div>
    </header>

    <!-- Settings dropdown -->
    <Transition name="dropdown">
      <div v-if="showSettings" class="settings-dropdown">
        <p class="dropdown-label">Ta bort medlem:</p>
        <div class="removable-members">
          <button
            v-for="member in members.filter(m => m.id !== userStore.currentUser?.id)"
            :key="member.id"
            class="remove-member-btn"
            @click="handleRemoveClick(member)"
          >
            <span class="member-initial" :class="member.role">
              {{ member.name.charAt(0).toUpperCase() }}
            </span>
            <span class="member-name">{{ member.name }}</span>
            <span class="remove-icon">×</span>
          </button>
        </div>
      </div>
    </Transition>

    <!-- Confirmation dialog -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="confirmRemove" class="confirm-overlay" @click.self="cancelRemove">
          <div class="confirm-dialog">
            <div class="confirm-icon"><AlertTriangle :size="48" color="var(--warning)" /></div>
            <h4>Ta bort {{ confirmRemove.name }}?</h4>
            <p>
              Är du säker att du vill ta bort <strong>{{ confirmRemove.name }}</strong> från hushållet?
              {{ confirmRemove.role === 'guest' ? 'Gästen' : 'Medlemmen' }} kommer inte längre ha tillgång.
            </p>
            <div class="confirm-actions">
              <button class="btn-cancel" @click="cancelRemove">Avbryt</button>
              <button class="btn-confirm" @click="confirmRemoveMember">Ta bort</button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Vertical-only lunchbox summary -->
    <div v-if="!horizontal && lunchBoxCount > 0" class="lunchbox-summary">
      <span class="lunchbox-icon"><Package :size="18" /></span>
      <span class="lunchbox-text">{{ lunchBoxCount }} {{ lunchBoxCount > 1 ? 'matlådor' : 'matlåda' }} imorgon</span>
    </div>

    <div class="members-list">
      <div
        v-for="member in members"
        :key="member.id"
        class="member-item"
        :class="{ 'not-eating': !member.isEatingToday }"
      >
        <!-- Avatar with status ring -->
        <div class="avatar-wrapper" :class="{ eating: member.isEatingToday }">
          <div class="member-avatar" :class="member.role">
            {{ member.name.charAt(0).toUpperCase() }}
          </div>
          <div v-if="member.wantsLunchBox && member.isEatingToday" class="lunchbox-indicator">
            <Package :size="10" />
          </div>
        </div>
        <div class="member-info">
          <span class="member-name">{{ member.name }}</span>
          <span class="member-status">
            <template v-if="member.isEatingToday">
              <span class="status-dot eating"></span>
              Äter idag
            </template>
            <template v-else>
              <span class="status-dot"></span>
              Äter inte idag
            </template>
          </span>
        </div>
        <div class="member-actions">
          <button
            v-if="userStore.isMember"
            class="eating-toggle"
            :class="{ active: member.isEatingToday }"
            :aria-label="`${member.name}: ${member.isEatingToday ? 'äter idag' : 'äter inte idag'}. Klicka för att ändra.`"
            @click="toggleEating(member)"
          >
            <span class="eating-toggle-icon">{{ member.isEatingToday ? '✓' : '✕' }}</span>
          </button>
          <button
            v-if="member.isEatingToday && userStore.isMember"
            class="lunchbox-toggle"
            :class="{ active: member.wantsLunchBox }"
            :title="member.wantsLunchBox ? 'Ta bort matlåda' : 'Lägg till matlåda'"
            :aria-label="member.wantsLunchBox ? `Ta bort matlåda för ${member.name}` : `Lägg till matlåda för ${member.name}`"
            @click="toggleLunchBox(member, $event)"
          >
            <Package :size="16" />
          </button>
        </div>
        <span v-if="member.role === 'owner'" class="owner-badge">Ägare</span>
        <span v-else-if="member.role === 'guest'" class="guest-badge">Gäst</span>
      </div>

      <!-- Invite as last item in the row (horizontal only) -->
      <button v-if="horizontal" class="invite-avatar" @click="emit('show-invite')" aria-label="Bjud in fler">
        <div class="invite-circle">
          <UserPlus :size="20" />
        </div>
        <span class="invite-label">Bjud in</span>
      </button>
    </div>

    <!-- Vertical-only invite button -->
    <button v-if="!horizontal" class="invite-button" @click="emit('show-invite')">
      <span class="invite-icon"><Link :size="16" /></span>
      <span>Bjud in fler</span>
    </button>
  </section>
</template>

<style scoped>
.household-widget {
  background: var(--bg-primary);
  border-radius: 20px;
  padding: 1.25rem;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
  position: relative;
}

.widget-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.widget-title {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 0.9rem;
  color: var(--text-primary);
  margin: 0;
}

.lunchbox-pill {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.2rem 0.6rem;
  background: linear-gradient(135deg, var(--warning-surface) 0%, var(--warning-surface-end) 100%);
  border-radius: 100px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.7rem;
  color: var(--warning-dark);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.member-count {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.75rem;
  color: var(--text-secondary);
  background: var(--bg-hover);
  padding: 0.2rem 0.6rem;
  border-radius: 100px;
}

.settings-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 1rem;
  transition: all 0.2s ease;
  opacity: 0.6;
}

.settings-btn:hover {
  background: var(--bg-hover);
  opacity: 1;
}

.settings-btn.active {
  background: var(--bg-hover);
  opacity: 1;
}

/* Settings dropdown */
.settings-dropdown {
  background: var(--bg-card);
  border-radius: 12px;
  padding: 0.75rem;
  margin-bottom: 1rem;
  box-shadow: var(--shadow-md);
  border: 1px solid var(--border-color);
}

.dropdown-label {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin: 0 0 0.5rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.removable-members {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.remove-member-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.5rem;
  background: transparent;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.remove-member-btn:hover {
  background: var(--error-bg);
}

.remove-member-btn .member-initial {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 0.7rem;
  color: var(--text-on-accent);
  flex-shrink: 0;
}

.remove-member-btn .member-name {
  flex: 1;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-primary);
  text-align: left;
}

.remove-icon {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1.25rem;
  color: var(--error);
  opacity: 0;
  transition: opacity 0.2s ease;
}

.remove-member-btn:hover .remove-icon {
  opacity: 1;
}

/* Lunch box summary (vertical only) */
.lunchbox-summary {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 0.85rem;
  background: linear-gradient(135deg, var(--warning-surface) 0%, var(--warning-surface-end) 100%);
  border-radius: 10px;
  margin-bottom: 1rem;
}

.lunchbox-icon {
  font-size: 1.1rem;
}

.lunchbox-text {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.8rem;
  color: var(--warning-dark);
}

/* ═══ Avatar wrapper with status ring ═══ */
.avatar-wrapper {
  position: relative;
  flex-shrink: 0;
  padding: 2px;
  border-radius: 50%;
  border: 2px solid transparent;
  transition: border-color 0.2s ease;
}

.avatar-wrapper.eating {
  border-color: var(--success);
}

.lunchbox-indicator {
  position: absolute;
  bottom: -2px;
  right: -2px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--warning);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--bg-primary);
  border: 2px solid var(--bg-card);
}

/* Members list */
.members-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.member-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  background: var(--bg-card);
  border-radius: 12px;
  transition: all 0.2s ease;
}

.member-item.not-eating .avatar-wrapper {
  opacity: 0.5;
}

.member-item.not-eating .member-name {
  color: var(--text-muted);
}

.member-item.not-eating .member-status {
  color: var(--text-muted);
}

.member-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 0.9rem;
  color: var(--text-on-accent);
  flex-shrink: 0;
}

.member-avatar.owner,
.member-initial.owner {
  background: var(--role-owner-bg);
}

.member-avatar.member,
.member-initial.member {
  background: var(--role-member-bg);
}

.member-avatar.guest,
.member-initial.guest {
  background: var(--role-guest-avatar);
}

.member-info {
  flex: 1;
  min-width: 0;
}

.member-name {
  display: block;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.85rem;
  color: var(--text-primary);
}

.member-status {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-family: 'Nunito', sans-serif;
  font-size: 0.7rem;
  color: var(--text-secondary);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--text-secondary);
  opacity: 0.4;
  position: relative;
}

.status-dot.eating {
  background: var(--success);
  opacity: 1;
}

.status-dot.eating::after {
  content: '';
  position: absolute;
  top: 1px;
  left: 2px;
  width: 3px;
  height: 5px;
  border: solid white;
  border-width: 0 1.5px 1.5px 0;
  transform: rotate(45deg);
}

/* Action buttons wrapper */
.member-actions {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.lunchbox-toggle {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1.5px solid var(--border-color);
  border-radius: 10px;
  font-size: 1rem;
  cursor: pointer;
  opacity: 0.35;
  transition: all 0.2s ease;
  flex-shrink: 0;
  -webkit-tap-highlight-color: transparent;
}

.lunchbox-toggle:hover {
  opacity: 0.7;
  border-color: var(--border-color-hover);
}

.lunchbox-toggle.active {
  opacity: 1;
  background: var(--warning-bg);
  border-color: #edc53f;
}

.owner-badge {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--role-owner-text);
  background: var(--role-owner-text-bg);
  padding: 0.2rem 0.5rem;
  border-radius: 100px;
}

.guest-badge {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--role-guest-text);
  background: var(--role-guest-bg);
  padding: 0.2rem 0.5rem;
  border-radius: 100px;
}

.eating-toggle {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1.5px solid var(--border-color);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
  flex-shrink: 0;
  -webkit-tap-highlight-color: transparent;
}

.eating-toggle .eating-toggle-icon {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.eating-toggle.active {
  background: var(--success-bg);
  border-color: var(--success);
}

.eating-toggle.active .eating-toggle-icon {
  color: var(--success-dark);
}

.eating-toggle:hover {
  border-color: var(--accent);
}

.invite-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.65rem;
  background: var(--bg-hover);
  border: 1px dashed var(--border-color-hover);
  border-radius: 12px;
  cursor: pointer;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.85rem;
  color: var(--accent-text);
  transition: all 0.3s ease;
}

.invite-button:hover {
  border-color: var(--accent);
}

.invite-icon {
  font-size: 1rem;
}

/* Transitions */
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.25s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* ═══ Horizontal layout ═══ */
.household-widget.horizontal {
  border-radius: 24px;
  padding: 1.25rem 1.5rem;
}

.horizontal .widget-header {
  margin-bottom: 0.75rem;
}

.horizontal .members-list {
  flex-direction: row;
  gap: 0.5rem;
  margin-bottom: 0;
  overflow-x: auto;
  scrollbar-width: none;
  -ms-overflow-style: none;
  align-items: stretch;
}

.horizontal .members-list::-webkit-scrollbar {
  display: none;
}

.horizontal .member-item {
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 0.35rem;
  padding: 0.75rem 0.5rem 0.6rem;
  border-radius: 16px;
  min-width: 0;
  flex: 1;
}

.horizontal .avatar-wrapper {
  padding: 3px;
  border-width: 2.5px;
}

.horizontal .member-avatar {
  width: 42px;
  height: 42px;
  font-size: 1rem;
}

.horizontal .member-info {
  flex: none;
  min-width: 0;
  width: 100%;
}

.horizontal .member-name {
  font-size: 0.78rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.horizontal .member-status {
  justify-content: center;
  font-size: 0.65rem;
}

.horizontal .member-actions {
  justify-content: center;
}

.horizontal .member-item .eating-toggle,
.horizontal .member-item .lunchbox-toggle {
  width: 30px;
  height: 30px;
  border-radius: 8px;
}

.horizontal .member-item .eating-toggle-icon {
  font-size: 0.65rem;
}

.horizontal .member-item .lunchbox-toggle :deep(svg) {
  width: 13px;
  height: 13px;
}

.horizontal .owner-badge,
.horizontal .guest-badge {
  font-size: 0.6rem;
  padding: 0.1rem 0.35rem;
}

/* Invite circle — lives inside the members row */
.invite-avatar {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  min-width: 72px;
  padding: 0.75rem 0.5rem;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.invite-avatar:hover .invite-circle {
  border-color: var(--accent);
  color: var(--accent);
  background: var(--bg-hover);
}

.invite-circle {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  border: 2px dashed var(--border-color-hover);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--accent-text);
  opacity: 0.6;
  transition: all 0.2s ease;
}

.invite-avatar:hover .invite-circle {
  opacity: 1;
}

.invite-label {
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.7rem;
  color: var(--accent-text);
  opacity: 0.6;
}

.invite-avatar:hover .invite-label {
  opacity: 1;
}

@media (max-width: 768px) {
  .household-widget.horizontal {
    padding: 1rem;
  }

  .horizontal .member-avatar {
    width: 36px;
    height: 36px;
    font-size: 0.85rem;
  }

  .invite-circle {
    width: 36px;
    height: 36px;
  }

  .invite-circle :deep(svg) {
    width: 16px;
    height: 16px;
  }
}
</style>

<!-- Non-scoped styles for teleported modal -->
<style>
.confirm-overlay {
  position: fixed;
  inset: 0;
  background: var(--overlay-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 1rem;
}

.confirm-dialog {
  background: var(--bg-primary);
  border-radius: 20px;
  padding: 2rem;
  max-width: 360px;
  width: 100%;
  text-align: center;
  box-shadow: var(--shadow-lg);
  border: 1px solid var(--border-color);
}

.confirm-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.confirm-dialog h4 {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 1.25rem;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.confirm-dialog p {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: var(--text-secondary);
  line-height: 1.5;
  margin: 0 0 1.5rem;
}

.confirm-dialog strong {
  color: var(--text-primary);
}

.confirm-actions {
  display: flex;
  gap: 0.75rem;
}

.btn-cancel,
.btn-confirm {
  flex: 1;
  padding: 0.75rem 1rem;
  border-radius: 12px;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-cancel {
  background: var(--bg-hover);
  border: none;
  color: var(--text-primary);
}

.btn-cancel:hover {
  background: var(--border-color);
}

.btn-confirm {
  background: #e53e3e;
  border: none;
  color: var(--text-on-accent);
}

.btn-confirm:hover {
  background: #c53030;
  transform: translateY(-1px);
}

/* Modal transitions */
.modal-enter-active,
.modal-leave-active {
  transition: all 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .confirm-dialog,
.modal-leave-to .confirm-dialog {
  transform: scale(0.9);
}
</style>
