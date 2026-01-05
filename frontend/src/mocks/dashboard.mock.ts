import type { DashboardData, MenuDay } from '@/api/types/dashboard.types'

/**
 * Generate dates for the current week starting from today
 */
function generateWeekDays(): MenuDay[] {
  const days = ['Söndag', 'Måndag', 'Tisdag', 'Onsdag', 'Torsdag', 'Fredag', 'Lördag']
  const daysShort = ['Sön', 'Mån', 'Tis', 'Ons', 'Tor', 'Fre', 'Lör']

  const meals = [
    { id: '1', name: 'Pasta Carbonara', emoji: '🍝', portions: 4 },
    { id: '2', name: 'Kycklingwok', emoji: '🥡', portions: 4 },
    { id: '3', name: 'Köttfärssås', emoji: '🍖', portions: 4 },
    { id: '4', name: 'Fiskgratäng', emoji: '🐟', portions: 4 },
    { id: '5', name: 'Tacos', emoji: '🌮', portions: 4 },
    null, // Lördag - äter ute
    { id: '6', name: 'Söndagsstek', emoji: '🥘', portions: 4 }
  ]

  const today = new Date()
  const currentDay = today.getDay()

  // Start from Monday of current week
  const monday = new Date(today)
  monday.setDate(today.getDate() - ((currentDay + 6) % 7))

  return Array.from({ length: 7 }, (_, i) => {
    const date = new Date(monday)
    date.setDate(monday.getDate() + i)

    const dayIndex = date.getDay()
    const isToday = date.toDateString() === today.toDateString()

    return {
      date: date.toISOString().split('T')[0],
      dayName: days[dayIndex],
      dayShort: daysShort[dayIndex],
      meal: meals[i],
      isToday,
      isSkipped: meals[i] === null
    }
  })
}

/**
 * Mock dashboard data
 * Simulates what the API would return
 */
export const mockDashboardData: DashboardData = {
  user: {
    id: 'user-1',
    name: 'Anna',
    email: 'anna@exempel.se',
    role: 'owner',
    avatarUrl: undefined
  },

  household: {
    id: 'household-1',
    name: 'Familjen Andersson',
    inviteCode: 'ABC123',
    members: [
      {
        id: 'user-1',
        name: 'Anna',
        role: 'owner',
        avatarUrl: undefined,
        isEatingToday: true,
        wantsLunchBox: true
      },
      {
        id: 'user-2',
        name: 'Erik',
        role: 'member',
        avatarUrl: undefined,
        isEatingToday: true,
        wantsLunchBox: false
      },
      {
        id: 'user-3',
        name: 'Lisa',
        role: 'guest',
        avatarUrl: undefined,
        isEatingToday: true,
        wantsLunchBox: true
      },
      {
        id: 'user-4',
        name: 'Oscar',
        role: 'guest',
        avatarUrl: undefined,
        isEatingToday: false,
        wantsLunchBox: false
      }
    ]
  },

  todaysMeal: {
    id: 'meal-today',
    name: 'Pasta Carbonara',
    emoji: '🍝',
    imageUrl: undefined,
    portions: 4
  },

  weeklyMenu: generateWeekDays(),

  shoppingList: {
    totalItems: 12,
    checkedItems: 3,
    categories: [
      { name: 'Mejeri', count: 4 },
      { name: 'Kött & Fisk', count: 3 },
      { name: 'Grönsaker', count: 5 }
    ]
  }
}

/**
 * Mock data for guest user view
 */
export const mockGuestDashboardData: DashboardData = {
  ...mockDashboardData,
  user: {
    id: 'user-3',
    name: 'Lisa',
    role: 'guest',
    avatarUrl: undefined
  }
}
