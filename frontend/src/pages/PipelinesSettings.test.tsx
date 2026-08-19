import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import PipelinesSettings from '@/pages/PipelinesSettings'
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
    key: 'CLOSED_WON',
    label: 'Closed Won',
    color: '#10b981',
    icon: 'trophy',
    sort_order: 2,
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
    description: 'Standard customer sales and deal qualification pipeline',
    is_default: true,
  },
]

describe('PipelinesSettings', () => {
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
      refetch: vi.fn(),
    } as any)
    vi.spyOn(useInboxModule, 'useCreatePipeline').mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)
    vi.spyOn(useInboxModule, 'useUpdatePipeline').mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)
    vi.spyOn(useInboxModule, 'useDeletePipeline').mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any)
  })

  it('renders pipelines summary and list of configured deal stages', () => {
    render(
      <QueryClientProvider client={queryClient}>
        <PipelinesSettings />
      </QueryClientProvider>
    )

    expect(screen.getByText('Pipelines & Deal Stages')).toBeInTheDocument()
    expect(screen.getAllByText('Sales Pipeline').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Default Active Pipeline').length).toBeGreaterThanOrEqual(1)

    // Stages table
    expect(screen.getAllByText('New Lead').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('Closed Won').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('NEW_LEAD')).toBeInTheDocument()
    expect(screen.getByText('CLOSED_WON')).toBeInTheDocument()
  })

  it('opens create modal and submits new deal stage', async () => {
    const createMutateMock = vi.fn((_input, { onSuccess }: any) => {
      onSuccess()
    })
    vi.spyOn(useInboxModule, 'useCreateDealStage').mockReturnValue({
      mutate: createMutateMock,
      isPending: false,
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <PipelinesSettings />
      </QueryClientProvider>
    )

    fireEvent.click(screen.getByRole('button', { name: /Add Stage/i }))

    expect(screen.getByText('Add New Deal Stage')).toBeInTheDocument()

    fireEvent.change(screen.getByLabelText(/Stage Key/i), {
      target: { value: 'NEGOTIATION' },
    })
    fireEvent.change(screen.getByLabelText(/Stage Display Label/i), {
      target: { value: 'Negotiation In Progress' },
    })

    fireEvent.click(screen.getByRole('button', { name: /Create Stage/i }))

    await waitFor(() => {
      expect(createMutateMock).toHaveBeenCalledWith(
        expect.objectContaining({
          key: 'NEGOTIATION',
          label: 'Negotiation In Progress',
        }),
        expect.any(Object)
      )
    })
  })

  it('deletes a stage upon confirmation', async () => {
    const deleteMutateMock = vi.fn((_id, { onSuccess }: any) => {
      onSuccess()
    })
    vi.spyOn(useInboxModule, 'useDeleteDealStage').mockReturnValue({
      mutate: deleteMutateMock,
      isPending: false,
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <PipelinesSettings />
      </QueryClientProvider>
    )

    const deleteButtons = screen.getAllByTitle('Delete Stage')
    fireEvent.click(deleteButtons[0])

    const confirmButton = screen.getByRole('button', { name: /Confirm/i })
    fireEvent.click(confirmButton)

    await waitFor(() => {
      expect(deleteMutateMock).toHaveBeenCalledWith('stage-1', expect.any(Object))
    })
  })
})
