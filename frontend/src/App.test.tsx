import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import App from './App'
import apiClient from '@/lib/apiClient'

vi.mock('@/lib/apiClient', () => ({
  TENANT_ID_KEY: 'whatsapp_dashboard_tenant_id',
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

describe('App', () => {
  it('renders login page when unauthenticated', async () => {
    vi.mocked(apiClient.get).mockRejectedValueOnce({
      response: { status: 401, data: { error: 'Unauthorized' } },
    })

    window.history.pushState({}, 'Test', '/dashboard/login')

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

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
    })
  })

  it('renders inbox when authenticated', async () => {
    const mockUser = { id: 'op-1', name: 'Operator One', email: 'op@example.com' }
    vi.mocked(apiClient.get).mockImplementation((url: string) => {
      if (url === '/dashboard/api/me') {
        return Promise.resolve({ data: { user: mockUser } })
      }
      if (url === '/dashboard/api/inbox') {
        return Promise.resolve({
          data: {
            conversations: [],
            pagination: { total: 0, limit: 20, offset: 0, has_more: false },
          },
        })
      }
      if (url === '/dashboard/api/activities') {
        return Promise.resolve({ data: { activities: [] } })
      }
      return Promise.resolve({ data: {} })
    })

    window.history.pushState({}, 'Test', '/dashboard/')

    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    })

    const { container } = render(
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    )

    await waitFor(() => {
      expect(screen.getByText(/WhatsApp Dashboard/i)).toBeInTheDocument()
      expect(screen.getByRole('heading', { name: /Inbox/i })).toBeInTheDocument()
    })

    // Regression guard: the sidebar must stay fixed (pinned) on desktop, never
    // dropped into normal block flow via a static override, which previously
    // pushed the WhatsApp Workspace header ~631px down the page.
    const sidebar = container.querySelector('div.fixed.inset-y-0')
    expect(sidebar).not.toBeNull()
    expect(sidebar!.className).toContain('lg:fixed')
    expect(sidebar!.className).not.toContain('lg:static')
  })
})
