import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import ConversationDetail from '@/pages/ConversationDetail'
import * as useInboxModule from '@/hooks/useInbox'

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: 'conv-123' }),
  useNavigate: () => vi.fn(),
}))

describe('ConversationDetail media support', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    vi.clearAllMocks()
  })

  it('renders image, video, audio, and document media messages in timeline', () => {
    vi.spyOn(useInboxModule, 'useConversation').mockReturnValue({
      data: {
        conversation: {
          id: 'conv-123',
          tenant_id: 'tenant-1',
          account_id: 'acc-1',
          contact_id: 'contact-1',
          ticket_number: 42,
          status: 'OPEN',
          bot_state: '',
          started_at: new Date().toISOString(),
          last_activity_at: new Date().toISOString(),
          closed_at: null,
          handoff_at: null,
          closure_reason: null,
          assignee: 'Alice',
          merged_into_id: null,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
        messages: [
          {
            id: 'msg-img',
            tenant_id: 'tenant-1',
            conversation_id: 'conv-123',
            actor: 'CONTACT',
            provider: 'whatsmeow',
            provider_message_id: 'p-1',
            direction: 'INCOMING',
            content: 'Check out this photo',
            message_type: 'IMAGE',
            media_url: '/api/v1/media/media/photo.jpg',
            status: 'SENT',
            provider_timestamp: new Date().toISOString(),
            is_internal: false,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          },
          {
            id: 'msg-doc',
            tenant_id: 'tenant-1',
            conversation_id: 'conv-123',
            actor: 'OPERATOR',
            provider: 'whatsmeow',
            provider_message_id: 'p-2',
            direction: 'OUTGOING',
            content: 'Here is the invoice',
            message_type: 'DOCUMENT',
            media_url: '/api/v1/media/media/invoice.pdf',
            status: 'SENT',
            provider_timestamp: new Date().toISOString(),
            is_internal: false,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
          },
        ],
      },
      isLoading: false,
      isError: false,
    } as any)

    render(
      <QueryClientProvider client={queryClient}>
        <ConversationDetail />
      </QueryClientProvider>
    )

    // Image rendered with /dashboard/api/media URL translation
    const img = screen.getByAltText('Attachment')
    expect(img).toBeInTheDocument()
    expect(img.getAttribute('src')).toBe('/dashboard/api/media/media/photo.jpg')

    // Document link rendered
    const docLink = screen.getByText('View attachment')
    expect(docLink).toBeInTheDocument()
    expect(docLink.getAttribute('href')).toBe('/dashboard/api/media/media/invoice.pdf')
  })

  it('handles paperclip file selection and attachment upload', async () => {
    vi.spyOn(useInboxModule, 'useConversation').mockReturnValue({
      data: {
        conversation: {
          id: 'conv-123',
          tenant_id: 'tenant-1',
          account_id: 'acc-1',
          contact_id: 'contact-1',
          ticket_number: 42,
          status: 'OPEN',
          bot_state: '',
          started_at: new Date().toISOString(),
          last_activity_at: new Date().toISOString(),
          closed_at: null,
          handoff_at: null,
          closure_reason: null,
          assignee: 'Alice',
          merged_into_id: null,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
        messages: [],
      },
      isLoading: false,
      isError: false,
    } as any)

    const mutateAsyncMock = vi.fn().mockResolvedValue({
      media_key: 'media/new-file.png',
      media_url: '/dashboard/api/media/media/new-file.png',
      mime_type: 'image/png',
      size: 1024,
    })

    vi.spyOn(useInboxModule, 'useUploadMedia').mockReturnValue({
      mutateAsync: mutateAsyncMock,
      isPending: false,
    } as any)

    const sendMessageMock = vi.fn()
    vi.spyOn(useInboxModule, 'useSendMessage').mockReturnValue({
      mutate: sendMessageMock,
      isPending: false,
    } as any)

    const { container } = render(
      <QueryClientProvider client={queryClient}>
        <ConversationDetail />
      </QueryClientProvider>
    )

    const fileInput = container.querySelector('input[type="file"]') as HTMLInputElement
    expect(fileInput).toBeInTheDocument()

    const testFile = new File(['fake content'], 'test-image.png', { type: 'image/png' })
    fireEvent.change(fileInput, { target: { files: [testFile] } })

    await waitFor(() => {
      expect(mutateAsyncMock).toHaveBeenCalledWith(
        expect.objectContaining({
          file: testFile,
        })
      )
    })

    // Expect preview banner with file name
    await waitFor(() => {
      expect(screen.getByText('test-image.png')).toBeInTheDocument()
    })
  })
})
