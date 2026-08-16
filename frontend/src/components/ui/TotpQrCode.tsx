import React, { useState } from 'react'
import { Copy, Check, QrCode, KeyRound, ShieldCheck } from 'lucide-react'
import Button from './button'
import { generateQrSvg, formatSecretKey, buildOtpauthUrl, TOTP_ISSUER } from '@/lib/qrCode'

interface TotpQrCodeProps {
  secret: string
  otpauthUrl?: string
  qrSvg?: string
  qrDataUrl?: string
  issuer?: string
  accountName?: string
  className?: string
}

export const TotpQrCode: React.FC<TotpQrCodeProps> = ({
  secret,
  otpauthUrl,
  qrSvg,
  qrDataUrl,
  issuer = TOTP_ISSUER,
  accountName,
  className = '',
}) => {
  const [copied, setCopied] = useState(false)
  const [showManual, setShowManual] = useState(false)

  const effectiveOtpUrl =
    otpauthUrl || buildOtpauthUrl(accountName || 'User', secret, issuer)

  const svgContent = qrSvg || (!qrDataUrl ? generateQrSvg(effectiveOtpUrl, 200) : '')

  const handleCopySecret = async () => {
    try {
      await navigator.clipboard.writeText(secret.replace(/\s+/g, ''))
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // fallback
      const el = document.createElement('textarea')
      el.value = secret.replace(/\s+/g, '')
      document.body.appendChild(el)
      el.select()
      document.execCommand('copy')
      document.body.removeChild(el)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  return (
    <div className={`flex flex-col items-center space-y-5 ${className}`}>
      <div className="bg-white p-4 rounded-xl border border-gray-200 shadow-sm flex flex-col items-center">
        {qrDataUrl ? (
          <img
            src={qrDataUrl}
            alt="Scan this QR code with your authenticator app"
            className="w-48 h-48 object-contain rounded"
          />
        ) : svgContent ? (
          <div
            className="w-48 h-48 flex items-center justify-center"
            dangerouslySetInnerHTML={{ __html: svgContent }}
            aria-label="TOTP QR Code"
          />
        ) : (
          <div className="w-48 h-48 flex items-center justify-center bg-gray-100 rounded text-gray-400">
            <QrCode className="h-16 w-16" />
          </div>
        )}
        <div className="mt-3 flex items-center gap-1.5 text-xs text-gray-500">
          <ShieldCheck className="h-4 w-4 text-primary-600" />
          <span>Scan with Google Authenticator, 1Password, or Authy</span>
        </div>
      </div>

      <div className="w-full max-w-sm">
        <button
          type="button"
          onClick={() => setShowManual(!showManual)}
          className="w-full text-center text-sm font-medium text-primary-600 hover:text-primary-700 flex items-center justify-center gap-1.5 py-1"
        >
          <KeyRound className="h-4 w-4" />
          {showManual ? 'Hide manual entry key' : "Can't scan QR code? Enter key manually"}
        </button>

        {showManual && (
          <div className="mt-3 p-3.5 bg-gray-50 rounded-lg border border-gray-200 text-left animate-fadeIn">
            <p className="text-xs text-gray-500 font-medium mb-1">Manual Account Secret Key:</p>
            <div className="flex items-center justify-between gap-2 bg-white px-3 py-2 rounded border border-gray-200 font-mono text-sm tracking-wider text-gray-800">
              <span className="select-all break-all">{formatSecretKey(secret)}</span>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={handleCopySecret}
                className="h-7 px-2 text-gray-500 hover:text-primary-600"
                aria-label="Copy secret key"
              >
                {copied ? <Check className="h-4 w-4 text-green-600" /> : <Copy className="h-4 w-4" />}
              </Button>
            </div>
            <p className="text-xs text-gray-400 mt-1.5">
              Type: Time-based (TOTP) • Digits: 6 • Interval: 30s
            </p>
          </div>
        )}
      </div>
    </div>
  )
}

export default TotpQrCode
