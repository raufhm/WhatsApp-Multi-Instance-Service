import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import App from '@/App'
import apiClient from '@/lib/apiClient'

// Mock the apiClient used throughout the app
vi.mock('@/lib/apiClient', () => ({
  TENANT_ID_KEY: 'whatsapp_dashboard_tenant_id',
  TENANT_SLUG_KEY: 'whatsapp_dashboard_tenant_slug',
  REMEMBER_ME_KEY: 'whatsapp_dashboard_remember_me',
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  },
  authApi: {
    login: vi.fn(),
    loginWithBackupCode: vi.fn(),
    logout: vi.fn(),
    getMe: vi.fn(),
  },
  onboardingApi: {
    signupTenant: vi.fn(),
    verifyEmail: vi.fn(),
    getTOTPSetup: vi.fn(),
    verifyTOTPSetup: vi.fn(),
    getInvitation: vi.fn(),
    signupOperator: vi.fn(),
    getSetupStatus: vi.fn(),
    updateSetup: vi.fn(),
    completeSetup: vi.fn(),
  },
}))

describe('Contacts page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows placeholder when there are no contacts', async () => {
    // Mock authenticated user response
    vi.mocked(apiClient.get).mockImplementation((url: string) => {
      if (url.startsWith('/dashboard/api/me')) {
        return Promise.resolve({ data: { user: { id: 'op-1', name: 'Operator One' }, tenant_name: 'Acme Corp' } })
      }
      if (url.startsWith('/api/v1/contacts')) {
        // Return empty paginated contacts list
        return Promise.resolve({ data: { items: [], total: 0, limit: 20, offset: 0 } })
      }
      // Fallback for any other GET requests (should not be called in this test)
      return Promise.resolve({ data: {} })
    })

    // Navigate to contacts route
    window.history.pushState({}, 'Test', '/dashboard/contacts')

    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    })

    render(
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    )

    // Wait for the page heading to appear, confirming the route rendered
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /contacts/i })).toBeInTheDocument()
    })

    // Wait for the placeholder text to appear after the contacts query resolves
    await waitFor(() => {
      expect(screen.getByText(/no contacts found/i)).toBeInTheDocument()
    })
  })
})
