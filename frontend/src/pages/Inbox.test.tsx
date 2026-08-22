import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Inbox from '@/pages/Inbox'
import * as useInboxModule from '@/hooks/useInbox'
import * as useDashboardModule from '@/hooks/useDashboard'

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
  useParams: () => ({ id: undefined }),
}))

const createQueryClient = () =>
  new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

describe('Inbox three-pane redesign', () => {
  beforeEach(() => {
    vi.spyOn(useDashboardModule, 'useDashboardAccounts').mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
    } as any)
  })

  it('renders the inbox title, search, and empty state', () => {
    render(
      <QueryClientProvider client={createQueryClient()}>
        <Inbox />
      </QueryClientProvider>
    )

    expect(screen.getByText('Inbox')).toBeInTheDocument()
    expect(screen.getByLabelText('Search conversations')).toBeInTheDocument()
    expect(screen.getByLabelText('Filter by status')).toBeInTheDocument()
    expect(screen.getByText('Select a conversation')).toBeInTheDocument()
  })

  it('defaults the status filter to Active', () => {
    render(
      <QueryClientProvider client={createQueryClient()}>
        <Inbox />
      </QueryClientProvider>
    )

    const filter = screen.getByLabelText('Filter by status') as HTMLSelectElement
    expect(filter.value).toBe('ACTIVE')
  })

  it('shows one inbox row for related conversations on the same contact', () => {
    vi.spyOn(useInboxModule, 'useInbox').mockReturnValue({
      data: [
        {
          id: 'conv-1',
          tenant_id: 't-1',
          account_id: 'a-1',
          contact_id: 'contact-1',
          ticket_number: 101,
          status: 'OPEN',
          bot_state: '',
          started_at: new Date().toISOString(),
          last_activity_at: new Date().toISOString(),
          closed_at: null,
          handoff_at: null,
          closure_reason: null,
          assignee: null,
          merged_into_id: null,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          contact_name: 'Alice Smith',
          contact_number: '15551234567',
          is_group: false,
          last_message_preview: 'Hello Alice 1',
          last_message_actor: 'CONTACT',
        },
        {
          id: 'conv-2',
          tenant_id: 't-1',
          account_id: 'a-1',
          contact_id: 'contact-1',
          ticket_number: 102,
          status: 'OPEN',
          bot_state: '',
          started_at: new Date().toISOString(),
          last_activity_at: new Date().toISOString(),
          closed_at: null,
          handoff_at: null,
          closure_reason: null,
          assignee: null,
          merged_into_id: null,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          contact_name: 'Alice Smith',
          contact_number: '15551234567',
          is_group: false,
          last_message_preview: 'Hello Alice 2',
          last_message_actor: 'CONTACT',
        },
      ],
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    render(
      <QueryClientProvider client={createQueryClient()}>
        <Inbox />
      </QueryClientProvider>
    )

    const matches = screen.getAllByText(/Alice Smith/)
    expect(matches).toHaveLength(1)
  })

  it('merges the same contact across different channels into one inbox row', () => {
    vi.spyOn(useDashboardModule, 'useDashboardAccounts').mockReturnValue({
      data: [
        { id: 'a-1', display_name: 'Sales Channel' },
        { id: 'a-2', display_name: 'Support Channel' },
      ],
      isLoading: false,
      isError: false,
    } as any)

    vi.spyOn(useInboxModule, 'useInbox').mockReturnValue({
      data: [
        {
          id: 'conv-1',
          tenant_id: 't-1',
          account_id: 'a-1',
          contact_id: 'contact-1',
          ticket_number: 201,
          status: 'OPEN',
          bot_state: '',
          started_at: new Date().toISOString(),
          last_activity_at: new Date().toISOString(),
          closed_at: null,
          handoff_at: null,
          closure_reason: null,
          assignee: null,
          merged_into_id: null,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          contact_name: 'Alice Smith',
          contact_number: '15551234567',
          is_group: false,
          last_message_preview: 'Channel one',
          last_message_actor: 'CONTACT',
        },
        {
          id: 'conv-2',
          tenant_id: 't-1',
          account_id: 'a-2',
          contact_id: 'contact-1',
          ticket_number: 202,
          status: 'OPEN',
          bot_state: '',
          started_at: new Date().toISOString(),
          last_activity_at: new Date().toISOString(),
          closed_at: null,
          handoff_at: null,
          closure_reason: null,
          assignee: null,
          merged_into_id: null,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          contact_name: 'Alice Smith',
          contact_number: '15551234567',
          is_group: false,
          last_message_preview: 'Channel two',
          last_message_actor: 'CONTACT',
        },
      ],
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    render(
      <QueryClientProvider client={createQueryClient()}>
        <Inbox />
      </QueryClientProvider>
    )

    const matches = screen.getAllByText(/Alice Smith/)
    expect(matches).toHaveLength(1)
  })

  it('renders You: indicator in conversation list preview for operator replies', () => {
    vi.spyOn(useInboxModule, 'useInbox').mockReturnValue({
      data: [
        {
          id: 'conv-1',
          tenant_id: 't-1',
          account_id: 'a-1',
          contact_id: 'contact-1',
          ticket_number: 101,
          status: 'OPEN',
          bot_state: '',
          started_at: new Date().toISOString(),
          last_activity_at: new Date().toISOString(),
          closed_at: null,
          handoff_at: null,
          closure_reason: null,
          assignee: null,
          merged_into_id: null,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          contact_name: 'Alice Smith',
          contact_number: '15551234567',
          is_group: false,
          last_message_preview: 'Hi Morning. What can i help you?',
          last_message_actor: 'OPERATOR',
        },
      ],
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    render(
      <QueryClientProvider client={createQueryClient()}>
        <Inbox />
      </QueryClientProvider>
    )

    expect(screen.getByText(/You:/)).toBeInTheDocument()
    expect(screen.getByText(/Hi Morning\. What can i help you\?/)).toBeInTheDocument()
  })
})
