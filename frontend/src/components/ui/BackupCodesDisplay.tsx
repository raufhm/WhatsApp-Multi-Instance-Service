import React, { useState } from 'react'
import { Copy, Check, Download, AlertTriangle, ShieldAlert } from 'lucide-react'
import Button from './button'

interface BackupCodesDisplayProps {
  codes: string[]
  title?: string
  description?: string
  onAcknowledge?: () => void
  acknowledgeLabel?: string
  className?: string
  showAcknowledgeCheckbox?: boolean
}

export const BackupCodesDisplay: React.FC<BackupCodesDisplayProps> = ({
  codes = [],
  title = 'Your Recovery Backup Codes',
  description = 'Store these 10 one-time recovery codes safely in a password manager or secure notes. If you ever lose access to your authenticator app, each code can be used once to log in.',
  onAcknowledge,
  acknowledgeLabel = 'Continue to Dashboard',
  className = '',
  showAcknowledgeCheckbox = false,
}) => {
  const [copied, setCopied] = useState(false)
  const [hasAcknowledged, setHasAcknowledged] = useState(false)

  const handleCopyAll = async () => {
    const text = codes.join('\n')
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      const el = document.createElement('textarea')
      el.value = text
      document.body.appendChild(el)
      el.select()
      document.execCommand('copy')
      document.body.removeChild(el)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const handleDownloadTxt = () => {
    const dateStr = new Date().toISOString().split('T')[0]
    const content = `WhatsApp Multi-Instance Service - Account Backup Codes
Created on: ${new Date().toLocaleString()}
==================================================

WARNING: Keep these codes in a safe, private place.
Each backup code is single-use and can be used to
log into your account if you lose your authenticator.

Backup Codes:
${codes.map((c, i) => `${(i + 1).toString().padStart(2, ' ')}. ${c}`).join('\n')}

==================================================
`
    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `whatsapp-backup-codes-${dateStr}.txt`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  }

  return (
    <div className={`space-y-6 ${className}`}>
      <div>
        <h3 className="text-xl font-bold text-gray-900 flex items-center gap-2">
          <ShieldAlert className="h-6 w-6 text-amber-500" />
          {title}
        </h3>
        <p className="text-sm text-gray-600 mt-1">{description}</p>
      </div>

      <div className="p-4 rounded-lg bg-amber-50 border border-amber-200 flex items-start gap-3 text-sm text-amber-800">
        <AlertTriangle className="h-5 w-5 text-amber-600 flex-shrink-0 mt-0.5" />
        <div>
          <p className="font-semibold text-amber-900">Important security notice</p>
          <p className="mt-0.5 text-xs text-amber-700">
            This is the only time these backup codes will be displayed. You will not be able to view
            them again.
          </p>
        </div>
      </div>

      {/* Grid of codes */}
      <div className="bg-gray-50 border border-gray-200 rounded-xl p-5">
        <div className="grid grid-cols-2 sm:grid-cols-2 gap-2.5 font-mono text-sm tracking-wider">
          {codes.map((code, index) => (
            <div
              key={index}
              className="bg-white px-3.5 py-2 rounded-lg border border-gray-200 text-gray-800 flex items-center justify-between font-semibold"
            >
              <span className="text-xs text-gray-400 select-none mr-2">{index + 1}.</span>
              <span className="select-all">{code}</span>
            </div>
          ))}
        </div>

        <div className="mt-5 pt-4 border-t border-gray-200 flex flex-wrap items-center justify-end gap-3">
          <Button
            type="button"
            variant="secondary"
            size="sm"
            onClick={handleCopyAll}
            className="flex items-center gap-1.5"
          >
            {copied ? (
              <>
                <Check className="h-4 w-4 text-green-600" />
                <span>Copied All Codes</span>
              </>
            ) : (
              <>
                <Copy className="h-4 w-4" />
                <span>Copy All Codes</span>
              </>
            )}
          </Button>

          <Button
            type="button"
            variant="secondary"
            size="sm"
            onClick={handleDownloadTxt}
            className="flex items-center gap-1.5"
          >
            <Download className="h-4 w-4" />
            <span>Download .txt</span>
          </Button>
        </div>
      </div>

      {showAcknowledgeCheckbox && (
        <label className="flex items-start gap-3 cursor-pointer text-sm text-gray-700 select-none">
          <input
            type="checkbox"
            checked={hasAcknowledged}
            onChange={(e) => setHasAcknowledged(e.target.checked)}
            className="mt-0.5 h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
          />
          <span>I have saved and securely backed up these 10 recovery codes.</span>
        </label>
      )}

      {onAcknowledge && (
        <Button
          type="button"
          variant="primary"
          size="lg"
          onClick={onAcknowledge}
          disabled={showAcknowledgeCheckbox && !hasAcknowledged}
          className="w-full"
        >
          {acknowledgeLabel}
        </Button>
      )}
    </div>
  )
}

export default BackupCodesDisplay
