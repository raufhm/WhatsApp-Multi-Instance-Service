import React, { useState, useEffect } from 'react'
import { useParams, useNavigate } from '@tanstack/react-router'
import { useInbox, useConversation } from '@/hooks/useInbox'
import { useDashboardAccounts } from '@/hooks/useDashboard'
import ConversationList from '@/components/inbox/ConversationList'
import ConversationThread from '@/components/inbox/ConversationThread'
import ConversationDetails from '@/components/inbox/ConversationDetails'
import { Bell, ChevronLeft, PanelRightClose } from 'lucide-react'

const Inbox: React.FC = () => {
  const navigate = useNavigate()
  const { id } = useParams({ strict: false })
  const selectedId = id || null

  const [detailsCollapsed, setDetailsCollapsed] = useState(() => {
    try {
      return localStorage.getItem('conversation-details-collapsed') === 'true'
    } catch {
      return false
    }
  })

  useEffect(() => {
    try {
      localStorage.setItem('conversation-details-collapsed', String(detailsCollapsed))
    } catch {
      // ignore
    }
  }, [detailsCollapsed])

  const {
    data: conversations = [],
    isLoading,
    isError,
    refetch,
  } = useInbox({ limit: 100 })

  // Load the consolidated contact history, not just the latest ticket.
  const { data: conversationData } = useConversation(selectedId || '', { limit: 500 })
  const { data: accountsData } = useDashboardAccounts()
  const accounts = Array.isArray(accountsData) ? accountsData : []
  const channels = accounts.map((account: any) => ({
    value: account.id,
    label: account.display_name || account.host_id,
  }))

  const handleSelect = (conversationId: string) => {
    if (conversationId === selectedId) return
    navigate({ to: '/conversations/$id', params: { id: conversationId } })
  }

  return (
    <div className="relative h-[calc(100vh-1.5rem)] overflow-hidden rounded-[28px] border border-white/70 bg-white/54 shadow-[0_30px_90px_-45px_rgba(15,23,42,0.55)] backdrop-blur-xl">
      <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(199,234,242,0.82)_0,rgba(199,234,242,0.82)_5.25rem,rgba(249,246,242,0.96)_5.25rem,rgba(249,246,242,0.96)_100%)]" />
      <div className="absolute inset-x-3 top-3 z-10 flex h-[5.25rem] items-center gap-4 rounded-[24px] border border-white/60 bg-white/25 px-3 shadow-sm backdrop-blur-sm">
        <div className="flex min-w-0 items-center gap-3">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-2xl rounded-bl-md bg-gradient-to-br from-orange-400 to-orange-600 text-sm font-black text-white shadow-sm">
            w
          </div>
        </div>
        <div className="ml-auto flex shrink-0 items-center gap-2">
          <button
            type="button"
            className="hidden h-11 w-11 items-center justify-center rounded-full bg-white/85 text-gray-800 shadow-sm transition-colors hover:bg-white sm:inline-flex"
            aria-label="Notifications"
            title="Notifications"
          >
            <Bell className="h-5 w-5" />
          </button>
        </div>
      </div>
      <div className="relative flex h-full overflow-hidden rounded-[28px] p-3 pt-[6.75rem]">
      <div className="relative w-80 shrink-0 overflow-hidden rounded-l-[24px] border border-r-0 border-black/10 bg-white/92 backdrop-blur-xl">
        <ConversationList
          conversations={conversations}
          channels={channels}
          selectedId={selectedId}
          isLoading={isLoading}
          isError={isError}
          onSelect={handleSelect}
          onRetry={() => refetch()}
        />
      </div>

      <div className="relative flex min-w-0 flex-1 flex-col overflow-hidden border-y border-black/10 bg-white/90">
        {selectedId ? (
          <ConversationThread conversationId={selectedId} />
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center text-gray-400 p-8">
            <div className="h-16 w-16 rounded-2xl bg-white shadow-lg shadow-orange-100 flex items-center justify-center mb-4">
              <svg
                className="h-8 w-8 text-orange-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M8 10h.01M12 10h.01M16 10h.01M9 16h6M7 4h10a2 2 0 012 2v12a2 2 0 01-2 2H7a2 2 0 01-2-2V6a2 0 012-2z"
                />
              </svg>
            </div>
            <p className="text-lg font-medium text-gray-600">Select a conversation</p>
            <p className="text-sm text-gray-500 mt-1">Choose a conversation to view its messages.</p>
          </div>
        )}
      </div>

      {/* Details Panel */}
      <div
        className={`relative overflow-y-auto rounded-r-[24px] border border-l-0 border-black/10 bg-white/92 backdrop-blur-xl transition-all duration-300 ${
          detailsCollapsed ? 'w-0 opacity-0 overflow-hidden' : 'w-80 opacity-100'
        }`}
      >
        <div className="relative min-w-[20rem]">
          <button
            type="button"
            onClick={() => setDetailsCollapsed(true)}
            className="absolute top-2 right-2 z-10 p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-full transition-colors"
            aria-label="Collapse conversation details"
            title="Collapse conversation details"
          >
            <PanelRightClose className="h-4 w-4" />
          </button>
          <ConversationDetails
            conversation={conversationData?.conversation || null}
            channelLabel={
              conversationData?.conversation
                ? channels.find((channel) => channel.value === conversationData.conversation.account_id)?.label ||
                  conversationData.conversation.account_id
                : null
            }
          />
        </div>
      </div>

      {/* Floating expand button when collapsed */}
      {detailsCollapsed && (
        <button
          type="button"
          onClick={() => setDetailsCollapsed(false)}
          className="fixed right-3 top-[4.5rem] z-20 p-2 bg-white/90 backdrop-blur border border-gray-200 rounded-full shadow-lg text-gray-500 hover:text-gray-700 hover:bg-white transition-colors"
          aria-label="Expand conversation details"
          title="Expand conversation details"
        >
          <ChevronLeft className="h-4 w-4" />
        </button>
      )}
      </div>
    </div>
  )
}

export default Inbox
