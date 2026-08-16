import React, { useRef, useEffect } from 'react'

interface TotpCodeInputProps {
  value: string
  onChange: (value: string) => void
  onComplete?: (code: string) => void
  disabled?: boolean
  autoFocus?: boolean
  error?: boolean
  className?: string
  id?: string
}

export const TotpCodeInput: React.FC<TotpCodeInputProps> = ({
  value = '',
  onChange,
  onComplete,
  disabled = false,
  autoFocus = true,
  error = false,
  className = '',
  id = 'totp-input',
}) => {
  const inputsRef = useRef<(HTMLInputElement | null)[]>([])

  // Ensure digits array has 6 items
  const digits = Array.from({ length: 6 }, (_, i) => value[i] || '')

  useEffect(() => {
    if (autoFocus && inputsRef.current[0] && !disabled) {
      inputsRef.current[0]?.focus()
    }
  }, [autoFocus, disabled])

  const handleInputChange = (index: number, val: string) => {
    // Sanitize input to only digits
    const cleanVal = val.replace(/\D/g, '')

    if (cleanVal.length > 1) {
      // User pasted or typed multiple digits in one box
      handlePasteData(cleanVal, index)
      return
    }

    const newDigits = [...digits]
    newDigits[index] = cleanVal
    const nextVal = newDigits.join('')
    onChange(nextVal)

    // Advance focus if a digit was entered
    if (cleanVal && index < 5) {
      inputsRef.current[index + 1]?.focus()
    }

    if (nextVal.length === 6 && onComplete) {
      onComplete(nextVal)
    }
  }

  const handlePasteData = (pasteData: string, fromIndex = 0) => {
    const cleanNumbers = pasteData.replace(/\D/g, '').slice(0, 6)
    if (!cleanNumbers) return

    const newDigits = [...digits]
    for (let i = 0; i < cleanNumbers.length && fromIndex + i < 6; i++) {
      newDigits[fromIndex + i] = cleanNumbers[i]
    }
    const nextVal = newDigits.join('')
    onChange(nextVal)

    const nextFocusIdx = Math.min(fromIndex + cleanNumbers.length, 5)
    inputsRef.current[nextFocusIdx]?.focus()

    if (nextVal.length === 6 && onComplete) {
      onComplete(nextVal)
    }
  }

  const handleKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace') {
      if (!digits[index] && index > 0) {
        // Move back and erase previous
        e.preventDefault()
        const newDigits = [...digits]
        newDigits[index - 1] = ''
        onChange(newDigits.join(''))
        inputsRef.current[index - 1]?.focus()
      } else if (digits[index]) {
        // Erase current
        const newDigits = [...digits]
        newDigits[index] = ''
        onChange(newDigits.join(''))
      }
    } else if (e.key === 'ArrowLeft' && index > 0) {
      e.preventDefault()
      inputsRef.current[index - 1]?.focus()
    } else if (e.key === 'ArrowRight' && index < 5) {
      e.preventDefault()
      inputsRef.current[index + 1]?.focus()
    }
  }

  const handlePaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault()
    const paste = e.clipboardData.getData('text')
    handlePasteData(paste, 0)
  }

  return (
    <div className={`flex flex-col items-center ${className}`}>
      <div className="flex items-center justify-center gap-2 sm:gap-3" role="group" aria-label="6-digit authentication code">
        {Array.from({ length: 6 }).map((_, idx) => (
          <React.Fragment key={idx}>
            {idx === 3 && <div className="w-2 h-0.5 bg-gray-300 rounded self-center" aria-hidden="true" />}
            <input
              ref={(el) => (inputsRef.current[idx] = el)}
              id={`${id}-${idx}`}
              type="text"
              inputMode="numeric"
              pattern="[0-9]*"
              maxLength={1}
              value={digits[idx]}
              disabled={disabled}
              onChange={(e) => handleInputChange(idx, e.target.value)}
              onKeyDown={(e) => handleKeyDown(idx, e)}
              onPaste={handlePaste}
              onFocus={(e) => e.target.select()}
              aria-label={`Digit ${idx + 1}`}
              className={`w-11 h-13 sm:w-12 sm:h-14 text-center font-mono text-xl sm:text-2xl font-bold rounded-lg border transition-all focus:outline-none focus:ring-2 ${
                error
                  ? 'border-red-300 bg-red-50 text-red-900 focus:ring-red-500'
                  : digits[idx]
                  ? 'border-primary-500 bg-white text-gray-900 shadow-sm focus:ring-primary-500'
                  : 'border-gray-300 bg-white text-gray-900 focus:border-primary-500 focus:ring-primary-500'
              } ${disabled ? 'opacity-50 cursor-not-allowed bg-gray-100' : ''}`}
            />
          </React.Fragment>
        ))}
      </div>
    </div>
  )
}

export default TotpCodeInput
