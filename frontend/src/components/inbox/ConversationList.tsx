import React, { useMemo, useState } from 'react'
import type { ConversationSummary } from '@/types'
import { statusColors, formatListTime } from './constants'
import { Loader2, Search, AlertCircle, Inbox as InboxIcon, Users } from 'lucide-react'

const STATUS_OPTIONS: { value: string; label: string }[] = [
  { value: 'ACTIVE', label: 'Active' },
  { value: 'OPEN', label: 'Open' },
  { value: 'BOT_ACTIVE', label: 'Bot active' },
  { value: 'WAITING', label: 'Waiting' },
  { value: 'HANDED_OFF', label: 'Handed off' },
  { value: 'CLOSED', label: 'Closed' },
  { value: 'ALL', label: 'All' },
]

interface ConversationListProps {
  conversations: ConversationSummary[]
  selectedId: string | null
  isLoading: boolean
  isError: boolean
  onSelect: (id: string) => void
  onRetry: () => void
}

const ConversationList: React.FC<ConversationListProps> = ({
  conversations,
  selectedId,
  isLoading,
  isError,
  onSelect,
  onRetry,
}) => {
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState('ACTIVE')

  const filtered = useMemo(() => {
    const list = Array.isArray(conversations) ? conversations : []
    const q = searchTerm.trim().toLowerCase()

    const seenContacts = new Set<string>()
    const uniqueList: ConversationSummary[] = []
    for (const c of list) {
      const contactKey = c.contact_id || c.contact_number || c.id
      if (!seenContacts.has(contactKey)) {
        seenContacts.add(contactKey)
        uniqueList.push(c)
      }
    }

    return uniqueList.filter((c) => {
      if (statusFilter === 'ACTIVE' && c.status === 'CLOSED') return false
      if (statusFilter !== 'ACTIVE' && statusFilter !== 'ALL' && c.status !== statusFilter) return false
      if (!q) return true
      return (
        String(c.ticket_number).includes(q) ||
        c.contact_name?.toLowerCase().includes(q) ||
        c.contact_number?.toLowerCase().includes(q)
      )
    })
  }, [conversations, searchTerm, statusFilter])

  return (
    <div className="h-full flex flex-col bg-white lg:rounded-lg overflow-hidden">
      <div className="p-3 border-b border-gray-200 space-y-2">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold text-gray-900">Inbox</h2>
          <span className="text-xs text-gray-500">{filtered.length}</span>
        </div>
        <div className="relative">
          <Search className="h-4 w-4 text-gray-400 absolute left-2.5 top-1/2 -translate-y-1/2" />
          <input
            className="form-control pl-8 py-1.5 text-sm w-full"
            placeholder="Search name, number, or ticket..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            aria-label="Search conversations"
          />
        </div>
        <select
          className="form-control py-1.5 text-sm w-full"
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          aria-label="Filter by status"
        >
          {STATUS_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
      </div>

      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <div className="flex justify-center py-10">
            <Loader2 className="h-6 w-6 animate-spin text-primary-500" />
          </div>
        ) : isError ? (
          <div className="text-center py-10 px-4">
            <AlertCircle className="h-10 w-10 text-red-400 mx-auto mb-2" />
            <p className="text-sm text-gray-600">Failed to load conversations.</p>
            <button
              type="button"
              onClick={onRetry}
              className="mt-3 text-sm font-medium text-primary-600 hover:text-primary-700"
            >
              Retry
            </button>
          </div>
        ) : filtered.length === 0 ? (
          <div className="text-center py-10 px-4">
            <InboxIcon className="h-10 w-10 text-gray-300 mx-auto mb-2" />
            <p className="text-sm text-gray-500">
              {searchTerm || statusFilter !== 'ACTIVE' ? 'No conversations match your filters.' : 'No active conversations yet.'}
            </p>
          </div>
        ) : (
          <ul className="divide-y divide-gray-100">
            {filtered.map((c) => {
              const name = c.contact_name || c.contact_number || 'Unknown user'
              const isSelected = c.id === selectedId
              return (
                <li key={c.id}>
                  <button
                    type="button"
                    onClick={() => onSelect(c.id)}
                    className={`w-full text-left px-3 py-2.5 transition-colors ${
                      isSelected ? 'bg-primary-50 border-l-2 border-primary-500' : 'border-l-2 border-transparent hover:bg-gray-50'
                    }`}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex items-center gap-2.5 min-w-0">
                        <div className="h-8 w-8 rounded-full bg-primary-100 flex items-center justify-center shrink-0">
                          <span className="text-[13px] font-semibold text-primary-600">
                            {name.charAt(0).toUpperCase()}
                          </span>
                        </div>
                        <div className="min-w-0">
                          <p className="text-sm font-medium text-gray-900 truncate flex items-center gap-1.5">
                            {name}
                            {c.is_group && (
                              <span className="inline-flex items-center gap-0.5 text-[10px] font-medium text-amber-600 bg-amber-50 px-1.5 py-0.5 rounded-full">
                                <Users className="h-3 w-3" />
                                Group
                              </span>
                            )}
                            <span className="text-[11px] text-gray-500 shrink-0">#{c.ticket_number}</span>
                          </p>
                          <p className="text-xs text-gray-500 truncate leading-tight">
                            {c.last_message_preview ? (
                              <>
                                {c.last_message_actor === 'OPERATOR' && (
                                  <span className="font-medium text-gray-700">You: </span>
                                )}
                                {c.last_message_actor === 'BOT' && (
                                  <span className="font-medium text-gray-700">Bot: </span>
                                )}
                                {c.last_message_preview}
                              </>
                            ) : (
                              c.contact_number ? c.contact_number : 'No messages yet'
                            )}
                          </p>
                        </div>
                      </div>
                      <div className="flex flex-col items-end gap-0.5 shrink-0">
                        <span className={`px-1.5 py-0.5 rounded-full text-[10px] font-medium ${statusColors[c.status] || 'bg-gray-100 text-gray-800'}`}>
                          {c.status.replace('_', ' ')}
                        </span>
                        <span className="text-[10px] text-gray-400">{formatListTime(c.last_activity_at)}</span>
                      </div>
                    </div>
                  </button>
                </li>
              )
            })}
          </ul>
        )}
      </div>
    </div>
  )
}

export default ConversationList