import React, { useEffect, useMemo, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { useContacts } from '@/hooks/useInbox'
import Card from '@/components/ui/card'
import Button from '@/components/ui/button'
import { Loader2, Users, AlertCircle, ChevronLeft, ChevronRight, Search, Users2, ArrowRight } from 'lucide-react'

const PAGE_SIZE = 20
const CONTACT_TYPE_OPTIONS = [
  { value: 'ALL', label: 'All contacts' },
  { value: 'GROUP', label: 'Group chats' },
  { value: 'PERSONAL', label: 'Personal chats' },
]

const Contacts: React.FC = () => {
  const [search, setSearch] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [offset, setOffset] = useState(0)
  const [contactType, setContactType] = useState('ALL')

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search)
      setOffset(0)
    }, 300)
    return () => clearTimeout(timer)
  }, [search])

  useEffect(() => {
    setOffset(0)
  }, [contactType])

  const { data, isLoading, isError, refetch } = useContacts({
    limit: PAGE_SIZE,
    offset,
    q: debouncedSearch,
  })

  const contacts = data?.items ?? []
  const total = data?.total ?? 0
  const hasNext = offset + PAGE_SIZE < total
  const hasPrev = offset > 0
  const sortedContacts = useMemo(() => {
    const list = [...contacts]
    return list
      .filter((c) => {
        if (contactType === 'GROUP') return !!c.is_group
        if (contactType === 'PERSONAL') return !c.is_group
        return true
      })
      .sort((a, b) => {
        const groupDiff = Number(Boolean(b.is_group)) - Number(Boolean(a.is_group))
        if (groupDiff !== 0) return groupDiff
        return (a.name || a.number || '').localeCompare(b.name || b.number || '')
      })
  }, [contactType, contacts])

  return (
    <div>
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-xl font-semibold text-gray-900">Contacts</h1>
          <p className="text-[13px] text-gray-600">Customer contact directory</p>
        </div>
        <Button variant="secondary" size="sm" onClick={() => refetch()}>
          Refresh
        </Button>
      </div>

      <Card className="p-4 mb-5">
        <div className="space-y-3">
          <div className="relative">
            <Search className="h-4 w-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              className="form-control pl-9 w-full"
              placeholder="Search by name or number..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
          <select
            className="form-control w-full"
            value={contactType}
            onChange={(e) => setContactType(e.target.value)}
            aria-label="Filter contact type"
          >
            {CONTACT_TYPE_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>
      </Card>

      <Card className="overflow-hidden">
        {isLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-primary-500" />
          </div>
        ) : isError ? (
          <div className="text-center py-12">
            <AlertCircle className="h-12 w-12 text-red-400 mx-auto mb-3" />
            <p className="text-gray-600">Failed to load contacts.</p>
            <Button variant="primary" size="sm" className="mt-4" onClick={() => refetch()}>
              Retry
            </Button>
          </div>
        ) : sortedContacts.length === 0 ? (
          <div className="text-center py-12">
            <Users className="h-12 w-12 text-gray-300 mx-auto mb-3" />
            <p className="text-gray-600">
              {contactType === 'ALL' ? 'No contacts found.' : 'No contacts match this type filter.'}
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 text-sm">
              <thead className="bg-gray-50 text-xs font-semibold text-gray-500 uppercase tracking-wider">
                <tr>
                  <th scope="col" className="px-4 py-3 text-left">Contact</th>
                  <th scope="col" className="px-4 py-3 text-left">Number / ID</th>
                  <th scope="col" className="px-4 py-3 text-left">Type</th>
                  <th scope="col" className="px-4 py-3 text-left">Deal Stage</th>
                  <th scope="col" className="px-4 py-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {sortedContacts.map((c) => (
                  <tr key={c.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 whitespace-nowrap">
                      <div className="flex items-center gap-3">
                        <div className="h-9 w-9 rounded-lg bg-primary-100 text-primary-700 font-bold flex items-center justify-center text-xs">
                          {(c.name || c.number || '?').charAt(0).toUpperCase()}
                        </div>
                        <div>
                          <p className="text-gray-900">{c.name || c.number}</p>
                          {c.email && <p className="text-xs text-gray-500">{c.email}</p>}
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-gray-600 font-mono text-xs">
                      {c.number}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap">
                      {c.is_group ? (
                        <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-100 text-amber-800">
                          <Users2 className="h-3 w-3" />
                          Group
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                          <Users className="h-3 w-3" />
                          Personal
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap">
                      {c.deal_stage ? (
                        <span
                          className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium text-white"
                          style={{ backgroundColor: c.deal_stage.color }}
                        >
                          {c.deal_stage.label}
                        </span>
                      ) : (
                        <span className="text-xs text-gray-400">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-right">
                      <Link
                        to="/contacts/$id"
                        params={{ id: c.id }}
                        className="inline-flex items-center gap-1 text-xs font-medium text-primary-600 hover:text-primary-700 hover:bg-primary-50 px-2 py-1 rounded-md transition-colors"
                        aria-label={`Open CRM record for ${c.name || c.number}`}
                      >
                        View
                        <ArrowRight className="h-3.5 w-3.5" />
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {!isLoading && !isError && sortedContacts.length > 0 && (
        <div className="flex items-center justify-between mt-4">
          <p className="text-sm text-gray-500">
            Showing {Math.min(offset + 1, total)}-{Math.min(offset + PAGE_SIZE, total)} of {total}
          </p>
          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setOffset((o) => Math.max(0, o - PAGE_SIZE))}
              disabled={!hasPrev}
            >
              <ChevronLeft className="h-4 w-4" />
              Previous
            </Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setOffset((o) => o + PAGE_SIZE)}
              disabled={!hasNext}
            >
              Next
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

export default Contacts
