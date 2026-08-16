import React from 'react'
import { Phone } from 'lucide-react'
import { Label } from './label'

interface PhoneInputProps {
  value: string
  onChange: (value: string) => void
  id?: string
  label?: string
  placeholder?: string
  required?: boolean
  disabled?: boolean
  error?: string
  hint?: string
  className?: string
}

const COMMON_PREFIXES = [
  { code: '+1', country: 'US/CA' },
  { code: '+44', country: 'UK' },
  { code: '+49', country: 'DE' },
  { code: '+33', country: 'FR' },
  { code: '+62', country: 'ID' },
  { code: '+91', country: 'IN' },
  { code: '+55', country: 'BR' },
  { code: '+61', country: 'AU' },
  { code: '+81', country: 'JP' },
  { code: '+65', country: 'SG' },
  { code: '+971', country: 'AE' },
  { code: '+34', country: 'ES' },
  { code: '+39', country: 'IT' },
  { code: '+52', country: 'MX' },
]

export const PhoneInput: React.FC<PhoneInputProps> = ({
  value = '',
  onChange,
  id = 'phone-input',
  label = 'WhatsApp Number',
  placeholder = '+1 (555) 000-0000',
  required = false,
  disabled = false,
  error,
  hint = 'Include country code (e.g. +14155552671)',
  className = '',
}) => {
  // Handle direct number changes
  const handleNumberChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    let raw = e.target.value
    // Allow digits, plus, spaces, dashes, parentheses
    raw = raw.replace(/[^\d+ ()-]/g, '')
    onChange(raw)
  }

  const handlePrefixSelect = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const prefix = e.target.value
    if (!prefix) return

    // Strip current leading +...
    const numberWithoutPrefix = value.replace(/^\+\d+[\s-]*/, '')
    onChange(`${prefix} ${numberWithoutPrefix}`.trim())
  }

  // Find matching prefix if any
  const matchedPrefix = COMMON_PREFIXES.find((p) => value.startsWith(p.code))?.code || ''

  return (
    <div className={`space-y-1.5 ${className}`}>
      {label && (
        <Label htmlFor={id} className="flex items-center gap-1.5 text-sm font-medium text-gray-700">
          <Phone className="h-4 w-4 text-gray-500" />
          <span>{label}</span>
          {required && <span className="text-red-500">*</span>}
        </Label>
      )}

      <div className="flex rounded-md shadow-sm">
        <select
          value={matchedPrefix}
          onChange={handlePrefixSelect}
          disabled={disabled}
          aria-label="Country prefix"
          className="inline-flex items-center px-3 rounded-l-md border border-r-0 border-gray-300 bg-gray-50 text-gray-600 text-xs sm:text-sm focus:ring-primary-500 focus:border-primary-500 disabled:opacity-50"
        >
          <option value="">Prefix</option>
          {COMMON_PREFIXES.map((p) => (
            <option key={p.code} value={p.code}>
              {p.code} ({p.country})
            </option>
          ))}
        </select>

        <input
          id={id}
          type="tel"
          value={value}
          onChange={handleNumberChange}
          placeholder={placeholder}
          required={required}
          disabled={disabled}
          className={`flex-1 min-w-0 block w-full px-3 py-2 rounded-none rounded-r-md border text-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 ${
            error ? 'border-red-300 bg-red-50 text-red-900' : 'border-gray-300 text-gray-900'
          } ${disabled ? 'bg-gray-100 opacity-50 cursor-not-allowed' : ''}`}
        />
      </div>

      {error ? (
        <p className="text-xs text-red-600">{error}</p>
      ) : hint ? (
        <p className="text-xs text-gray-500">{hint}</p>
      ) : null}
    </div>
  )
}

export default PhoneInput
