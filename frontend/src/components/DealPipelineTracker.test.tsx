import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { DealPipelineTracker } from '@/components/DealPipelineTracker'
import * as useInboxModule from '@/hooks/useInbox'

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, ...props }: any) => <a {...props}>{children}</a>,
}))

const mockStages = [
  {
    id: 'stage-1',
    tenant_id: 'tenant-1',
    key: 'NEW_LEAD',
    label: 'New Lead',
    color: '#3b82f6',
    icon: 'user-plus',
    sort_order: 1,
    is_active: true,
    is_won: false,
    is_lost: false,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
  {
    id: 'stage-2',
    tenant_id: 'tenant-1',
    key: 'APPOINTMENT_SCHEDULED',
    label: 'Appointment Scheduled',
    color: '#6366f1',
    icon: 'calendar',
    sort_order: 2,
    is_active: true,
    is_won: false,
    is_lost: false,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
  {
    id: 'stage-3',
    tenant_id: 'tenant-1',
    key: 'CLOSED_WON',
    label: 'Closed Won',
    color: '#10b981',
    icon: 'trophy',
    sort_order: 3,
    is_active: true,
    is_won: true,
    is_lost: false,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
]

const mockPipelines = [
  {
    id: 'default-sales',
    key: 'sales',
    name: 'Sales Pipeline',
    description: 'Standard sales pipeline',
    is_default: true,
  },
]

const mockHistory = [
  {
    id: 'hist-1',
    tenant_id: 'tenant-1',
    contact_id: 'contact-1',
    from_stage: 'NEW_LEAD',
    to_stage: 'APPOINTMENT_SCHEDULED',
    note: 'Demo booked for Friday',
    moved_by: 'operator-1',
    created_at: '2025-01-02T10:00:00Z',
  },
]

describe('DealPipelineTracker', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    vi.clearAllMocks()

    vi.spyOn(useInboxModule, 'usePipelines').mockReturnValue({
      data: mockPipelines,
      isLoading: false,
      isError: false,
    } as any)

    vi.spyOn(useInboxModule, 'useDealStages').mockReturnValue({
      data: mockStages,
      isLoading: false,
      isError: false,
    } as any)

    vi.spyOn(useInboxModule, 'useDealStageHistory').mockReturnValue({
      data: mockHistory,
      isLoading: false,
      isError: false,
    } as any)
  })

  it('renders Pipeline and Stage dropdowns and interactive stage step buttons', () => {
    const mutateMock = vi.fn()
    vi.spyOn(useInboxModule, 'useMoveContactToStage').mockReturnValue({
      mutate: mutateMock,
      isPending: false,
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <DealPipelineTracker contactId="contact-1" currentStageKey="APPOINTMENT_SCHEDULED" />
      </QueryClientProvider>
    )

    // Pipeline dropdown check
    expect(screen.getByLabelText('Pipeline')).toBeInTheDocument()
    expect(screen.getByText('Sales Pipeline')).toBeInTheDocument()

    // Stage dropdown check
    const stageDropdown = screen.getByLabelText('Deal Stage') as HTMLSelectElement
    expect(stageDropdown).toBeInTheDocument()
    expect(stageDropdown.value).toBe('APPOINTMENT_SCHEDULED')

    // Clickable stage buttons check
    expect(screen.getByRole('button', { name: /New Lead/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Appointment Scheduled/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Closed Won/i })).toBeInTheDocument()

    // Current stage and last transition info
    expect(screen.getByText(/Moved from NEW_LEAD/i)).toBeInTheDocument()
    expect(screen.getByText(/"Demo booked for Friday"/i)).toBeInTheDocument()
  })

  it('triggers stage transition when a stage step button is clicked', async () => {
    const mutateMock = vi.fn()
    vi.spyOn(useInboxModule, 'useMoveContactToStage').mockReturnValue({
      mutate: mutateMock,
      isPending: false,
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <DealPipelineTracker contactId="contact-1" currentStageKey="NEW_LEAD" />
      </QueryClientProvider>
    )

    const targetButton = screen.getByRole('button', { name: /Appointment Scheduled/i })
    fireEvent.click(targetButton)

    expect(mutateMock).toHaveBeenCalledWith(
      expect.objectContaining({
        stageKey: 'APPOINTMENT_SCHEDULED',
      }),
      expect.any(Object)
    )
  })

  it('triggers stage transition when changing the Stage dropdown selector', async () => {
    const mutateMock = vi.fn()
    vi.spyOn(useInboxModule, 'useMoveContactToStage').mockReturnValue({
      mutate: mutateMock,
      isPending: false,
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <DealPipelineTracker contactId="contact-1" currentStageKey="NEW_LEAD" />
      </QueryClientProvider>
    )

    const stageDropdown = screen.getByLabelText('Deal Stage')
    fireEvent.change(stageDropdown, { target: { value: 'CLOSED_WON' } })

    expect(mutateMock).toHaveBeenCalledWith(
      expect.objectContaining({
        stageKey: 'CLOSED_WON',
      }),
      expect.any(Object)
    )
  })

  it('allows adding a transition note before moving stages', async () => {
    const mutateMock = vi.fn()
    vi.spyOn(useInboxModule, 'useMoveContactToStage').mockReturnValue({
      mutate: mutateMock,
      isPending: false,
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <DealPipelineTracker contactId="contact-1" currentStageKey="NEW_LEAD" />
      </QueryClientProvider>
    )

    // Click "+ Add transition note" toggle
    fireEvent.click(screen.getByText(/\+ Add transition note/i))

    // Type a note
    const noteInput = screen.getByPlaceholderText(/Reason or note for stage change/i)
    fireEvent.change(noteInput, { target: { value: 'Customer signed contract' } })

    // Click Save Note
    fireEvent.click(screen.getByRole('button', { name: /Save Note/i }))

    expect(mutateMock).toHaveBeenCalledWith(
      expect.objectContaining({
        stageKey: 'NEW_LEAD',
        note: 'Customer signed contract',
      }),
      expect.any(Object)
    )
  })

  it('displays empty state with link to configure stages if no stages exist', () => {
    vi.spyOn(useInboxModule, 'useDealStages').mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <DealPipelineTracker contactId="contact-1" currentStageKey={null} />
      </QueryClientProvider>
    )

    expect(screen.getByText('No Deal Stages Configured')).toBeInTheDocument()
    expect(screen.getByText('Configure Deal Stages')).toBeInTheDocument()
  })
})
