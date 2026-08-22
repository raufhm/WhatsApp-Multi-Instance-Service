import React, { useMemo, useState } from 'react'
import type { ConversationSummary } from '@/types'
import { statusColors, formatListTime } from './constants'
import { Loader2, Search, AlertCircle, Inbox as InboxIcon, Users, Workflow } from 'lucide-react'

const STATUS_OPTIONS: { value: string; label: string }[] = [
  { value: 'ACTIVE', label: 'Active' },
  { value: 'OPEN', label: 'Open' },
  { value: 'BOT_ACTIVE', label: 'Bot active' },
  { value: 'WAITING', label: 'Waiting' },
  { value: 'HANDED_OFF', label: 'Handed off' },
  { value: 'CLOSED', label: 'Closed' },
  { value: 'ALL', label: 'All' },
]

export interface ChannelOption {
  value: string
  label: string
}

const TEAM_SIGNAL_PREFIX = '__TEAM_SIGNAL__:'

const formatPreview = (preview: string) => {
  if (!preview?.startsWith(TEAM_SIGNAL_PREFIX)) return preview
  try {
    const signal = JSON.parse(preview.slice(TEAM_SIGNAL_PREFIX.length)) as { kind?: string; emoji?: string }
    if (signal.kind === 'reaction' && signal.emoji) return `Team marked ${signal.emoji}`
  } catch {
    return 'Team marked a message'
  }
  return 'Team marked a message'
}

interface ConversationListProps {
  conversations: ConversationSummary[]
  channels: ChannelOption[]
  selectedId: string | null
  isLoading: boolean
  isError: boolean
  onSelect: (id: string) => void
  onRetry: () => void
}

const ConversationList: React.FC<ConversationListProps> = ({
  conversations,
  channels,
  selectedId,
  isLoading,
  isError,
  onSelect,
  onRetry,
}) => {
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState('ACTIVE')
  const [channelFilter, setChannelFilter] = useState('ALL')

  const filtered = useMemo(() => {
    const rawList = Array.isArray(conversations) ? conversations : []
    // A contact can have multiple tickets (including tickets on different
    // channels), but the inbox should present one consolidated row.
    const byContact = new Map<string, ConversationSummary>()
    for (const conversation of rawList) {
      const key = conversation.contact_id || conversation.id
      const current = byContact.get(key)
      if (!current || new Date(conversation.last_activity_at).getTime() > new Date(current.last_activity_at).getTime()) {
        byContact.set(key, conversation)
      }
    }
    const list = Array.from(byContact.values())
    const q = searchTerm.trim().toLowerCase()
    const selectedChannel = channelFilter === 'ALL' ? null : channelFilter

    return list.filter((c) => {
      if (statusFilter === 'ACTIVE' && c.status === 'CLOSED') return false
      if (statusFilter !== 'ACTIVE' && statusFilter !== 'ALL' && c.status !== statusFilter) return false
      if (selectedChannel && c.account_id !== selectedChannel) return false
      if (!q) return true
      return (
        String(c.ticket_number).includes(q) ||
        c.contact_name?.toLowerCase().includes(q) ||
        c.contact_number?.toLowerCase().includes(q)
      )
    })
  }, [channelFilter, conversations, searchTerm, statusFilter])

  return (
    <div className="h-full flex flex-col bg-slate-50/95 backdrop-blur overflow-hidden">
      <div className="p-4 border-b border-slate-200/90 space-y-3 bg-gradient-to-b from-slate-50 to-slate-100">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-sm font-semibold text-gray-900">Inbox</h2>
            <p className="text-[11px] text-gray-500 mt-0.5">Triaged conversations and active tickets</p>
          </div>
          <span className="text-xs font-medium text-gray-500 bg-gray-100 rounded-full px-2.5 py-1">{filtered.length}</span>
        </div>
        <div className="relative">
          <Search className="h-4 w-4 text-slate-500 absolute left-2.5 top-1/2 -translate-y-1/2" />
          <input
            className="form-control pl-8 py-2 text-sm w-full rounded-xl border-2 border-slate-300 bg-white text-slate-900 placeholder:text-slate-500 shadow-sm focus:border-primary-500 focus:bg-white"
            placeholder="Search name, number, or ticket..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            aria-label="Search conversations"
          />
        </div>
        <select
          className="form-control py-2 text-sm w-full rounded-xl border-2 border-slate-300 bg-white text-slate-900 shadow-sm"
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
        <select
          className="form-control py-2 text-sm w-full rounded-xl border-2 border-slate-300 bg-white text-slate-900 shadow-sm"
          value={channelFilter}
          onChange={(e) => setChannelFilter(e.target.value)}
          aria-label="Filter by channel"
        >
          <option value="ALL">All channels</option>
          {channels.map((channel) => (
            <option key={channel.value} value={channel.value}>
              {channel.label}
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
              const preview = formatPreview(c.last_message_preview)
              return (
                <li key={c.id}>
                  <button
                    type="button"
                    onClick={() => onSelect(c.id)}
                    className={`w-full text-left px-3 py-3 transition-colors ${
                      isSelected
                        ? 'bg-gradient-to-r from-primary-50 to-sky-50 border-l-2 border-primary-500 shadow-[inset_0_1px_0_rgba(255,255,255,0.8)]'
                        : 'border-l-2 border-transparent hover:bg-slate-50'
                    }`}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex items-center gap-2.5 min-w-0">
                        <div className={`h-9 w-9 rounded-2xl flex items-center justify-center shrink-0 ${isSelected ? 'bg-primary-600 text-white' : 'bg-primary-100 text-primary-600'}`}>
                          <span className="text-[13px] font-semibold">
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
                            {c.account_id && (
                              <span className="inline-flex items-center gap-0.5 text-[10px] font-medium text-sky-700 bg-sky-50 px-1.5 py-0.5 rounded-full">
                                <Workflow className="h-3 w-3" />
                                {channels.find((channel) => channel.value === c.account_id)?.label || c.account_id}
                              </span>
                            )}
                            <span className="text-[11px] text-gray-500 shrink-0">#{c.ticket_number}</span>
                          </p>
                          <p className="text-xs text-gray-500 truncate leading-tight">
                            {preview ? (
                              <>
                                {c.last_message_actor === 'OPERATOR' && (
                                  <span className="font-medium text-gray-700">You: </span>
                                )}
                                {c.last_message_actor === 'BOT' && (
                                  <span className="font-medium text-gray-700">Bot: </span>
                                )}
                                {preview}
                              </>
                            ) : (
                              c.contact_number ? c.contact_number : 'No messages yet'
                            )}
                          </p>
                        </div>
                      </div>
                      <div className="flex flex-col items-end gap-1 shrink-0">
                        <span className={`px-2 py-0.5 rounded-full text-[10px] font-medium ${statusColors[c.status] || 'bg-gray-100 text-gray-800'}`}>
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
