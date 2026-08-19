import { describe, it, expect, vi } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import React from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider, useAuth } from './useAuth'
import apiClient from '@/lib/apiClient'

vi.mock('@/lib/apiClient', () => ({
  TENANT_ID_KEY: 'whatsapp_dashboard_tenant_id',
  TENANT_SLUG_KEY: 'whatsapp_dashboard_tenant_slug',
  REMEMBER_ME_KEY: 'whatsapp_dashboard_remember_me',
  default: {
    get: vi.fn(),
    post: vi.fn(),
  },
  authApi: {
    login: vi.fn(),
    loginWithBackupCode: vi.fn(),
    logout: vi.fn(),
    getMe: vi.fn(),
  },
}))

const createWrapper = () => {
  const queryClient = new QueryClient()
  const wrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>{children}</AuthProvider>
    </QueryClientProvider>
  )
  return wrapper
}

describe('useAuth', () => {
  it('handles unauthenticated state gracefully without errors or redirects', async () => {
    vi.mocked(apiClient.get).mockRejectedValueOnce({
      response: { status: 401, data: { error: 'Unauthorized' } },
    })

    const { result } = renderHook(() => useAuth(), { wrapper: createWrapper() })

    expect(result.current.isLoading).toBe(true)

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.user).toBeNull()
  })

  it('handles authenticated state successfully', async () => {
    const mockUser = { id: 'op-1', name: 'Operator One', email: 'op@example.com' }
    vi.mocked(apiClient.get).mockResolvedValueOnce({
      data: { user: mockUser },
    })

    const { result } = renderHook(() => useAuth(), { wrapper: createWrapper() })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.user).toEqual(mockUser)
  })

  it('logs in successfully and updates state', async () => {
    vi.mocked(apiClient.get).mockRejectedValueOnce({
      response: { status: 401 },
    })

    const mockUser = { id: 'op-2', name: 'Operator Two', email: 'op2@example.com' }
    vi.mocked(apiClient.post).mockResolvedValueOnce({
      data: { user: mockUser, tenant_id: 'tenant-1', tenant_slug: 'acme-corp' },
    })

    const { result } = renderHook(() => useAuth(), { wrapper: createWrapper() })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    await act(async () => {
      await result.current.login('acme-corp', 'op2@example.com', '123456')
    })

    await waitFor(() => {
      expect(result.current.isAuthenticated).toBe(true)
    })
    expect(result.current.user).toEqual(mockUser)
    expect(result.current.tenantSlug).toBe('acme-corp')
    expect(result.current.tenantId).toBe('tenant-1')
  })

  it('logs in with backup code and company slug successfully', async () => {
    vi.mocked(apiClient.get).mockRejectedValueOnce({
      response: { status: 401 },
    })

    const mockUser = { id: 'op-3', name: 'Operator Three', email: 'op3@example.com' }
    vi.mocked(apiClient.post).mockResolvedValueOnce({
      data: { user: mockUser, tenant_id: 'tenant-3', tenant_slug: 'acme-hq', tenant_name: 'Acme HQ' },
    })

    const { result } = renderHook(() => useAuth(), { wrapper: createWrapper() })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    await act(async () => {
      await result.current.loginWithBackupCode('Acme HQ', 'op3@example.com', 'A1B2-C3D4')
    })

    await waitFor(() => {
      expect(result.current.isAuthenticated).toBe(true)
    })
    expect(result.current.user).toEqual(mockUser)
    expect(result.current.tenantSlug).toBe('acme-hq')
    expect(result.current.tenantName).toBe('Acme HQ')
  })
})
