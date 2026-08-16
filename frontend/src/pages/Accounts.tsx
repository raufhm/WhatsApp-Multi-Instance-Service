import React, { useState } from 'react'
import { useDashboardAccounts } from '@/hooks/useDashboard'
import Card from '@/components/ui/card'
import Button from '@/components/ui/button'
import { PairingModal } from '@/components/ui/PairingModal'
import { Loader2, Smartphone, AlertCircle, CheckCircle2, Signal, Plus, QrCode, ShieldAlert } from 'lucide-react'

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

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">WhatsApp Accounts</h1>
          <p className="text-sm text-gray-600">Monitor paired accounts and connection health</p>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="secondary" size="sm" onClick={() => refetch()}>
            Refresh
          </Button>
          <Button variant="primary" size="sm" onClick={() => setIsPairingOpen(true)} className="gap-1.5">
            <Plus className="h-4 w-4" />
            <span>Link Device</span>
          </Button>
        </div>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary-500" />
        </div>
      ) : isError ? (
        isForbidden ? (
          <Card className="p-12 text-center max-w-lg mx-auto">
            <ShieldAlert className="h-12 w-12 text-amber-500 mx-auto mb-3" />
            <h3 className="text-lg font-semibold text-gray-900">You don't have permission to view accounts</h3>
            <p className="text-sm text-gray-500 mt-1 max-w-sm mx-auto">
              Contact an administrator to request access to account management.
            </p>
          </Card>
        ) : (
          <div className="text-center py-12 text-red-500">
            <AlertCircle className="h-12 w-12 mx-auto mb-3" />
            <p>Failed to load accounts.</p>
            <Button variant="primary" size="sm" className="mt-4" onClick={() => refetch()}>
              Retry
            </Button>
          </div>
        )
      ) : accounts.length === 0 ? (
        <Card className="p-12 text-center max-w-lg mx-auto">
          <div className="p-3 bg-green-50 text-green-600 rounded-2xl w-14 h-14 mx-auto mb-4 flex items-center justify-center">
            <QrCode className="h-7 w-7" />
          </div>
          <h3 className="text-lg font-semibold text-gray-900">No WhatsApp accounts linked</h3>
          <p className="text-sm text-gray-500 mt-1 max-w-sm mx-auto">
            Pair your WhatsApp phone number to start receiving and sending messages through the platform.
          </p>
          <div className="mt-6">
            <Button variant="primary" size="md" onClick={() => setIsPairingOpen(true)} className="gap-2">
              <Plus className="h-4 w-4" />
              <span>Link Your First Device</span>
            </Button>
          </div>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {(accounts as AccountWithHealth[]).map((a) => (
            <Card key={a.id} className="p-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className={`p-2 rounded-full ${a.is_connected ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600'}`}>
                    <Smartphone className="h-5 w-5" />
                  </div>
                  <div>
                    <p className="font-medium text-gray-900">{a.display_name || a.host_id}</p>
                    <p className="text-xs text-gray-500">{a.host_id}</p>
                  </div>
                </div>
                {a.is_connected ? <CheckCircle2 className="h-5 w-5 text-green-500" /> : <Signal className="h-5 w-5 text-red-500" />}
              </div>
              <div className="mt-4 flex items-center justify-between text-sm">
                <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                  a.is_connected ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                }`}>
                  {a.health}
                </span>
                <span className="text-gray-500">Queue: {a.queue_size}</span>
              </div>
            </Card>
          ))}
        </div>
      )}

      <PairingModal
        isOpen={isPairingOpen}
        onClose={() => setIsPairingOpen(false)}
        onSuccess={() => {
          refetch()
        }}
      />
    </div>
  )
}

export default Accounts
