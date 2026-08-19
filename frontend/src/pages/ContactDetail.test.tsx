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
  is_group: false,
  tags: ['VIP', 'Returning'],
  custom_values: {
    company: 'Acme Corp',
    lead_score: 95,
  },
  created_at: '2025-01-01T10:00:00Z',
  updated_at: '2025-01-01T10:00:00Z',
}

const mockConversations = [
  {
    id: 'conv-1',
    tenant_id: 'tenant-1',
    account_id: 'acc-1',
    contact_id: 'contact-1',
    ticket_number: 1001,
    status: 'OPEN' as const,
    bot_state: 'IDLE',
    started_at: '2025-01-05T12:00:00Z',
    last_activity_at: '2025-01-05T14:00:00Z',
    closed_at: null,
    handoff_at: null,
    closure_reason: null,
    assignee: 'Alice Operator',
    merged_into_id: null,
    created_at: '2025-01-05T12:00:00Z',
    updated_at: '2025-01-05T14:00:00Z',
  },
]

const mockFieldDefinitions = [
  {
    id: 'cf-1',
    tenant_id: 'tenant-1',
    key: 'company',
    label: 'Company Name',
    field_type: 'text' as const,
    options: [],
    position: 1,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
]

describe('ContactDetail', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    vi.clearAllMocks()

    vi.spyOn(useInboxModule, 'useContactFieldDefinitions').mockReturnValue({
      data: mockFieldDefinitions,
      isLoading: false,
      isError: false,
    } as any)
    vi.spyOn(useInboxModule, 'useContactConversations').mockReturnValue({
      data: mockConversations,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)
    vi.spyOn(useInboxModule, 'useDealStages').mockReturnValue({
      data: [],
      isLoading: false,
    } as any)
    vi.spyOn(useInboxModule, 'usePipelines').mockReturnValue({
      data: [],
      isLoading: false,
    } as any)
    vi.spyOn(useInboxModule, 'useCreateContactActivity').mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)
    vi.spyOn(useInboxModule, 'useUpdateContact').mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)
  })

  it('renders contact profile, properties table, and CRM hub tabs', () => {
    vi.spyOn(useInboxModule, 'useContact').mockReturnValue({
      data: mockContact,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <ContactDetail />
      </QueryClientProvider>
    )

    // Identity and properties
    expect(screen.getAllByText('Jane Doe').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('+15550123').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('jane@example.com')).toBeInTheDocument()
    expect(screen.getByText('VIP')).toBeInTheDocument()
    expect(screen.getByText('Returning')).toBeInTheDocument()
    expect(screen.getByText('Company Name')).toBeInTheDocument()
    expect(screen.getByText('Acme Corp')).toBeInTheDocument()

    // CRM Snapshot stats
    expect(screen.getByText('Tickets')).toBeInTheDocument()
    expect(screen.getByText('Open Tasks')).toBeInTheDocument()

    // CRM Hub tabs exist
    expect(screen.getByRole('button', { name: /summary feed/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /tasks & follow-ups/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /tickets & conversations/i })).toBeInTheDocument()
  })

  it('shows an error state with retry when the contact fails to load', () => {
    const retryMock = vi.fn()
    vi.spyOn(useInboxModule, 'useContact').mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      refetch: retryMock,
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

  it('starts with empty tasks on the default Tasks & Follow-ups tab', () => {
    vi.spyOn(useInboxModule, 'useContact').mockReturnValue({
      data: mockContact,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <ContactDetail />
      </QueryClientProvider>
    )

    // Default tab is tasks - should show empty state
    expect(screen.getByText('No tasks or follow-ups yet')).toBeInTheDocument()
    expect(
      screen.getByText(/Manually add tasks and follow-ups/i)
    ).toBeInTheDocument()
    // Inline task form should be present
    expect(screen.getByPlaceholderText(/Enter task summary/i)).toBeInTheDocument()
  })

  it('submits a new task from the inline task form', async () => {
    vi.spyOn(useInboxModule, 'useContact').mockReturnValue({
      data: mockContact,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    const mutateMock = vi.fn((_input: any, opts: any) => {
      opts.onSuccess()
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

    // Fill in task form
    fireEvent.change(
      screen.getByPlaceholderText(/Enter task summary/i),
      { target: { value: 'Follow up on proposal' } }
    )
    fireEvent.click(screen.getByRole('button', { name: /add follow-up/i }))

    await waitFor(() => {
      expect(mutateMock).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'FOLLOW_UP',
          summary: 'Follow up on proposal',
          priority: 'NORMAL',
        }),
        expect.any(Object)
      )
    })
  })

  it('prevents task submit without a summary', () => {
    vi.spyOn(useInboxModule, 'useContact').mockReturnValue({
      data: mockContact,
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

    // Click Add Follow-up without entering summary
    fireEvent.click(screen.getByRole('button', { name: /add follow-up/i }))
    expect(screen.getByText('Please enter a task summary')).toBeInTheDocument()
    expect(mutateMock).not.toHaveBeenCalled()
  })

  it('switches between CRM Hub tabs: Summary Feed, Tasks, and Tickets', () => {
    vi.spyOn(useInboxModule, 'useContact').mockReturnValue({
      data: mockContact,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <ContactDetail />
      </QueryClientProvider>
    )

    // Default is tasks tab
    expect(screen.getByText('No tasks or follow-ups yet')).toBeInTheDocument()

    // Switch to Summary Feed tab
    fireEvent.click(screen.getByRole('button', { name: /summary feed/i }))
    // SummaryFeed component renders (it has its own empty state)
    expect(screen.getByText('No summary entries yet')).toBeInTheDocument()

    // Switch to Tickets tab
    fireEvent.click(screen.getByRole('button', { name: /tickets & conversations/i }))
    expect(screen.getByText('Ticket #')).toBeInTheDocument()
    expect(screen.getByText('#1001')).toBeInTheDocument()
    expect(screen.getByText('Alice Operator')).toBeInTheDocument()
  })

  it('does not auto-populate tasks from useContactActivities', () => {
    vi.spyOn(useInboxModule, 'useContact').mockReturnValue({
      data: mockContact,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    // Even if useContactActivities returns data, tasks tab should still be empty
    // because it now uses local state only
    render(
      <QueryClientProvider client={queryClient}>
        <ContactDetail />
      </QueryClientProvider>
    )

    // Tasks tab is empty by default - no auto-logged entries
    expect(screen.getByText('No tasks or follow-ups yet')).toBeInTheDocument()
  })

  it('edits contact properties and custom fields', async () => {
    vi.spyOn(useInboxModule, 'useContact').mockReturnValue({
      data: mockContact,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    const updateMutateMock = vi.fn((_input: any, opts: any) => {
      opts.onSuccess()
    })
    vi.spyOn(useInboxModule, 'useUpdateContact').mockReturnValue({
      mutate: updateMutateMock,
      isPending: false,
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <ContactDetail />
      </QueryClientProvider>
    )

    // Click Edit button
    fireEvent.click(screen.getByRole('button', { name: /edit/i }))
    expect(screen.getByText('Edit Contact Properties')).toBeInTheDocument()

    // Modify Display Name and Email
    fireEvent.change(screen.getByLabelText(/display name/i), {
      target: { value: 'Jane Smith' },
    })
    fireEvent.change(screen.getByLabelText(/email address/i), {
      target: { value: 'jane.smith@example.com' },
    })

    // Click Save Properties
    fireEvent.click(screen.getByRole('button', { name: /save properties/i }))

    await waitFor(() => {
      expect(updateMutateMock).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Jane Smith',
          email: 'jane.smith@example.com',
        }),
        expect.any(Object)
      )
    })
  })

  it('Summary Feed starts empty and does not show conversation content', () => {
    vi.spyOn(useInboxModule, 'useContact').mockReturnValue({
      data: mockContact,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <ContactDetail />
      </QueryClientProvider>
    )

    // Switch to Summary Feed tab
    fireEvent.click(screen.getByRole('button', { name: /summary feed/i }))

    // Should show empty state - no auto-logged conversation entries
    expect(screen.getByText('No summary entries yet')).toBeInTheDocument()
    // Should NOT leak ticket numbers from conversations
    expect(screen.queryByText(/#1001/i)).not.toBeInTheDocument()
  })

  it('adds a task and it appears in the task list with compact fonts', async () => {
    vi.spyOn(useInboxModule, 'useContact').mockReturnValue({
      data: mockContact,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    const mutateMock = vi.fn((_input: any, opts: any) => {
      opts.onSuccess()
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

    // Add a task
    fireEvent.change(
      screen.getByPlaceholderText(/Enter task summary/i),
      { target: { value: 'Send contract to client' } }
    )
    fireEvent.click(screen.getByRole('button', { name: /add follow-up/i }))

    await waitFor(() => {
      // Task should appear in the list
      expect(screen.getByText('Send contract to client')).toBeInTheDocument()
      // Priority badge should be visible
      expect(screen.getByText('NORMAL')).toBeInTheDocument()
      // Type label should be visible
      expect(screen.getAllByText('Follow-up').length).toBeGreaterThanOrEqual(1)
    })
  })
})
