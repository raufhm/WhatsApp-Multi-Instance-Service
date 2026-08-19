import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import UploadJobs from '@/pages/UploadJobs'
import * as useDashboardModule from '@/hooks/useDashboard'

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
}))

describe('UploadJobs', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    vi.clearAllMocks()
  })

  it('renders the empty state when the list response is null', () => {
    vi.spyOn(useDashboardModule, 'useUploadJobs').mockReturnValue({
      data: null,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <UploadJobs />
      </QueryClientProvider>
    )

    expect(screen.getByText('No upload jobs found.')).toBeInTheDocument()
  })

  it('renders jobs when the list response is a non-empty array', () => {
    vi.spyOn(useDashboardModule, 'useUploadJobs').mockReturnValue({
      data: [
        {
          id: 'job-1',
          tenant_id: 'tenant-1',
          message_id: 'm-1',
          host_id: 'host-1',
          object_key: 'media/k1',
          mime_type: 'image/jpeg',
          media_path: '',
          status: 'COMPLETED',
          attempt_count: 1,
          next_attempt_at: '2025-01-01T00:00:00Z',
          last_error: '',
          media_url: '',
          lease_until: null,
          created_at: '2025-01-01T00:00:00Z',
          updated_at: '2025-01-01T00:00:00Z',
        },
      ],
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <UploadJobs />
      </QueryClientProvider>
    )

    expect(screen.getByText('media/k1')).toBeInTheDocument()
  })
})