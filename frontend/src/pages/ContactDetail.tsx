import React, { useState } from 'react'
import { useParams, Link } from '@tanstack/react-router'
import {
  useContact,
  useContactActivities,
  useCreateContactActivity,
} from '@/hooks/useInbox'
import Button from '@/components/ui/button'
import Card from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Loader2,
  ArrowLeft,
  Users,
  AlertCircle,
  Mail,
  Phone,
  Tag,
  Plus,
  Activity as ActivityIcon,
} from 'lucide-react'
import type { Activity } from '@/types'

const priorityColors: Record<string, string> = {
  HIGH: 'bg-red-100 text-red-800',
  NORMAL: 'bg-yellow-100 text-yellow-800',
  LOW: 'bg-green-100 text-green-800',
}

const statusColors: Record<string, string> = {
  OPEN: 'bg-blue-100 text-blue-800',
  DONE: 'bg-green-100 text-green-800',
  CLOSED: 'bg-gray-100 text-gray-800',
}

const formatDate = (iso: string | null) =>
  iso ? new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' }) : '-'

export const ContactDetail: React.FC = () => {
  const { id } = useParams({ strict: false })
  const { data: contact, isLoading: contactLoading, isError: contactError, refetch: refetchContact } = useContact(id || '')
  const {
    data: activities = [],
    isLoading: activitiesLoading,
    isError: activitiesError,
    refetch: refetchActivities,
  } = useContactActivities(id || '')
  const createActivity = useCreateContactActivity(id || '')

  const [summary, setSummary] = useState('')
  const [priority, setPriority] = useState('NORMAL')
  const [dueAt, setDueAt] = useState('')
  const [formError, setFormError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormError(null)
    if (!summary.trim()) {
      setFormError('Please enter a follow-up summary')
      return
    }
    createActivity.mutate(
      {
        type: 'FOLLOW_UP',
        summary: summary.trim(),
        priority,
        due_at: dueAt ? new Date(dueAt).toISOString() : null,
      },
      {
        onError: (err: any) =>
          setFormError(err?.response?.data?.error || err?.message || 'Failed to create follow-up'),
        onSuccess: () => {
          setSummary('')
          setDueAt('')
          setPriority('NORMAL')
        },
      }
    )
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div>
        <Link
          to="/contacts"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-primary-600 hover:text-primary-700 mb-3"
        >
          <ArrowLeft className="h-4 w-4" /> Back to Contacts
        </Link>
        <h1 className="text-2xl font-bold text-gray-900">Contact Detail</h1>
      </div>

      {contactLoading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary-500" />
        </div>
      ) : contactError || !contact ? (
        <div className="text-center py-12">
          <AlertCircle className="h-12 w-12 text-red-400 mx-auto mb-3" />
          <p className="text-gray-600">Failed to load contact.</p>
          <Button variant="primary" size="sm" className="mt-4" onClick={() => refetchContact()}>
            Retry
          </Button>
        </div>
      ) : (
        <>
          <Card className="p-6">
            <div className="flex items-start gap-4">
              <div className="h-14 w-14 rounded-full bg-primary-100 flex items-center justify-center">
                <span className="text-xl font-bold text-primary-600">
                  {(contact.name || '?').charAt(0).toUpperCase()}
                </span>
              </div>
              <div className="flex-1">
                <h2 className="text-xl font-bold text-gray-900">{contact.name || 'Unnamed contact'}</h2>
                <div className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-gray-600">
                  <span className="inline-flex items-center gap-1">
                    <Phone className="h-3.5 w-3.5" /> {contact.number}
                  </span>
                  {contact.email && (
                    <span className="inline-flex items-center gap-1">
                      <Mail className="h-3.5 w-3.5" /> {contact.email}
                    </span>
                  )}
                  <span className="inline-flex items-center gap-1">
                    <ActivityIcon className="h-3.5 w-3.5" /> Joined {formatDate(contact.created_at)}
                  </span>
                </div>
              </div>
            </div>

            {contact.tags && contact.tags.length > 0 && (
              <div className="mt-4 flex flex-wrap items-center gap-2">
                <Tag className="h-4 w-4 text-gray-400" />
                {contact.tags.map((t) => (
                  <span
                    key={t}
                    className="px-2 py-0.5 rounded-full bg-gray-100 text-xs font-medium text-gray-700"
                  >
                    {t}
                  </span>
                ))}
              </div>
            )}
          </Card>

          <Card className="p-6">
            <h3 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
              <Users className="h-5 w-5 text-primary-600" /> CRM Timeline
            </h3>

            <div className="mt-4">
              {activitiesError ? (
                <div className="text-center py-8">
                  <AlertCircle className="h-10 w-10 text-red-400 mx-auto mb-2" />
                  <p className="text-gray-600">Failed to load activities.</p>
                  <Button
                    variant="secondary"
                    size="sm"
                    className="mt-3"
                    onClick={() => refetchActivities()}
                  >
                    Retry
                  </Button>
                </div>
              ) : activitiesLoading ? (
                <div className="flex justify-center py-8">
                  <Loader2 className="h-6 w-6 animate-spin text-primary-500" />
                </div>
              ) : activities.length === 0 ? (
                <div className="text-center py-8">
                  <p className="text-gray-500">No activities recorded yet.</p>
                </div>
              ) : (
                <ul className="space-y-3">
                  {activities.map((a: Activity) => (
                    <li key={a.id} className="rounded-lg border border-gray-100 bg-gray-50 p-4">
                      <div className="flex items-center justify-between gap-2 flex-wrap">
                        <div className="flex items-center gap-2">
                          <span className="text-xs font-semibold text-gray-700 uppercase">
                            {a.type}
                          </span>
                          <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${priorityColors[a.priority] || 'bg-gray-100 text-gray-700'}`}>
                            {a.priority}
                          </span>
                          <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${statusColors[a.status] || 'bg-gray-100 text-gray-700'}`}>
                            {a.status}
                          </span>
                        </div>
                        <span className="text-xs text-gray-500">{formatDate(a.created_at)}</span>
                      </div>
                      <p className="mt-2 text-sm text-gray-800">{a.summary}</p>
                      {a.next_action && (
                        <p className="mt-1 text-xs text-gray-500">Next action: {a.next_action}</p>
                      )}
                      {a.due_at && (
                        <p className="mt-1 text-xs text-amber-600">Due: {formatDate(a.due_at)}</p>
                      )}
                    </li>
                  ))}
                </ul>
              )}
            </div>

            <form onSubmit={handleSubmit} className="mt-6 space-y-4 border-t border-gray-100 pt-6">
              <h4 className="text-sm font-semibold text-gray-900">Add Follow-Up</h4>
              {formError && (
                <div className="flex items-start gap-2 p-3 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700">
                  <AlertCircle className="h-5 w-5 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>{formError}</span>
                </div>
              )}
              <div>
                <Label htmlFor="activitySummary">Summary</Label>
                <Input
                  id="activitySummary"
                  type="text"
                  value={summary}
                  onChange={(e) => setSummary(e.target.value)}
                  placeholder="e.g. Customer requested pricing sheet"
                  className="mt-1"
                />
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="activityPriority">Priority</Label>
                  <select
                    id="activityPriority"
                    value={priority}
                    onChange={(e) => setPriority(e.target.value)}
                    className="mt-1 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm"
                  >
                    <option value="HIGH">High</option>
                    <option value="NORMAL">Normal</option>
                    <option value="LOW">Low</option>
                  </select>
                </div>
                <div>
                  <Label htmlFor="activityDueAt">Due Date</Label>
                  <Input
                    id="activityDueAt"
                    type="datetime-local"
                    value={dueAt}
                    onChange={(e) => setDueAt(e.target.value)}
                    className="mt-1"
                  />
                </div>
              </div>
              <div className="flex justify-end">
                <Button type="submit" variant="primary" size="md" disabled={createActivity.isPending}>
                  {createActivity.isPending ? (
                    <span className="inline-flex items-center gap-2">
                      <Loader2 className="h-4 w-4 animate-spin" /> Adding...
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-2">
                      <Plus className="h-4 w-4" /> Add Follow-Up
                    </span>
                  )}
                </Button>
              </div>
            </form>
          </Card>
        </>
      )}
    </div>
  )
}

export default ContactDetail