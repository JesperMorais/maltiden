<script setup lang="ts">
import { ref, computed } from 'vue'
import type { HouseholdMember } from '@/api/types/dashboard.types'
import { useUserStore } from '@/stores/user'

interface Props {
  members: HouseholdMember[]
  inviteCode: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'show-invite': []
  'remove-member': [memberId: string]
}>()

const userStore = useUserStore()

const showSettings = ref(false)
const confirmRemove = ref<HouseholdMember | null>(null)

// Count members wanting lunch box
const lunchBoxCount = computed(() =>
  props.members.filter(m => m.wantsLunchBox && m.isEatingToday).length
)

function toggleSettings() {
  showSettings.value = !showSettings.value
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
  <section class="household-widget">
    <header class="widget-header">
      <h3 class="widget-title">Hushållet</h3>
      <div class="header-actions">
        <span class="member-count">{{ members.length }} personer</span>
        <button
          v-if="userStore.isMember"
          class="settings-btn"
          :class="{ active: showSettings }"
          @click="toggleSettings"
          aria-label="Inställningar"
        >
          ⚙️
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

    <!-- Confirmation dialog (teleported to body for proper z-index) -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="confirmRemove" class="confirm-overlay" @click.self="cancelRemove">
          <div class="confirm-dialog">
            <div class="confirm-icon">⚠️</div>
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

    <!-- Lunch box summary -->
    <div v-if="lunchBoxCount > 0" class="lunchbox-summary">
      <span class="lunchbox-icon">🍱</span>
      <span class="lunchbox-text">{{ lunchBoxCount }} matlåda{{ lunchBoxCount > 1 ? 'or' : '' }} imorgon</span>
    </div>

    <div class="members-list">
      <div
        v-for="member in members"
        :key="member.id"
        class="member-item"
        :class="{ 'not-eating': !member.isEatingToday }"
      >
        <div class="member-avatar" :class="member.role">
          {{ member.name.charAt(0).toUpperCase() }}
        </div>
        <div class="member-info">
          <span class="member-name">{{ member.name }}</span>
          <span class="member-status">
            <template v-if="member.isEatingToday">
              <span class="status-dot eating"></span>
              Äter idag
              <span v-if="member.wantsLunchBox" class="lunchbox-badge" title="Vill ha matlåda">🍱</span>
            </template>
            <template v-else>
              <span class="status-dot"></span>
              Äter inte idag
            </template>
          </span>
        </div>
        <span v-if="member.role === 'owner'" class="owner-badge">Ägare</span>
        <span v-else-if="member.role === 'guest'" class="guest-badge">Gäst</span>
      </div>
    </div>

    <button class="invite-button" @click="emit('show-invite')">
      <span class="invite-icon">🔗</span>
      <span>Bjud in fler</span>
    </button>
  </section>
</template>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Nunito:wght@600;700;800&display=swap');

.household-widget {
  --coral: #ff6b5b;
  --coral-light: #ff8a7d;
  --peach: #ffb599;
  --cream: #fff8f0;
  --warm-white: #fffcf7;
  --text-dark: #3d2c29;
  --text-muted: #6b5a56;
  --green: #48bb78;
  --red: #e53e3e;

  background: var(--warm-white);
  border-radius: 20px;
  padding: 1.25rem;
  box-shadow:
    0 4px 20px rgba(61, 44, 41, 0.05),
    0 0 0 1px rgba(255, 107, 91, 0.06);
  position: relative;
}

.widget-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.widget-title {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 0.9rem;
  color: var(--text-dark);
  margin: 0;
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
  color: var(--text-muted);
  background: rgba(61, 44, 41, 0.05);
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
  background: rgba(61, 44, 41, 0.08);
  opacity: 1;
}

.settings-btn.active {
  background: rgba(255, 107, 91, 0.15);
  opacity: 1;
}

/* Settings dropdown */
.settings-dropdown {
  background: white;
  border-radius: 12px;
  padding: 0.75rem;
  margin-bottom: 1rem;
  box-shadow: 0 4px 16px rgba(61, 44, 41, 0.1);
  border: 1px solid rgba(61, 44, 41, 0.08);
}

.dropdown-label {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.75rem;
  color: var(--text-muted);
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
  background: rgba(229, 62, 62, 0.08);
}

.remove-member-btn .member-initial {
  width: 24px;
  height: 24px;
  font-size: 0.7rem;
}

.remove-member-btn .member-name {
  flex: 1;
  font-family: 'Nunito', sans-serif;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-dark);
  text-align: left;
}

