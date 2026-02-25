import apiClient from './client'
import { USE_MOCKS } from '@/mocks'

export interface FeedbackRequest {
  mood: 'good' | 'okay' | 'bad'
  categories?: string[]
  comment?: string
  page?: string
  viewportWidth?: number
}

export interface FeedbackResponse {
  id: string
}

export async function submitFeedback(req: FeedbackRequest): Promise<FeedbackResponse> {
  if (USE_MOCKS) {
    await new Promise((r) => setTimeout(r, 500))
    return { id: 'fb_mock-' + Date.now() }
  }

  const { data } = await apiClient.post<FeedbackResponse>('/feedback', req)
  return data
}
