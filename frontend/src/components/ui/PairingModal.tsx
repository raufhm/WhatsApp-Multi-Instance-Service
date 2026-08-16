import React, { useState, useEffect, useRef } from 'react'
import { QrCode, Smartphone, CheckCircle2, AlertCircle, Loader2, X, RefreshCw } from 'lucide-react'
import Button from './button'
import { Input } from './input'
import { Label } from './label'
import { pairingApi } from '@/lib/apiClient'
import type { PairingSnapshot } from '@/types'

interface PairingModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess?: (hostPhone: string) => void
  initialDisplayName?: string
}

export const PairingModal: React.FC<PairingModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
  initialDisplayName = '',
}) => {
  const [displayName, setDisplayName] = useState(initialDisplayName)
  const [pairingId, setPairingId] = useState<string | null>(null)
  const [snapshot, setSnapshot] = useState<PairingSnapshot | null>(null)
  const [step, setStep] = useState<'config' | 'pairing' | 'connected' | 'error'>('config')
  const [isLoading, setIsLoading] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  const pollingTimerRef = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => {
    if (isOpen) {
      setDisplayName(initialDisplayName)
      setPairingId(null)
      setSnapshot(null)
      setStep('config')
      setIsLoading(false)
      setErrorMessage(null)
    } else {
      stopPolling()
    }
    return () => {
      stopPolling()
    }
  }, [isOpen, initialDisplayName])

  const stopPolling = () => {
    if (pollingTimerRef.current) {
      clearInterval(pollingTimerRef.current)
      pollingTimerRef.current = null
    }
  }

  const startPolling = (id: string) => {
    stopPolling()
    const poll = async () => {
      try {
        const snap = await pairingApi.get(id)
        setSnapshot(snap)

        if (snap.status === 'connected') {
          stopPolling()
          setStep('connected')
          if (onSuccess && snap.host_phone) {
            onSuccess(snap.host_phone)
          }
        } else if (snap.status === 'failed') {
          stopPolling()
          setStep('error')
          setErrorMessage(snap.error || 'Pairing failed. Please try again.')
        } else if (snap.status === 'cancelled') {
          stopPolling()
          onClose()
        }
      } catch (err: any) {
        // If 404 or network error occurs repeatedly, treat as error
        if (err.response?.status === 404) {
          stopPolling()
          setStep('error')
          setErrorMessage('Pairing session expired or not found.')
        }
      }
    }

    // Initial check immediately, then every 2 seconds
    poll()
    pollingTimerRef.current = setInterval(poll, 2000)
  }

  const handleStartPairing = async (e?: React.FormEvent) => {
    if (e) e.preventDefault()
    setIsLoading(true)
    setErrorMessage(null)
    setStep('pairing')

    try {
      const data = await pairingApi.start(displayName.trim() || undefined)
      setPairingId(data.pairing_id)
      startPolling(data.pairing_id)
    } catch (err: any) {
      setStep('error')
      setErrorMessage(err.response?.data?.error || err.message || 'Failed to start pairing session')
    } finally {
      setIsLoading(false)
    }
  }

  const handleCancel = async () => {
    stopPolling()
    if (pairingId && step === 'pairing') {
      try {
        await pairingApi.cancel(pairingId)
      } catch {
        // ignore cancellation error on cleanup
      }
    }
    onClose()
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-full items-center justify-center p-4 text-center sm:p-0">
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" onClick={handleCancel} />

        <div className="relative transform overflow-hidden rounded-2xl bg-white text-left shadow-2xl transition-all sm:my-8 sm:w-full sm:max-w-lg p-6 sm:p-8">
          {/* Header */}
          <div className="flex items-center justify-between pb-4 border-b border-gray-100 mb-6">
            <div className="flex items-center gap-3">
              <div className="p-2.5 bg-green-100 text-green-700 rounded-xl">
                <Smartphone className="h-6 w-6" />
              </div>
              <div>
                <h3 className="text-lg font-bold text-gray-900">Link WhatsApp Device</h3>
                <p className="text-xs text-gray-500">Pair your WhatsApp account via QR code</p>
              </div>
            </div>
            <button
              onClick={handleCancel}
              className="text-gray-400 hover:text-gray-600 rounded-lg p-1 hover:bg-gray-100 transition-colors"
            >
              <X className="h-5 w-5" />
            </button>
          </div>

          {/* STEP: Config */}
          {step === 'config' && (
            <form onSubmit={handleStartPairing} className="space-y-5">
              <div>
                <Label htmlFor="accountDisplayName">Display Name (Optional)</Label>
                <Input
                  id="accountDisplayName"
                  type="text"
                  placeholder="e.g. Support Line, Sales Desk"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  className="mt-1"
                />
                <p className="text-xs text-gray-500 mt-1.5">
                  A recognizable label for this WhatsApp account in your dashboard.
                </p>
              </div>

              <div className="p-4 bg-gray-50 rounded-xl border border-gray-200 text-xs text-gray-600 space-y-2">
                <p className="font-semibold text-gray-800 flex items-center gap-1.5">
                  <QrCode className="h-4 w-4 text-green-600" />
                  Prerequisites for pairing:
                </p>
                <ul className="list-disc list-inside space-y-1 text-gray-600 pl-1">
                  <li>Have WhatsApp open on your mobile phone</li>
                  <li>Ensure your device has an active internet connection</li>
                  <li>Click below to generate a pairing QR code</li>
                </ul>
              </div>

              <div className="flex items-center justify-end gap-3 pt-4 border-t border-gray-100">
                <Button type="button" variant="ghost" size="md" onClick={handleCancel}>
                  Cancel
                </Button>
                <Button type="submit" variant="primary" size="md" disabled={isLoading} className="gap-2">
                  {isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <QrCode className="h-4 w-4" />}
                  <span>Generate QR Code</span>
                </Button>
              </div>
            </form>
          )}

          {/* STEP: Pairing */}
          {step === 'pairing' && (
            <div className="flex flex-col items-center space-y-5">
              {isLoading || (!snapshot?.qr_data_url && snapshot?.status !== 'expired') ? (
                <div className="py-12 flex flex-col items-center gap-3">
                  <Loader2 className="h-10 w-10 animate-spin text-green-600" />
                  <p className="text-sm font-medium text-gray-700">Initializing pairing session...</p>
                  <p className="text-xs text-gray-500">Establishing connection with WhatsApp servers</p>
                </div>
              ) : (
                <>
                  <div className="relative bg-white p-3 rounded-2xl border-2 border-gray-200 shadow-sm flex flex-col items-center">
                    {snapshot?.qr_data_url ? (
                      <img
                        src={snapshot.qr_data_url}
                        alt="WhatsApp Pairing QR Code"
                        className="w-56 h-56 object-contain rounded-lg"
                      />
                    ) : (
                      <div className="w-56 h-56 flex flex-col items-center justify-center bg-gray-50 rounded-lg text-gray-400 gap-2">
                        <RefreshCw className="h-8 w-8 animate-spin text-amber-500" />
                        <span className="text-xs text-gray-500">Refreshing QR code...</span>
                      </div>
                    )}
                  </div>

                  <div className="w-full bg-gray-50 rounded-xl p-4 border border-gray-200 text-left space-y-2">
                    <p className="text-xs font-bold text-gray-800">Scan with WhatsApp:</p>
                    <ol className="text-xs text-gray-600 space-y-1 list-decimal list-inside">
                      <li>Open WhatsApp on your phone</li>
                      <li>Go to <strong>Settings</strong> &gt; <strong>Linked Devices</strong></li>
                      <li>Tap <strong>Link a Device</strong> and point camera here</li>
                    </ol>
                  </div>

                  {snapshot?.status === 'expired' ? (
                    <div className="flex items-center gap-2 text-xs font-medium text-amber-700 bg-amber-50 px-3 py-1.5 rounded-full border border-amber-200">
                      <RefreshCw className="h-3.5 w-3.5 animate-spin text-amber-600" />
                      <span>QR code expired, generating fresh code...</span>
                    </div>
                  ) : (
                    <div className="flex items-center gap-2 text-xs font-medium text-green-700 bg-green-50 px-3 py-1.5 rounded-full border border-green-200">
                      <span className="h-2 w-2 rounded-full bg-green-500 animate-pulse" />
                      <span>Waiting for scan...</span>
                    </div>
                  )}

                  <div className="w-full pt-4 border-t border-gray-100 flex justify-center">
                    <Button type="button" variant="ghost" size="sm" onClick={handleCancel}>
                      Cancel Pairing
                    </Button>
                  </div>
                </>
              )}
            </div>
          )}

          {/* STEP: Connected */}
          {step === 'connected' && (
            <div className="py-6 flex flex-col items-center text-center space-y-4">
              <div className="p-4 bg-green-100 text-green-700 rounded-full">
                <CheckCircle2 className="h-12 w-12" />
              </div>
              <div>
                <h4 className="text-xl font-bold text-gray-900">Device Linked Successfully!</h4>
                <p className="text-sm text-gray-600 mt-1">
                  Connected host phone: <strong className="font-mono text-gray-900">{snapshot?.host_phone || 'Online'}</strong>
                </p>
                <p className="text-xs text-gray-400 mt-1">
                  Your WhatsApp instance is now active and ready to receive and send messages.
                </p>
              </div>

              <div className="pt-4 w-full">
                <Button type="button" variant="primary" size="md" className="w-full" onClick={onClose}>
                  Done
                </Button>
              </div>
            </div>
          )}

          {/* STEP: Error */}
          {step === 'error' && (
            <div className="py-4 flex flex-col items-center text-center space-y-4">
              <div className="p-4 bg-red-100 text-red-700 rounded-full">
                <AlertCircle className="h-10 w-10" />
              </div>
              <div>
                <h4 className="text-lg font-bold text-gray-900">Pairing Failed</h4>
                <p className="text-sm text-red-600 mt-1">{errorMessage || 'An error occurred during pairing.'}</p>
              </div>

              <div className="flex items-center gap-3 pt-4 w-full justify-center">
                <Button type="button" variant="ghost" size="md" onClick={handleCancel}>
                  Cancel
                </Button>
                <Button type="button" variant="primary" size="md" onClick={() => handleStartPairing()}>
                  Try Again
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default PairingModal