.remove-icon {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 1.25rem;
  color: var(--red);
  opacity: 0;
  transition: opacity 0.2s ease;
}

.remove-member-btn:hover .remove-icon {
  opacity: 1;
}

/* Lunch box summary */
.lunchbox-summary {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 0.85rem;
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
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
  color: #92400e;
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
  background: white;
  border-radius: 12px;
  transition: all 0.2s ease;
}

.member-item.not-eating {
  opacity: 0.6;
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
  color: white;
  flex-shrink: 0;
}

.member-avatar.owner,
.member-initial.owner {
  background: linear-gradient(135deg, #f6ad55 0%, #ed8936 100%);
}

.member-avatar.member,
.member-initial.member {
  background: linear-gradient(135deg, var(--coral) 0%, var(--peach) 100%);
}

.member-avatar.guest,
.member-initial.guest {
  background: linear-gradient(135deg, #a0aec0 0%, #718096 100%);
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
  color: var(--text-dark);
}

.member-status {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-family: 'Nunito', sans-serif;
  font-size: 0.7rem;
  color: var(--text-muted);
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--text-muted);
  opacity: 0.4;
}

.status-dot.eating {
  background: var(--green);
  opacity: 1;
}

.lunchbox-badge {
  font-size: 0.75rem;
  margin-left: 0.15rem;
}

.owner-badge {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.6rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #c05621;
  background: rgba(237, 137, 54, 0.15);
  padding: 0.2rem 0.5rem;
  border-radius: 100px;
}

.guest-badge {
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.6rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  background: rgba(61, 44, 41, 0.08);
  padding: 0.2rem 0.5rem;
  border-radius: 100px;
}

.invite-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.65rem;
  background: rgba(255, 107, 91, 0.08);
  border: 1px dashed rgba(255, 107, 91, 0.3);
  border-radius: 12px;
  cursor: pointer;
  font-family: 'Nunito', sans-serif;
  font-weight: 700;
  font-size: 0.85rem;
  color: var(--coral);
  transition: all 0.3s ease;
}

.invite-button:hover {
  background: rgba(255, 107, 91, 0.15);
  border-color: var(--coral);
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
</style>

<!-- Non-scoped styles for teleported modal -->
<style>
.confirm-overlay {
  position: fixed;
  inset: 0;
  background: rgba(61, 44, 41, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 1rem;
}

.confirm-dialog {
  background: white;
  border-radius: 20px;
  padding: 2rem;
  max-width: 360px;
  width: 100%;
  text-align: center;
  box-shadow: 0 20px 60px rgba(61, 44, 41, 0.25);
}

.confirm-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.confirm-dialog h4 {
  font-family: 'Nunito', sans-serif;
  font-weight: 800;
  font-size: 1.25rem;
  color: #3d2c29;
  margin: 0 0 0.5rem;
}

.confirm-dialog p {
  font-family: 'Nunito', sans-serif;
  font-size: 0.9rem;
  color: #6b5a56;
  line-height: 1.5;
  margin: 0 0 1.5rem;
}

.confirm-dialog strong {
  color: #3d2c29;
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
  background: rgba(61, 44, 41, 0.08);
  border: none;
  color: #3d2c29;
}

.btn-cancel:hover {
  background: rgba(61, 44, 41, 0.15);
}

.btn-confirm {
  background: #e53e3e;
  border: none;
  color: white;
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
