import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Inbox from '@/pages/Inbox'

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
}))

describe('Inbox', () => {
  it('renders the inbox title and empty state', () => {
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })

    render(
      <QueryClientProvider client={queryClient}>
        <Inbox />
      </QueryClientProvider>
    )

    expect(screen.getByText('Inbox')).toBeInTheDocument()
  })
})
