import React, { useState } from 'react'
import { useDashboardAccounts, useDisconnectAccount } from '@/hooks/useDashboard'
import Card from '@/components/ui/card'
import Button from '@/components/ui/button'
import { PairingModal } from '@/components/ui/PairingModal'
import { Loader2, Smartphone, AlertCircle, Plus, QrCode, ShieldAlert, Power, RefreshCw, Search } from 'lucide-react'

interface AccountWithHealth {
  id: string
  host_id: string
  display_name: string
  health: 'healthy' | 'disconnected' | 'unknown'
  is_connected: boolean
  queue_size: number
}

const Accounts: React.FC = () => {
  const { data: accounts = [], isLoading, isError, error, refetch } = useDashboardAccounts()
  const isForbidden = (error as { response?: { status?: number } } | null)?.response?.status === 403
  const [isPairingOpen, setIsPairingOpen] = useState(false)
  const [reconnectAccount, setReconnectAccount] = useState<AccountWithHealth | null>(null)
  const [actingId, setActingId] = useState<string | null>(null)
  const [searchTerm, setSearchTerm] = useState('')
  const disconnect = useDisconnectAccount()

  const filteredAccounts = (accounts as AccountWithHealth[]).filter((a) => {
    const q = searchTerm.trim().toLowerCase()
    if (!q) return true
    return (
      a.display_name?.toLowerCase().includes(q) ||
      a.host_id?.toLowerCase().includes(q)
    )
  })

  const handleDisconnect = async (id: string) => {
    setActingId(id)
    try {
      await disconnect.mutateAsync(id)
    } finally {
      setActingId(null)
    }
  }

  const openNewPairing = () => {
    setReconnectAccount(null)
    setIsPairingOpen(true)
  }

  const openReconnect = (a: AccountWithHealth) => {
    setReconnectAccount(a)
    setIsPairingOpen(true)
  }

  const closePairing = () => {
    setIsPairingOpen(false)
    setReconnectAccount(null)
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-xl font-semibold text-gray-900">WhatsApp Channels</h1>
          <p className="text-[13px] text-gray-600">Manage connected WhatsApp channels and their health</p>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="secondary" size="sm" onClick={() => refetch()}>
            Refresh
          </Button>
          <Button variant="primary" size="sm" onClick={openNewPairing} className="gap-1.5">
            <Plus className="h-4 w-4" />
            <span>New Channel</span>
          </Button>
        </div>
      </div>

      <Card className="overflow-hidden border border-gray-200/80">
        <div className="px-4 py-3 border-b border-gray-100 bg-gray-50/50 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div className="relative max-w-xs">
            <Search className="h-4 w-4 text-gray-400 absolute left-2.5 top-1/2 -translate-y-1/2" />
            <input
              className="form-control pl-8 py-1.5 text-sm w-full"
              placeholder="Search channels..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              aria-label="Search channels"
            />
          </div>
          <span className="text-xs text-gray-500">
            {filteredAccounts.length} channel{filteredAccounts.length !== 1 ? 's' : ''}
          </span>
        </div>

        {isLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-primary-500" />
          </div>
        ) : isError ? (
        isForbidden ? (
          <Card className="p-12 text-center max-w-lg mx-auto">
            <ShieldAlert className="h-12 w-12 text-amber-500 mx-auto mb-3" />
            <h3 className="text-lg font-semibold text-gray-900">You don't have permission to view channels</h3>
            <p className="text-sm text-gray-500 mt-1 max-w-sm mx-auto">
              Contact an administrator to request access to channel management.
            </p>
          </Card>
        ) : (
          <div className="text-center py-12 text-red-500">
            <AlertCircle className="h-12 w-12 mx-auto mb-3" />
            <p>Failed to load channels.</p>
            <Button variant="primary" size="sm" className="mt-4" onClick={() => refetch()}>
              Retry
            </Button>
          </div>
        )
      ) : filteredAccounts.length === 0 ? (
        <div className="p-12 text-center max-w-lg mx-auto">
          <div className="p-3 bg-green-50 text-green-600 rounded-2xl w-14 h-14 mx-auto mb-4 flex items-center justify-center">
            <QrCode className="h-7 w-7" />
          </div>
          <h3 className="text-lg font-semibold text-gray-900">No WhatsApp channels linked</h3>
          <p className="text-sm text-gray-500 mt-1 max-w-sm mx-auto">
            Pair a WhatsApp phone number to create a channel and start receiving and sending messages.
          </p>
          <div className="mt-6">
            <Button variant="primary" size="md" onClick={openNewPairing} className="gap-2">
              <Plus className="h-4 w-4" />
              <span>Link Your First Channel</span>
            </Button>
          </div>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-gray-50/80 text-gray-500 uppercase tracking-wider text-[11px] border-b border-gray-100">
              <tr>
                <th className="py-2.5 px-4 font-semibold">Channel</th>
                <th className="py-2.5 px-4 font-semibold">Status</th>
                <th className="py-2.5 px-4 font-semibold">Health</th>
                <th className="py-2.5 px-4 font-semibold text-right">Queue</th>
                <th className="py-2.5 px-4 font-semibold text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {(filteredAccounts as AccountWithHealth[]).map((a: AccountWithHealth) => (
                <tr key={a.id} className="hover:bg-gray-50/60 transition-colors">
                  <td className="py-3 px-4">
                    <div className="flex items-center gap-3">
                      <div className={`p-2 rounded-full shrink-0 ${a.is_connected ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600'}`}>
                        <Smartphone className="h-4 w-4" />
                      </div>
                      <div>
                        <p className="font-medium text-gray-900">{a.display_name || a.host_id}</p>
                        <p className="text-[11px] text-gray-500 font-mono">{a.host_id}</p>
                      </div>
                    </div>
                  </td>
                  <td className="py-3 px-4">
                    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium ${
                      a.is_connected ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                    }`}>
                      {a.is_connected ? (
                        <>
                          <span className="w-1.5 h-1.5 rounded-full bg-green-500" />
                          Connected
                        </>
                      ) : (
                        <>
                          <span className="w-1.5 h-1.5 rounded-full bg-red-500" />
                          Disconnected
                        </>
                      )}
                    </span>
                  </td>
                  <td className="py-3 px-4">
                    <span className={`px-2 py-0.5 rounded-full text-[11px] font-medium ${
                      a.health === 'healthy' ? 'bg-green-100 text-green-800' :
                      a.health === 'disconnected' ? 'bg-red-100 text-red-800' :
                      'bg-gray-100 text-gray-800'
                    }`}>
                      {a.health}
                    </span>
                  </td>
                  <td className="py-3 px-4 text-right font-mono text-gray-700">
                    {a.queue_size}
                  </td>
                  <td className="py-3 px-4 text-right">
                    {a.is_connected ? (
                      <Button
                        variant="secondary"
                        size="sm"
                        className="gap-1.5"
                        disabled={actingId === a.id}
                        onClick={() => handleDisconnect(a.id)}
                      >
                        {actingId === a.id ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Power className="h-3.5 w-3.5" />}
                        Disconnect
                      </Button>
                    ) : (
                      <Button
                        variant="primary"
                        size="sm"
                        className="gap-1.5"
                        disabled={actingId === a.id}
                        onClick={() => openReconnect(a)}
                      >
                        {actingId === a.id ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <RefreshCw className="h-3.5 w-3.5" />}
                        Reconnect
                      </Button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      </Card>

      <PairingModal
        isOpen={isPairingOpen}
        onClose={closePairing}
        onSuccess={() => {
          closePairing()
          refetch()
        }}
        initialDisplayName={reconnectAccount?.display_name || reconnectAccount?.host_id || ''}
      />
    </div>
  )
}

export default Accounts
