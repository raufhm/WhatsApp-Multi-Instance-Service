import React, { useState } from 'react'
import { useCreateContactActivity } from '@/hooks/useInbox'
import Button from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Plus, Loader2, FileText, AlertCircle, Pencil, Trash2, Check, X } from 'lucide-react'

// ── Local entry types (manual-only) ──────────────────────────────────
interface SummaryEntry {
  id: string
  summary: string
  description: string
  type: 'note' | 'summary'
  timestamp: string
}

interface SummaryFeedProps {
  contactId: string
}

export function SummaryFeed({ contactId }: SummaryFeedProps) {
  const createActivity = useCreateContactActivity(contactId)
  const [entries, setEntries] = useState<SummaryEntry[]>([])
  const [summary, setSummary] = useState('')
  const [description, setDescription] = useState('')
  const [entryType, setEntryType] = useState<'note' | 'summary'>('note')
  const [formError, setFormError] = useState<string | null>(null)

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormError(null)
    if (!summary.trim()) {
      setFormError('Please enter a summary')
      return
    }

    const newEntry: SummaryEntry = {
      id: `local-${Date.now()}`,
      summary: summary.trim(),
      description: description.trim(),
      type: entryType,
      timestamp: new Date().toISOString(),
    }

    createActivity.mutate(
      {
        type: entryType === 'note' ? 'NOTE' : 'TICKET_SUMMARY',
        summary: summary.trim(),
        priority: 'NORMAL',
        due_at: null,
      },
      {
        onError: (err: any) =>
          setFormError(err?.response?.data?.error || err?.message || 'Failed to add entry'),
        onSuccess: () => {
          setEntries((prev) => [newEntry, ...prev])
          setSummary('')
          setDescription('')
          setEntryType('note')
        },
      }
    )
  }

  const handleUpdateEntry = (id: string, updates: Partial<SummaryEntry>) => {
    setEntries((prev) =>
      prev.map((e) => (e.id === id ? { ...e, ...updates, timestamp: new Date().toISOString() } : e))
    )
  }

  const handleDeleteEntry = (id: string) => {
    setEntries((prev) => prev.filter((e) => e.id !== id))
  }

  const renderForm = () => (
    <form onSubmit={handleAdd} className="space-y-2.5">
      <div className="flex items-center gap-2 mb-1">
        <Plus className="h-3.5 w-3.5 text-primary-600" />
        <span className="text-[11px] font-semibold uppercase tracking-wider text-gray-600">
          Add Summary Entry
        </span>
      </div>
      {formError && (
        <div className="flex items-start gap-2 p-2 rounded-lg bg-red-50 border border-red-200 text-[11px] text-red-700">
          <AlertCircle className="h-3.5 w-3.5 text-red-500 shrink-0 mt-0.5" />
          <span>{formError}</span>
        </div>
      )}
      <div>
        <Input
          type="text"
          value={summary}
          onChange={(e) => setSummary(e.target.value)}
          placeholder="Enter summary title..."
          className="w-full text-xs"
        />
      </div>
      <div>
        <Input
          type="text"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Additional details (optional)"
          className="w-full text-xs"
        />
      </div>
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-2 flex-1">
          <Label className="text-[11px] text-gray-500 shrink-0">Type:</Label>
          <select
            value={entryType}
            onChange={(e) => setEntryType(e.target.value as 'note' | 'summary')}
            className="rounded-lg border border-gray-200 bg-white px-2 py-1 text-[11px] text-gray-800 focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="note">Note</option>
            <option value="summary">Summary</option>
          </select>
        </div>
        <Button
          type="submit"
          variant="primary"
          size="sm"
          disabled={createActivity.isPending}
          className="shrink-0 gap-1 text-[11px]"
        >
          {createActivity.isPending ? (
            <>
              <Loader2 className="h-3 w-3 animate-spin" />
              <span>Adding...</span>
            </>
          ) : (
            <>
              <Plus className="h-3 w-3" />
              <span>Add {entryType === 'note' ? 'Note' : 'Summary'}</span>
            </>
          )}
        </Button>
      </div>
    </form>
  )

  // Empty state
  if (entries.length === 0) {
    return (
      <div className="p-4">
        <div className="text-center py-8 px-4 text-gray-400">
          <div className="w-10 h-10 rounded-full bg-gray-50 border border-gray-100 flex items-center justify-center mx-auto mb-3 text-gray-400">
            <FileText className="h-5 w-5 text-gray-400 stroke-[1.5]" />
          </div>
          <p className="text-xs font-medium text-gray-700">No summary entries yet</p>
          <p className="text-[11px] text-gray-400 mt-1 max-w-sm mx-auto">
            Manually add summary notes and updates for this contact using the form below.
          </p>
        </div>
        <div className="border-t border-gray-100 pt-4">{renderForm()}</div>
      </div>
    )
  }

  // Populated state
  return (
    <div className="p-4 space-y-3">
      {renderForm()}
      <div className="border-t border-gray-100 pt-3 space-y-0.5">
        {entries.map((entry) => (
          <SummaryEntryRow
            key={entry.id}
            entry={entry}
            onUpdate={handleUpdateEntry}
            onDelete={handleDeleteEntry}
          />
        ))}
      </div>
    </div>
  )
}

