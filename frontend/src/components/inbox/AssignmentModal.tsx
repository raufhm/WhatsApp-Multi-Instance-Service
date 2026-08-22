import React, { useEffect, useMemo, useState } from 'react'
import { Loader2, Search, Sparkles, UserCircle2, Users } from 'lucide-react'
import Button from '@/components/ui/button'
import Modal from '@/components/ui/Modal'
import { useAssignConversation, useOperators } from '@/hooks/useInbox'

interface AssignmentTarget {
  id: string
  assignee?: string | null
  ticket_number?: number
  contact_name?: string
  contact_number?: string
  is_group?: boolean
}

interface AssignmentModalProps {
  isOpen: boolean
  conversation: AssignmentTarget | null
  onClose: () => void
}

const AssignmentModal: React.FC<AssignmentModalProps> = ({ isOpen, conversation, onClose }) => {
  const assign = useAssignConversation()
  const { data: operators = [], isLoading } = useOperators()
  const [selectedAssignee, setSelectedAssignee] = useState('')
  const [query, setQuery] = useState('')

  const activeOperators = useMemo(() => {
    const list = Array.isArray(operators) ? operators : []
    return list
      .filter((operator) => operator.is_active !== false)
      .sort((a, b) => {
        const aRole = (a.role || '').toLowerCase()
        const bRole = (b.role || '').toLowerCase()
        if (aRole !== bRole) return aRole.localeCompare(bRole)
        return a.name.localeCompare(b.name)
      })
  }, [operators])

  useEffect(() => {
    if (!isOpen) return
    setQuery('')
    setSelectedAssignee(conversation?.assignee || activeOperators[0]?.name || 'Admin')
  }, [isOpen, conversation?.assignee, activeOperators])

  const filteredOperators = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return activeOperators
    return activeOperators.filter((operator) => {
      const haystack = `${operator.name} ${operator.email || ''} ${operator.role}`.toLowerCase()
      return haystack.includes(q)
    })
  }, [activeOperators, query])

  const fallbackOnlyAdmin = activeOperators.length === 0
  const canSubmit = selectedAssignee.trim().length > 0 && !!conversation && !assign.isPending

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!conversation || !selectedAssignee.trim()) return
    assign.mutate(
      { id: conversation.id, assignee: selectedAssignee.trim() },
      {
        onSuccess: () => onClose(),
      }
    )
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Assign conversation"
      className="max-w-2xl"
    >
      <form onSubmit={handleSubmit} className="space-y-5">
        <div className="rounded-2xl border border-primary-100 bg-gradient-to-br from-primary-50 via-white to-sky-50 p-4">
          <div className="flex items-center gap-3">
            <div className="h-11 w-11 rounded-2xl bg-primary-600 text-white flex items-center justify-center shadow-lg shadow-primary-600/20">
              <Users className="h-5 w-5" />
            </div>
            <div className="min-w-0">
              <p className="text-sm font-semibold text-gray-900 truncate">
                {conversation?.contact_name || conversation?.contact_number || `Ticket #${conversation?.ticket_number || '-'}`}
              </p>
              <p className="text-xs text-gray-500">
                Pick an operator from the active team, or fall back to Admin if no operator is available.
              </p>
            </div>
          </div>
        </div>

        {fallbackOnlyAdmin ? (
          <div className="rounded-xl border border-amber-200 bg-amber-50 p-4">
            <div className="flex items-start gap-3">
              <Sparkles className="h-5 w-5 text-amber-600 mt-0.5" />
              <div>
                <p className="text-sm font-semibold text-amber-900">No active operators found</p>
                <p className="text-sm text-amber-800 mt-1">Only Admin is available for assignment right now.</p>
              </div>
            </div>
          </div>
        ) : (
          <div className="space-y-2">
            <label className="text-xs font-semibold uppercase tracking-wide text-gray-500">Search team</label>
            <div className="relative">
              <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
              <input
                className="form-control w-full pl-9"
                placeholder="Filter by name or role"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
              />
            </div>
          </div>
        )}

        <div className="space-y-2">
          <label className="text-xs font-semibold uppercase tracking-wide text-gray-500">Assign to</label>
          <select
            className="form-control w-full"
            value={selectedAssignee}
            onChange={(e) => setSelectedAssignee(e.target.value)}
            disabled={assign.isPending || (fallbackOnlyAdmin && !isLoading)}
          >
            {fallbackOnlyAdmin ? (
              <option value="Admin">Admin</option>
            ) : (
              <>
                {conversation?.assignee && !filteredOperators.some((op) => op.name === conversation.assignee) && (
                  <option value={conversation.assignee}>{conversation.assignee} (current)</option>
                )}
                {filteredOperators.map((operator) => (
                  <option key={operator.id} value={operator.name}>
                    {operator.name} {operator.role ? `• ${operator.role.toLowerCase()}` : ''}
                    {!operator.is_active ? ' (inactive)' : ''}
                  </option>
                ))}
                {filteredOperators.length === 0 && (
                  <option value={conversation?.assignee || 'Admin'}>{conversation?.assignee || 'Admin'}</option>
                )}
              </>
            )}
          </select>
          {isLoading && !fallbackOnlyAdmin && (
            <p className="text-xs text-gray-500 flex items-center gap-1.5">
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
              Loading operators
            </p>
          )}
        </div>

        <div className="rounded-xl border border-gray-200 bg-gray-50 p-3 text-xs text-gray-600">
          <div className="flex items-center gap-2">
            <UserCircle2 className="h-4 w-4 text-gray-400" />
            <span className="font-medium text-gray-700">
              Current assignee: {conversation?.assignee || 'Unassigned'}
            </span>
          </div>
        </div>

        <div className="flex items-center justify-end gap-2">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={!canSubmit}>
            {assign.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
            Confirm assignment
          </Button>
        </div>
      </form>
    </Modal>
  )
}

export default AssignmentModal
