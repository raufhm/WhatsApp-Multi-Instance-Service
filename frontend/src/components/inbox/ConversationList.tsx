import React, { useMemo, useState } from 'react'
import type { ConversationSummary } from '@/types'
import { statusColors, formatListTime } from './constants'
import { Loader2, Search, AlertCircle, Inbox as InboxIcon, Users, Workflow, SlidersHorizontal } from 'lucide-react'

const STATUS_OPTIONS: { value: string; label: string }[] = [
  { value: 'ALL', label: 'All' },
  { value: 'CLOSED', label: 'Close' },
  { value: 'OPEN', label: 'Open' },
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
  const [statusFilter, setStatusFilter] = useState('OPEN')
  const [channelFilter, setChannelFilter] = useState('ALL')
  const [filtersOpen, setFiltersOpen] = useState(false)

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
      if (statusFilter === 'OPEN' && c.status === 'CLOSED') return false
      if (statusFilter === 'CLOSED' && c.status !== 'CLOSED') return false
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
    <div className="h-full flex flex-col bg-white/95 backdrop-blur overflow-hidden">
      <div className="px-3 py-4 space-y-3 bg-white">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-[1.35rem] font-bold tracking-normal text-gray-950">Your Inbox ({filtered.length})</h2>
            <p className="mt-1 text-[11px] text-gray-500">Triaged conversations and active tickets</p>
          </div>
          <span className="inline-flex h-8 min-w-8 items-center justify-center rounded-full bg-[#f1ede9] px-2 text-xs font-semibold text-gray-600">{filtered.length}</span>
        </div>
        <div className="grid grid-cols-[1fr_2.75rem] gap-2">
          <div className="relative">
          <Search className="h-4 w-4 text-gray-500 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            className="h-11 w-full rounded-xl border border-gray-200 bg-white/90 pl-9 pr-3 text-sm text-gray-950 placeholder:text-gray-500 shadow-sm outline-none transition-colors focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20"
            placeholder="Search name, number, or ticket..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            aria-label="Search conversations"
          />
          </div>
          <div className="relative group">
            <button
              type="button"
              className="inline-flex h-11 w-full items-center justify-center rounded-xl border border-gray-200 bg-white/90 text-gray-800 shadow-sm transition-colors hover:bg-[#f8f4ef]"
              aria-label={filtersOpen ? 'Hide channel filter' : 'Show channel filter'}
              aria-expanded={filtersOpen}
              title="Filter by channel"
              onClick={() => setFiltersOpen((open) => !open)}
            >
              <SlidersHorizontal className="h-5 w-5" />
            </button>
            <div
              className={`absolute right-0 top-12 z-30 w-48 origin-top-right rounded-xl border border-gray-200 bg-white p-2 shadow-lg transition-all duration-200 ease-out ${
                filtersOpen
                  ? 'translate-y-0 scale-100 opacity-100 pointer-events-auto'
                  : 'translate-y-1 scale-95 opacity-0 pointer-events-none group-hover:translate-y-0 group-hover:scale-100 group-hover:opacity-100 group-hover:pointer-events-auto group-focus-within:translate-y-0 group-focus-within:scale-100 group-focus-within:opacity-100 group-focus-within:pointer-events-auto'
              }`}
            >
              <select
                className="h-9 w-full rounded-lg border border-gray-200 bg-white px-2 text-xs text-gray-800 shadow-sm outline-none focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20"
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
          </div>
        </div>
        <div className="rounded-xl bg-[#f1ede9] p-1" role="tablist" aria-label="Inbox filters">
          <div className="grid grid-cols-3 gap-1 text-xs text-gray-600">
            {STATUS_OPTIONS.map((option) => (
              <button
                key={option.value}
                type="button"
                className={`h-8 rounded-lg transition-colors ${
                  statusFilter === option.value
                    ? 'bg-white font-semibold text-gray-950 shadow-sm'
                    : 'hover:bg-white/60'
                }`}
                onClick={() => setStatusFilter(option.value)}
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>

      </div>

      <div className="flex-1 overflow-y-auto px-3 pb-4">
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
              {searchTerm || statusFilter !== 'OPEN' ? 'No conversations match your filters.' : 'No open conversations yet.'}
            </p>
          </div>
        ) : (
          <ul className="space-y-1.5">
            {filtered.map((c) => {
              const name = c.contact_name || c.contact_number || 'Unknown user'
              const isSelected = c.id === selectedId
              const preview = formatPreview(c.last_message_preview)
              return (
                <li key={c.id}>
                  <button
                    type="button"
                    onClick={() => onSelect(c.id)}
                    className={`w-full rounded-xl text-left px-3 py-3 transition-colors ${
                      isSelected
                        ? 'bg-[#f2efec] shadow-sm'
                        : 'hover:bg-[#f8f4ef]'
                    }`}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex items-center gap-2.5 min-w-0">
                        <div className={`h-10 w-10 rounded-full flex items-center justify-center shrink-0 shadow-sm ${isSelected ? 'bg-gradient-to-br from-orange-400 to-orange-600 text-white' : 'bg-gradient-to-br from-cyan-100 to-orange-100 text-gray-700'}`}>
                          <span className="text-[13px] font-semibold">
                            {name.charAt(0).toUpperCase()}
                          </span>
                        </div>
                        <div className="min-w-0">
                          <p className="text-sm font-medium text-gray-900 truncate flex items-center gap-1.5">
                            {name}
                            {c.is_group && (
                              <span className="inline-flex items-center gap-0.5 text-[10px] font-medium text-amber-700 bg-amber-50 px-1.5 py-0.5 rounded-full">
                                <Users className="h-3 w-3" />
                                Group
                              </span>
                            )}
                            {c.account_id && (
                              <span className="inline-flex items-center gap-0.5 text-[10px] font-medium text-cyan-700 bg-cyan-50 px-1.5 py-0.5 rounded-full">
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
