import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import BotRules from '@/pages/BotRules'
import * as useDashboardModule from '@/hooks/useDashboard'

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
}))

describe('BotRules', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    vi.clearAllMocks()
    vi.spyOn(useDashboardModule, 'useCreateBotRuleSet').mockReturnValue({
      isPending: false,
      mutate: vi.fn(),
    } as any)
    vi.spyOn(useDashboardModule, 'useActivateBotRuleSet').mockReturnValue({
      isPending: false,
      mutate: vi.fn(),
    } as any)
  })

  it('renders the empty state when the list response is null', () => {
    vi.spyOn(useDashboardModule, 'useDashboardBotRules').mockReturnValue({
      data: null,
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <BotRules />
      </QueryClientProvider>
    )

    expect(screen.getByText('No rulesets yet.')).toBeInTheDocument()
  })

  it('renders a zero rule count for a ruleset whose rules field is null', () => {
    vi.spyOn(useDashboardModule, 'useDashboardBotRules').mockReturnValue({
      data: [
        {
          id: 'rule-set-1',
          tenant_id: 'tenant-1',
          version: 1,
          rules: null,
          is_active: true,
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
        <BotRules />
      </QueryClientProvider>
    )

    expect(screen.getByText('0')).toBeInTheDocument()
  })
})