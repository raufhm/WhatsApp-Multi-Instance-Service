import { describe, it, expect } from 'vitest'
import {
  TOTP_ISSUER,
  TOTP_ALGORITHM,
  TOTP_DIGITS,
  TOTP_PERIOD,
  buildOtpauthUrl,
  formatSecretKey,
  generateQrSvg,
} from './qrCode'

describe('qrCode and TOTP URI builder', () => {
  it('has standard RFC 6238 constants matching backend', () => {
    expect(TOTP_ISSUER).toBe('whops')
    expect(TOTP_ALGORITHM).toBe('SHA1')
    expect(TOTP_DIGITS).toBe(6)
    expect(TOTP_PERIOD).toBe(30)
  })

  it('matches Go backend golden string URI output exactly', () => {
    const account = 'user@example.com'
    const secret = 'JBSWY3DPEHPK3PXP'
    const expected =
      'otpauth://totp/whops:user@example.com?algorithm=SHA1&digits=6&issuer=whops&period=30&secret=JBSWY3DPEHPK3PXP'

    const actual = buildOtpauthUrl(account, secret)
    expect(actual).toBe(expected)
  })

  it('handles custom issuer and handles whitespace in secret', () => {
    const account = 'operator1'
    const secret = 'jbsw y3dp ehpk 3pxp'
    const issuer = 'Acme Workspace'
    const expected =
      'otpauth://totp/Acme%20Workspace:operator1?algorithm=SHA1&digits=6&issuer=Acme+Workspace&period=30&secret=JBSWY3DPEHPK3PXP'

    const actual = buildOtpauthUrl(account, secret, issuer)
    expect(actual).toBe(expected)
  })

  it('formats secret key into 4-character chunks', () => {
    expect(formatSecretKey('JBSWY3DPEHPK3PXP')).toBe('JBSW Y3DP EHPK 3PXP')
    expect(formatSecretKey('jbsw y3dp ehpk 3pxp')).toBe('JBSW Y3DP EHPK 3PXP')
    expect(formatSecretKey('')).toBe('')
  })

  it('generates valid QR SVG markup', () => {
    const url = buildOtpauthUrl('admin', 'JBSWY3DPEHPK3PXP')
    const svg = generateQrSvg(url, 200)

    expect(svg).toContain('<svg')
    expect(svg).toContain('viewBox=')
    expect(svg).toContain('<rect')
    expect(svg).toContain('<path')
    expect(svg).toContain('</svg>')
  })
})
