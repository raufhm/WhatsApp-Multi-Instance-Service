import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import TotpQrCode from './TotpQrCode'

describe('TotpQrCode component', () => {
  it('renders QR code using provided otpauthUrl or fallback', () => {
    const { container } = render(
      <TotpQrCode
        secret="JBSWY3DPEHPK3PXP"
        accountName="test@example.com"
      />
    )

    expect(screen.getByLabelText('TOTP QR Code')).toBeInTheDocument()
    expect(screen.getByText("Can't scan QR code? Enter key manually")).toBeInTheDocument()
    const svg = container.querySelector('svg')
    expect(svg).toBeInTheDocument()
  })

  it('renders with qrDataUrl if provided', () => {
    render(
      <TotpQrCode
        secret="JBSWY3DPEHPK3PXP"
        qrDataUrl="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
      />
    )

    const img = screen.getByAltText('Scan this QR code with your authenticator app')
    expect(img).toBeInTheDocument()
  })
})
