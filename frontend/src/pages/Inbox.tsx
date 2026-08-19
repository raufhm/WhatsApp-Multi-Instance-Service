import React, { useState, useEffect } from 'react'
import { useParams, useNavigate } from '@tanstack/react-router'
import { useInbox, useConversation } from '@/hooks/useInbox'
import ConversationList from '@/components/inbox/ConversationList'
import ConversationThread from '@/components/inbox/ConversationThread'
import ConversationDetails from '@/components/inbox/ConversationDetails'
import { ChevronLeft, PanelRightClose } from 'lucide-react'

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

  const { data: conversationData } = useConversation(selectedId || '')

  const handleSelect = (conversationId: string) => {
    if (conversationId === selectedId) return
    navigate({ to: '/conversations/$id', params: { id: conversationId } })
  }

  return (
    <div className="h-[calc(100vh-5.5rem)] -mx-4 -my-5 flex bg-white">
      <div className="w-80 border-r border-gray-200 flex flex-col bg-gray-50">
        <ConversationList
          conversations={conversations}
          selectedId={selectedId}
          isLoading={isLoading}
          isError={isError}
          onSelect={handleSelect}
          onRetry={() => refetch()}
        />
      </div>

      <div className="flex-1 flex flex-col min-w-0 border-r border-gray-200 relative">
        {selectedId ? (
          <ConversationThread conversationId={selectedId} />
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center text-gray-400 p-8">
            <div className="h-16 w-16 rounded-full bg-gray-100 flex items-center justify-center mb-4">
              <svg
                className="h-8 w-8 text-gray-400"
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
        className={`bg-white border-l border-gray-200 overflow-y-auto transition-all duration-300 ${
          detailsCollapsed ? 'w-0 opacity-0 overflow-hidden' : 'w-80 opacity-100'
        }`}
      >
        <div className="relative min-w-[20rem]">
          <button
            type="button"
            onClick={() => setDetailsCollapsed(true)}
            className="absolute top-2 right-2 z-10 p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-md transition-colors"
            aria-label="Collapse conversation details"
            title="Collapse conversation details"
          >
            <PanelRightClose className="h-4 w-4" />
          </button>
          <ConversationDetails conversation={conversationData?.conversation || null} />
        </div>
      </div>

      {/* Floating expand button when collapsed */}
      {detailsCollapsed && (
        <button
          type="button"
          onClick={() => setDetailsCollapsed(false)}
          className="fixed right-3 top-[4.5rem] z-20 p-2 bg-white border border-gray-200 rounded-md shadow-sm text-gray-500 hover:text-gray-700 hover:bg-gray-50 transition-colors"
          aria-label="Expand conversation details"
          title="Expand conversation details"
        >
          <ChevronLeft className="h-4 w-4" />
        </button>
      )}
    </div>
  )
}

export default Inbox
