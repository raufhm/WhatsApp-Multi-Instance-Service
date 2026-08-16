import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import ContactDetail from '@/pages/ContactDetail'
import * as useInboxModule from '@/hooks/useInbox'

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: 'contact-1' }),
  Link: ({ children, ...props }: any) => <a {...props}>{children}</a>,
}))

const mockContact = {
  id: 'contact-1',
  tenant_id: 'tenant-1',
  name: 'Jane Doe',
  number: '+15550123',
  email: 'jane@example.com',
  tags: ['VIP', 'Returning'],
  created_at: '2025-01-01T10:00:00Z',
  updated_at: '2025-01-01T10:00:00Z',
}

const mockActivities = [
  {
    id: 'act-1',
    tenant_id: 'tenant-1',
    conversation_id: 'conv-1',
    contact_id: 'contact-1',
    type: 'FOLLOW_UP',
    summary: 'Customer requested pricing sheet',
    next_action: 'Send pricing sheet',
    priority: 'HIGH',
    status: 'OPEN',
    due_at: '2025-02-01T10:00:00Z',
    acknowledged_by: null,
    acknowledged_at: null,
    created_at: '2025-01-10T10:00:00Z',
    updated_at: '2025-01-10T10:00:00Z',
  },
]

describe('ContactDetail', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    vi.clearAllMocks()
  })

  it('renders contact profile and activity timeline', () => {
    vi.spyOn(useInboxModule, 'useContact').mockReturnValue({
      data: mockContact,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)
    vi.spyOn(useInboxModule, 'useContactActivities').mockReturnValue({
      data: mockActivities,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <ContactDetail />
      </QueryClientProvider>
    )

    expect(screen.getByText('Jane Doe')).toBeInTheDocument()
    expect(screen.getByText('+15550123')).toBeInTheDocument()
    expect(screen.getByText('jane@example.com')).toBeInTheDocument()
    expect(screen.getByText('VIP')).toBeInTheDocument()
    expect(screen.getByText('Returning')).toBeInTheDocument()
    expect(screen.getByText('Customer requested pricing sheet')).toBeInTheDocument()
    expect(screen.getByText('HIGH')).toBeInTheDocument()
  })

  it('shows an error state with retry when the contact fails to load', () => {
    const retryMock = vi.fn()
    vi.spyOn(useInboxModule, 'useContact').mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      refetch: retryMock,
    } as any)
    vi.spyOn(useInboxModule, 'useContactActivities').mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <ContactDetail />
      </QueryClientProvider>
    )

    expect(screen.getByText('Failed to load contact.')).toBeInTheDocument()
    fireEvent.click(screen.getByText('Retry'))
    expect(retryMock).toHaveBeenCalled()
  })

  it('submits a new follow-up activity', async () => {
    vi.spyOn(useInboxModule, 'useContact').mockReturnValue({
      data: mockContact,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)
    vi.spyOn(useInboxModule, 'useContactActivities').mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    const mutateMock = vi.fn((_input, { onSuccess }: any) => {
      onSuccess()
    })
    vi.spyOn(useInboxModule, 'useCreateContactActivity').mockReturnValue({
      mutate: mutateMock,
      isPending: false,
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <ContactDetail />
      </QueryClientProvider>
    )

    fireEvent.change(screen.getByLabelText('Summary'), {
      target: { value: 'Follow up on proposal' },
    })
    fireEvent.click(screen.getByRole('button', { name: /add follow-up/i }))

    await waitFor(() => {
      expect(mutateMock).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'FOLLOW_UP',
          summary: 'Follow up on proposal',
        }),
        expect.any(Object)
      )
    })
  })

  it('prevents submit without a summary', () => {
    vi.spyOn(useInboxModule, 'useContact').mockReturnValue({
      data: mockContact,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)
    vi.spyOn(useInboxModule, 'useContactActivities').mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    const mutateMock = vi.fn()
    vi.spyOn(useInboxModule, 'useCreateContactActivity').mockReturnValue({
      mutate: mutateMock,
      isPending: false,
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <ContactDetail />
      </QueryClientProvider>
    )

    fireEvent.click(screen.getByRole('button', { name: /add follow-up/i }))
    expect(screen.getByText('Please enter a follow-up summary')).toBeInTheDocument()
    expect(mutateMock).not.toHaveBeenCalled()
  })
})