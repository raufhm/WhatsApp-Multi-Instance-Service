import React, { useMemo, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { useInbox, useActivities, useAcknowledgeActivity, useAssignConversation, type ConversationFilters } from '@/hooks/useInbox'
import type { Conversation } from '@/types'
import Button from '@/components/ui/button'
import Card from '@/components/ui/card'
import { DataTable, createColumnHelper } from '@/components/ui/DataTable'
import type { ColumnDef } from '@tanstack/react-table'
import { Inbox as InboxIcon, AlertCircle, CheckCircle2, Loader2 } from 'lucide-react'

const statusColors: Record<string, string> = {
  OPEN: 'bg-green-100 text-green-800',
  BOT_ACTIVE: 'bg-blue-100 text-blue-800',
  WAITING: 'bg-yellow-100 text-yellow-800',
  HANDED_OFF: 'bg-purple-100 text-purple-800',
  CLOSED: 'bg-gray-100 text-gray-800',
}

const formatTime = (iso: string) => {
  const d = new Date(iso)
  return d.toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const Inbox: React.FC = () => {
  const navigate = useNavigate()
  const [statusFilter, setStatusFilter] = useState('')
  const [assigneeFilter, setAssigneeFilter] = useState('')
  const [searchTerm, setSearchTerm] = useState('')

  const filters: ConversationFilters = useMemo(
    () => ({
      status: statusFilter || undefined,
      assignee: assigneeFilter || undefined,
      limit: 50,
    }),
    [statusFilter, assigneeFilter]
  )

  const { data: conversations = [], isLoading, isError, refetch } = useInbox(filters)
  const { data: activities = [] } = useActivities()
  const acknowledgeActivity = useAcknowledgeActivity()

  const assignMutation = useAssignConversation()

  const filtered = useMemo(() => {
    const list = Array.isArray(conversations) ? conversations : []
    if (!searchTerm) return list
    const q = searchTerm.toLowerCase()
    return list.filter(
      (c) =>
        String(c.ticket_number).includes(q) ||
        (c.assignee && c.assignee.toLowerCase().includes(q)) ||
        c.status.toLowerCase().includes(q)
    )
  }, [conversations, searchTerm])

  const pendingActivities = useMemo(() => {
    const list = Array.isArray(activities) ? activities : []
    return list.filter((a) => a.status === 'PENDING')
  }, [activities])

  const handleAssign = (c: Conversation) => {
    const name = window.prompt('Assign to operator:', c.assignee || '')
    if (name) {
      assignMutation.mutate({ id: c.id, assignee: name })
    }
  }

  const columns: ColumnDef<Conversation, any>[] = useMemo(() => {
    const helper = createColumnHelper<Conversation>()
    return [
      helper.accessor('ticket_number', {
        header: 'Ticket',
        cell: (info) => (
          <span className="font-medium text-primary-600">#{info.getValue()}</span>
        ),
      }),
      helper.accessor('status', {
        header: 'Status',
        cell: (info) => (
          <span className={`px-2 py-1 rounded-full text-xs font-medium ${statusColors[info.getValue()] || 'bg-gray-100 text-gray-800'}`}>
            {info.getValue()}
          </span>
        ),
      }),
      helper.accessor('assignee', {
        header: 'Assignee',
        cell: (info) => info.getValue() || <span className="text-gray-400">Unassigned</span>,
      }),
      helper.accessor('last_activity_at', {
        header: 'Last activity',
        cell: (info) => formatTime(info.getValue()),
      }),
    ]
  }, [])

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">Inbox</h1>
          <p className="text-sm text-gray-600">Manage customer conversations and follow-ups</p>
        </div>
        <Button variant="secondary" size="sm" onClick={() => refetch()}>
          Refresh
        </Button>
      </div>

      {/* Filters */}
      <Card className="mb-6 p-4">
        <div className="flex flex-wrap gap-3">
          <select
            className="form-control max-w-[180px]"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            aria-label="Filter by status"
          >
            <option value="">All statuses</option>
            <option value="OPEN">Open</option>
            <option value="BOT_ACTIVE">Bot active</option>
            <option value="WAITING">Waiting</option>
            <option value="HANDED_OFF">Handed off</option>
            <option value="CLOSED">Closed</option>
          </select>
          <input
            className="form-control max-w-[200px]"
            placeholder="Search ticket, assignee..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            aria-label="Search conversations"
          />
          <input
            className="form-control max-w-[180px]"
            placeholder="Filter by assignee"
            value={assigneeFilter}
            onChange={(e) => setAssigneeFilter(e.target.value)}
            aria-label="Filter by assignee"
          />
        </div>
      </Card>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary-500" />
        </div>
      ) : isError ? (
        <div className="text-center py-12">
          <AlertCircle className="h-12 w-12 text-red-400 mx-auto mb-3" />
          <p className="text-gray-600">Failed to load conversations.</p>
          <Button variant="primary" size="sm" className="mt-4" onClick={() => refetch()}>
            Retry
          </Button>
        </div>
      ) : filtered.length === 0 ? (
        <div className="text-center py-12">
          <InboxIcon className="h-12 w-12 text-gray-300 mx-auto mb-3" />
          <p className="text-gray-600">No tickets match your filters.</p>
          <Button
            variant="secondary"
            size="sm"
            className="mt-4"
            onClick={() => {
              setStatusFilter('')
              setSearchTerm('')
              setAssigneeFilter('')
            }}
          >
            Clear filters
          </Button>
        </div>
      ) : (
        <DataTable
          data={filtered}
          columns={columns}
          emptyMessage="No conversations found"
          enableGlobalFilter={false}
          renderActions={(c) => (
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                onClick={() => handleAssign(c)}
              >
                Assign
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => navigate({ to: '/conversations/$id', params: { id: c.id } })}
              >
                Open
              </Button>
            </div>
          )}
        />
      )}

      {/* Activity queue */}
      <div className="mt-8">
        <h2 className="text-lg font-semibold text-gray-900 mb-3">
          Activity queue{' '}
          {pendingActivities.length > 0 && (
            <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
              {pendingActivities.length} pending
            </span>
          )}
        </h2>
        {pendingActivities.length === 0 ? (
          <div className="bg-white rounded-lg border border-gray-200 p-6 text-center text-gray-500">
            No pending activities
          </div>
        ) : (
          <div className="space-y-2">
            {pendingActivities.map((a) => (
              <Card key={a.id} className="p-4 flex items-center justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <AlertCircle className="h-4 w-4 text-amber-500" />
                    <span className="text-sm font-medium text-gray-900">{a.summary}</span>
                  </div>
                  <p className="text-xs text-gray-500 mt-1">
                    {a.next_action || a.type} · {formatTime(a.created_at)}
                  </p>
                </div>
                <Button
                  variant="primary"
                  size="sm"
                  onClick={() => acknowledgeActivity.mutate(a.id)}
                  disabled={acknowledgeActivity.isPending}
                >
                  {acknowledgeActivity.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <CheckCircle2 className="h-4 w-4" />}
                  <span className="ml-1">Acknowledge</span>
                </Button>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export default Inbox