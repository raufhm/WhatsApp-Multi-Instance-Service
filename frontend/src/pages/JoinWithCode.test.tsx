import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import JoinWithCode from './JoinWithCode'
import { onboardingApi } from '@/lib/apiClient'

vi.mock('@/lib/apiClient', () => ({
  onboardingApi: {
    getInvitationDetails: vi.fn(),
  },
}))

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
  Link: ({ children, to }: any) => <a href={to}>{children}</a>,
}))

describe('JoinWithCode', () => {
  it('renders input for invitation code', () => {
    render(<JoinWithCode />)
    expect(screen.getByLabelText(/invitation code/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /validate invitation/i })).toBeInTheDocument()
  })

  it('validates code and displays invitation summary', async () => {
    vi.mocked(onboardingApi.getInvitationDetails).mockResolvedValueOnce({
      id: 'inv-1',
      token: 'valid-token-123',
      tenant_id: 'tenant-1',
      tenant_name: 'Acme Corp',
      role: 'OPERATOR',
      whatsapp_number: '+1234567890',
    } as any)

    render(<JoinWithCode />)

    const input = screen.getByLabelText(/invitation code/i)
    fireEvent.change(input, { target: { value: 'valid-token-123' } })

    const button = screen.getByRole('button', { name: /validate invitation/i })
    fireEvent.click(button)

    await waitFor(() => {
      expect(screen.getByText(/Invitation Verified!/i)).toBeInTheDocument()
      expect(screen.getByText(/Acme Corp/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /continue to setup/i })).toBeInTheDocument()
    })
  })

  it('displays error message on invalid token', async () => {
    vi.mocked(onboardingApi.getInvitationDetails).mockRejectedValueOnce(
      new Error('Invitation not found or expired')
    )

    render(<JoinWithCode />)

    const input = screen.getByLabelText(/invitation code/i)
    fireEvent.change(input, { target: { value: 'invalid-token' } })

    const button = screen.getByRole('button', { name: /validate invitation/i })
    fireEvent.click(button)

    await waitFor(() => {
      expect(screen.getByText(/Invitation not found or expired/i)).toBeInTheDocument()
    })
  })
})
