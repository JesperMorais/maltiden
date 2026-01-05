/**
 * Unified API Error Types
 * Standardized error handling across all API calls
 */

/**
 * Standard API error response structure
 */
export interface ApiError {
  status: number
  code: string
  message: string
  details?: Record<string, unknown>
}

/**
 * Common API error codes
 */
export type ApiErrorCode =
  | 'invalid_credentials'
  | 'email_taken'
  | 'unauthorized'
  | 'forbidden'
  | 'not_found'
  | 'validation_error'
  | 'server_error'
  | 'network_error'
  | 'timeout'
  | 'unknown'

/**
 * Type guard to check if an error is an ApiError
 */
export function isApiError(error: unknown): error is ApiError {
  return (
    typeof error === 'object' &&
    error !== null &&
    'status' in error &&
    'code' in error &&
    'message' in error
  )
}

/**
 * Create a standardized ApiError from various error types
 */
export function createApiError(
  status: number,
  code: ApiErrorCode,
  message: string,
  details?: Record<string, unknown>
): ApiError {
  return { status, code, message, details }
}

/**
 * Parse an Axios error into a standardized ApiError
 */
export function parseAxiosError(error: unknown): ApiError {
  // Check if it's an Axios error with a response
  if (
    typeof error === 'object' &&
    error !== null &&
    'response' in error &&
    typeof (error as { response?: unknown }).response === 'object'
  ) {
    const axiosError = error as {
      response?: {
        status?: number
        data?: { error?: string; message?: string; code?: string }
      }
      message?: string
    }

    const status = axiosError.response?.status ?? 500
    const data = axiosError.response?.data

    // Try to extract error info from response body
    const code = (data?.code ?? data?.error ?? getErrorCodeFromStatus(status)) as ApiErrorCode
    const message = data?.message ?? getDefaultMessageForStatus(status)

    return createApiError(status, code, message)
  }

  // Check if it's a network error (no response)
  if (
    typeof error === 'object' &&
    error !== null &&
    'message' in error &&
    !('response' in error)
  ) {
    const networkError = error as { message?: string }
    if (networkError.message?.includes('timeout')) {
      return createApiError(0, 'timeout', 'Förfrågan tog för lång tid')
    }
    return createApiError(0, 'network_error', 'Kunde inte ansluta till servern')
  }

  // Unknown error type
  return createApiError(500, 'unknown', 'Ett oväntat fel uppstod')
}

/**
 * Get error code from HTTP status
 */
function getErrorCodeFromStatus(status: number): ApiErrorCode {
  switch (status) {
    case 400:
      return 'validation_error'
    case 401:
      return 'unauthorized'
    case 403:
      return 'forbidden'
    case 404:
      return 'not_found'
    case 500:
    case 502:
    case 503:
      return 'server_error'
    default:
      return 'unknown'
  }
}

/**
 * Get default Swedish error message for HTTP status
 */
function getDefaultMessageForStatus(status: number): string {
  switch (status) {
    case 400:
      return 'Ogiltig förfrågan'
    case 401:
      return 'Du måste logga in'
    case 403:
      return 'Du har inte behörighet'
    case 404:
      return 'Resursen hittades inte'
    case 500:
      return 'Serverfel - försök igen senare'
    case 502:
    case 503:
      return 'Tjänsten är tillfälligt otillgänglig'
    default:
      return 'Ett fel uppstod'
  }
}