// ── SummaryEntryRow ──────────────────────────────────────────────────

function SummaryEntryRow({
  entry,
  onUpdate,
  onDelete,
}: {
  entry: SummaryEntry
  onUpdate: (id: string, updates: Partial<SummaryEntry>) => void
  onDelete: (id: string) => void
}) {
  const [isEditing, setIsEditing] = useState(false)
  const [editSummary, setEditSummary] = useState(entry.summary)
  const [editDescription, setEditDescription] = useState(entry.description)
  const [editType, setEditType] = useState(entry.type)

  const handleSave = () => {
    if (!editSummary.trim()) return
    onUpdate(entry.id, {
      summary: editSummary.trim(),
      description: editDescription.trim(),
      type: editType,
    })
    setIsEditing(false)
  }

  const handleCancel = () => {
    setEditSummary(entry.summary)
    setEditDescription(entry.description)
    setEditType(entry.type)
    setIsEditing(false)
  }

  const handleDelete = () => {
    onDelete(entry.id)
  }

  if (isEditing) {
    return (
      <div className="p-2.5 rounded-lg bg-gray-50 border border-gray-200 space-y-2">
        <Input
          type="text"
          value={editSummary}
          onChange={(e) => setEditSummary(e.target.value)}
          placeholder="Summary title..."
          className="w-full text-xs"
        />
        <Input
          type="text"
          value={editDescription}
          onChange={(e) => setEditDescription(e.target.value)}
          placeholder="Additional details (optional)"
          className="w-full text-xs"
        />
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Label className="text-[11px] text-gray-500 shrink-0">Type:</Label>
            <select
              value={editType}
              onChange={(e) => setEditType(e.target.value as 'note' | 'summary')}
              className="rounded-lg border border-gray-200 bg-white px-2 py-1 text-[11px] text-gray-800 focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="note">Note</option>
              <option value="summary">Summary</option>
            </select>
          </div>
          <div className="flex items-center gap-1.5">
            <button
              type="button"
              onClick={handleCancel}
              className="p-1 rounded hover:bg-gray-200 text-gray-500 hover:text-gray-700 transition-colors"
              title="Cancel"
            >
              <X className="h-3.5 w-3.5" />
            </button>
            <button
              type="button"
              onClick={handleSave}
              className="p-1 rounded hover:bg-primary-100 text-primary-600 hover:text-primary-700 transition-colors"
              title="Save"
            >
              <Check className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex gap-2.5 p-2.5 rounded-lg hover:bg-gray-50 transition-colors group">
      <div className="flex-shrink-0 mt-0.5">
        {entry.type === 'note' ? (
          <svg viewBox="0 0 24 24" fill="none" stroke="#60a5fa" strokeWidth={2} className="w-3.5 h-3.5">
            <path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z" />
          </svg>
        ) : (
          <svg viewBox="0 0 24 24" fill="none" stroke="#f97316" strokeWidth={2} className="w-3.5 h-3.5">
            <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" />
            <polyline points="14,2 14,8 20,8" />
          </svg>
        )}
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-xs font-medium text-gray-800 leading-snug">{entry.summary}</p>
        {entry.description && (
          <p className="text-[11px] text-gray-500 mt-0.5 leading-snug">{entry.description}</p>
        )}
        <p className="text-[10px] text-gray-400 mt-0.5">{formatRelativeTime(entry.timestamp)}</p>
      </div>
      <div className="flex items-start gap-1 shrink-0">
        <span
          className={`text-[10px] font-medium px-1.5 py-0.5 rounded-full ${
            entry.type === 'note'
              ? 'bg-blue-50 text-blue-600'
              : 'bg-orange-50 text-orange-600'
          }`}
        >
          {entry.type === 'note' ? 'Note' : 'Summary'}
        </span>
        <div className="opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-0.5 ml-1">
          <button
            type="button"
            onClick={() => setIsEditing(true)}
            className="p-1 rounded hover:bg-gray-200 text-gray-400 hover:text-gray-700 transition-colors"
            title="Edit"
          >
            <Pencil className="h-3 w-3" />
          </button>
          <button
            type="button"
            onClick={handleDelete}
            className="p-1 rounded hover:bg-red-50 text-gray-400 hover:text-red-600 transition-colors"
            title="Delete"
          >
            <Trash2 className="h-3 w-3" />
          </button>
        </div>
      </div>
    </div>
  )
}

// ── Helpers ──────────────────────────────────────────────────────────

function formatRelativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const minutes = Math.floor(diff / 60_000)
  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}d ago`
  return new Date(iso).toLocaleDateString()
}
