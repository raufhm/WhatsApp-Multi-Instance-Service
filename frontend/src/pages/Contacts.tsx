import React, { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { useContacts } from '@/hooks/useInbox'
import Card from '@/components/ui/card'
import Button from '@/components/ui/button'
import { Loader2, Users, AlertCircle } from 'lucide-react'

const Contacts: React.FC = () => {
  const { data: contacts = [], isLoading, isError, refetch } = useContacts()
  const [search, setSearch] = useState('')

  const filtered = contacts.filter((c) => {
    const q = search.toLowerCase()
    return !q || c.name.toLowerCase().includes(q) || c.number.includes(q)
  })

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">Contacts</h1>
          <p className="text-sm text-gray-600">Customer contact directory</p>
        </div>
        <Button variant="secondary" size="sm" onClick={() => refetch()}>
          Refresh
        </Button>
      </div>

      <Card className="p-4 mb-6">
        <input
          className="form-control"
          placeholder="Search by name or number..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </Card>

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
      ) : filtered.length === 0 ? (
        <div className="text-center py-12">
          <Users className="h-12 w-12 text-gray-300 mx-auto mb-3" />
          <p className="text-gray-600">No contacts found.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filtered.map((c) => (
            <Link
              key={c.id}
              to="/contacts/$id"
              params={{ id: c.id }}
              className="block"
              aria-label={`Open CRM record for ${c.name || c.number}`}
            >
              <Card className="p-4 hover:shadow-md transition-shadow">
                <div className="flex items-center gap-3">
                  <div className="h-10 w-10 rounded-full bg-primary-100 flex items-center justify-center">
                    <span className="font-medium text-primary-600">
                      {(c.name || '?').charAt(0).toUpperCase()}
                    </span>
                  </div>
                  <div>
                    <p className="font-medium text-gray-900">{c.name || 'Unnamed contact'}</p>
                    <p className="text-sm text-gray-500">{c.number}</p>
                  </div>
                </div>
                {c.email && (
                  <p className="mt-2 text-sm text-gray-500">{c.email}</p>
                )}
                {c.tags && c.tags.length > 0 && (
                  <div className="mt-3 flex flex-wrap gap-1">
                    {c.tags.map((tag) => (
                      <span key={tag} className="px-2 py-0.5 rounded-full text-xs bg-gray-100 text-gray-600">
                        {tag}
                      </span>
                    ))}
                  </div>
                )}
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}

export default Contacts