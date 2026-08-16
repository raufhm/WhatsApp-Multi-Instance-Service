# UI Migration Plan - TanStack Implementation

**TOTP-Based Authentication with TanStack Router + React Query**

---

## Overview

This document outlines the UI implementation plan for migrating from password-based to TOTP-based authentication using:
- **TanStack Router** - Type-safe routing with loaders
- **TanStack Query (v5)** - Server state management
- **React Hook Form** - Form handling with validation
- **Zod** - Schema validation
- **Lucide React** - Icons
- **Tailwind CSS** - Styling

---

## Architecture

### Tech Stack

```json
{
  "dependencies": {
    "@tanstack/react-router": "^1.x",
    "@tanstack/react-query": "^5.x",
    "@tanstack/react-query-devtools": "^5.x",
    "react-hook-form": "^7.x",
    "@hookform/resolvers": "^3.x",
    "zod": "^3.x",
    "lucide-react": "^0.x",
    "axios": "^1.x",
    "qrcode.react": "^3.x"
  }
}
```

### Project Structure

```
frontend/src/
├── routes/                    # TanStack Router file-based routing
│   ├── __root.tsx            # Root route with AuthProvider
│   ├── index.tsx             # Dashboard (inbox) - protected
│   ├── login.tsx             # Login page (TOTP)
│   ├── signup/
│   │   ├── tenant.tsx        # Tenant signup + TOTP setup
│   │   └── operator.tsx      # Operator signup via invitation
│   ├── invitation/
│   │   └── $token.tsx        # Accept invitation with TOTP setup
│   ├── recovery.tsx          # Backup code login / recovery request
│   └── setup/
│       └── wizard.tsx        # Tenant setup wizard
├── components/
│   ├── auth/
│   │   ├── LoginForm.tsx     # TOTP login form
│   │   ├── TOTPSetup.tsx     # QR code + verification
│   │   ├── BackupCodes.tsx   # Backup codes display/download
│   │   └── RecoveryForm.tsx  # Backup code login
│   ├── ui/                   # Reusable UI components
│   └── layout/
├── hooks/
│   ├── useAuth.ts            # Auth context + queries
│   ├── useTOTP.ts            # TOTP-specific hooks
│   └── useSession.ts         # Session management
├── lib/
│   ├── apiClient.ts          # Axios instance with interceptors
│   ├── queryClient.ts        # TanStack Query configuration
│   └── utils.ts              # Utility functions
├── schemas/                  # Zod schemas
│   ├── auth.ts               # Login, signup, TOTP schemas
│   └── invitation.ts         # Invitation schemas
└── types/
    └── auth.ts               # TypeScript types
```

---

## Route Structure

### TanStack Router Configuration

```tsx
// routes/__root.tsx
import { createRootRoute, Outlet } from '@tanstack/react-router'
import { AuthProvider } from '@/hooks/useAuth'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from '@/lib/queryClient'

export const Route = createRootRoute({
  component: RootComponent,
})

function RootComponent() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <Outlet />
      </AuthProvider>
    </QueryClientProvider>
  )
}
```

### Route Tree

```tsx
// routes/index.ts
import { createRouter } from '@tanstack/react-router'
import { routeTree } from './routeTree.gen'

export const router = createRouter({
  routeTree,
  basepath: '/dashboard',
  defaultPreload: 'intent',
})

// Type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
```

### Protected Routes

```tsx
// routes/_protected.tsx
import { createRoute, Outlet, redirect } from '@tanstack/react-router'
import { useAuth } from '@/hooks/useAuth'
import { Loader2 } from 'lucide-react'

export const Route = createRoute({
  id: '_protected',
  component: ProtectedLayout,
})

function ProtectedLayout() {
  const { isAuthenticated, isLoading } = useAuth()

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    )
  }

  if (!isAuthenticated) {
    throw redirect({ to: '/login' })
  }

  return <Outlet />
}
```

### Complete Route Tree

```tsx
// routes/routeTree.gen.ts (auto-generated)
export const routeTree = rootRoute.addChildren([
  // Public routes
  loginRoute,
  recoveryRoute,
  signupTenantRoute,
  signupOperatorRoute,
  invitationRoute,
  
  // Protected routes
  protectedRoute.addChildren([
    indexRoute,              // Dashboard inbox
    conversationsRoute,
    contactsRoute,
    accountsRoute,
    botRulesRoute,
    uploadJobsRoute,
    setupWizardRoute,        // Tenant setup (if not completed)
  ]),
])
```

---

## Page Implementations

### 1. Login Page (TOTP)

```tsx
// routes/login.tsx
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { loginSchema, type LoginInput } from '@/schemas/auth'
import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { MessageCircle, AlertCircle, Loader2 } from 'lucide-react'

export const Route = createFileRoute('/login')({
  component: LoginPage,
})

function LoginPage() {
  const navigate = useNavigate()
  const { login, isAuthenticated } = useAuth()
  
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      tenantId: '',
      identifier: '', // email or WhatsApp number
      totpCode: '',
    },
  })

  const loginMutation = useMutation({
    mutationFn: (data: LoginInput) => login(data),
    onSuccess: () => {
      navigate({ to: '/' })
    },
  })

  if (isAuthenticated) {
    navigate({ to: '/' })
    return null
  }

  const onSubmit = async (data: LoginInput) => {
    await loginMutation.mutateAsync(data)
  }

  return (
    <div className="min-h-screen flex">
      {/* Left side - branding */}
      <div className="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-primary-600 to-primary-800">
        {/* Branding content */}
      </div>

      {/* Right side - login form */}
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="w-full max-w-md space-y-6">
          <div className="text-center">
            <MessageCircle className="h-12 w-12 mx-auto text-primary-600" />
            <h2 className="mt-4 text-2xl font-bold">Welcome back</h2>
            <p className="text-gray-600">
              Enter your TOTP code from authenticator app
            </p>
          </div>

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div>
              <Label htmlFor="tenantId">Tenant ID</Label>
              <Input
                id="tenantId"
                type="text"
                placeholder="e.g., 550e8400-e29b-41d4-a716-446655440000"
                {...register('tenantId')}
              />
              {errors.tenantId && (
                <p className="text-sm text-red-600">{errors.tenantId.message}</p>
              )}
            </div>

            <div>
              <Label htmlFor="identifier">Email or WhatsApp Number</Label>
              <Input
                id="identifier"
                type="text"
                placeholder="operator@example.com or +1234567890"
                {...register('identifier')}
              />
              {errors.identifier && (
                <p className="text-sm text-red-600">{errors.identifier.message}</p>
              )}
            </div>

            <div>
              <Label htmlFor="totpCode">Authentication Code</Label>
              <Input
                id="totpCode"
                type="text"
                inputMode="numeric"
                pattern="\d{6}"
                maxLength={6}
                placeholder="123 456"
                className="text-center text-2xl tracking-widest"
                {...register('totpCode')}
              />
              <TOTPCountdown />
              {errors.totpCode && (
                <p className="text-sm text-red-600">{errors.totpCode.message}</p>
              )}
            </div>

            <div className="text-right">
              <Button
                type="button"
                variant="link"
                onClick={() => navigate({ to: '/recovery' })}
                className="text-sm"
              >
                Lost access to authenticator?
              </Button>
            </div>

            <Button
              type="submit"
              variant="primary"
              size="lg"
              className="w-full"
              disabled={isSubmitting || loginMutation.isPending}
            >
              {loginMutation.isPending ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                  Signing in...
                </>
              ) : (
                'Sign In'
              )}
            </Button>

            {loginMutation.error && (
              <div className="flex items-start gap-2 p-3 bg-red-50 rounded-md">
                <AlertCircle className="h-5 w-5 text-red-600 flex-shrink-0" />
                <p className="text-sm text-red-700">
                  {loginMutation.error.message || 'Invalid credentials'}
                </p>
              </div>
            )}
          </form>
        </div>
      </div>
    </div>
  )
}
```

### 2. TOTP Setup Component

```tsx
// components/auth/TOTPSetup.tsx
import { useState } from 'react'
import { QRCodeSVG } from 'qrcode.react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card } from '@/components/ui/card'
import { Keyboard, Download, Copy, CheckCircle, AlertTriangle } from 'lucide-react'

const totpSetupSchema = z.object({
  totpCode: z.string().regex(/^\d{6}$/, 'Must be 6 digits'),
})

interface TOTPSetupProps {
  otpauthUrl: string
  secret: string
  accountName: string
  issuer: string
  onVerify: (code: string) => Promise<void>
  onSuccess: (backupCodes: string[]) => void
  onError: (error: Error) => void
}

export function TOTPSetup({
  otpauthUrl,
  secret,
  accountName,
  issuer,
  onVerify,
  onSuccess,
  onError,
}: TOTPSetupProps) {
  const [showManualEntry, setShowManualEntry] = useState(false)
  const [backupCodes, setBackupCodes] = useState<string[] | null>(null)
  const [acknowledged, setAcknowledged] = useState(false)

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(totpSetupSchema),
  })

  const verifyMutation = useMutation({
    mutationFn: onVerify,
    onSuccess: (codes) => {
      setBackupCodes(codes)
    },
    onError,
  })

  const onSubmit = async (data: { totpCode: string }) => {
    await verifyMutation.mutateAsync(data.totpCode)
  }

  const handleDownload = () => {
    if (!backupCodes) return
    const content = `Backup Codes for ${issuer}\nAccount: ${accountName}\n\n${backupCodes.join('\n')}`
    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'backup-codes.txt'
    a.click()
    URL.revokeObjectURL(url)
  }

  const handleCopy = () => {
    if (!backupCodes) return
    navigator.clipboard.writeText(backupCodes.join('\n'))
  }

  if (backupCodes) {
    return (
      <Card className="p-6">
        <div className="space-y-4">
          <div className="flex items-center gap-2 text-green-600">
            <CheckCircle className="h-6 w-6" />
            <h3 className="text-lg font-semibold">TOTP Setup Complete!</h3>
          </div>

          <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4">
            <AlertTriangle className="h-5 w-5 text-yellow-600 mb-2" />
            <h4 className="font-medium text-yellow-800">Save Your Backup Codes</h4>
            <p className="text-sm text-yellow-700 mt-1">
              Store these codes in a safe place. Each code can only be used once.
            </p>
          </div>

          <div className="bg-gray-50 rounded-md p-4 font-mono text-sm">
            <div className="grid grid-cols-2 gap-2">
              {backupCodes.map((code, index) => (
                <div key={index} className="text-center py-2 bg-white rounded border">
                  {code}
                </div>
              ))}
            </div>
          </div>

          <div className="flex gap-2">
            <Button onClick={handleDownload} variant="outline">
              <Download className="h-4 w-4 mr-2" />
              Download
            </Button>
            <Button onClick={handleCopy} variant="outline">
              <Copy className="h-4 w-4 mr-2" />
              Copy
            </Button>
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="acknowledge"
              checked={acknowledged}
              onChange={(e) => setAcknowledged(e.target.checked)}
              className="h-4 w-4"
            />
            <Label htmlFor="acknowledge" className="text-sm">
              I have saved these backup codes securely
            </Label>
          </div>

          <Button
            onClick={() => onSuccess(backupCodes)}
            disabled={!acknowledged}
            variant="primary"
            className="w-full"
          >
            Continue to Dashboard
          </Button>
        </div>
      </Card>
    )
  }

  return (
    <Card className="p-6">
      <div className="space-y-6">
        <div>
          <h3 className="text-lg font-semibold">Set Up Two-Factor Authentication</h3>
          <p className="text-sm text-gray-600">
            Scan the QR code with your authenticator app
          </p>
        </div>

        <div className="space-y-4">
          <div className="flex items-start gap-3">
            <div className="flex-shrink-0 w-6 h-6 rounded-full bg-primary-100 text-primary-600 flex items-center justify-center text-sm font-medium">
              1
            </div>
            <div>
              <h4 className="font-medium">Install an authenticator app</h4>
              <p className="text-sm text-gray-600">
                Download Google Authenticator, Authy, or 1Password
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <div className="flex-shrink-0 w-6 h-6 rounded-full bg-primary-100 text-primary-600 flex items-center justify-center text-sm font-medium">
              2
            </div>
            <div className="flex-1">
              <h4 className="font-medium">Scan the QR code</h4>
              <div className="mt-2 flex justify-center">
                <QRCodeSVG
                  value={otpauthUrl}
                  size={200}
                  level="H"
                  includeMargin={true}
                />
              </div>
              <Button
                type="button"
                variant="link"
                onClick={() => setShowManualEntry(!showManualEntry)}
                className="mt-2"
              >
                <Keyboard className="h-4 w-4 mr-2" />
                Can't scan? Enter manually
              </Button>
              {showManualEntry && (
                <div className="mt-2 p-3 bg-gray-50 rounded-md">
                  <p className="text-sm font-medium mb-1">Manual entry key:</p>
                  <code className="text-lg font-mono">
                    {secret.match(/.{1,4}/g)?.join(' ') || secret}
                  </code>
                </div>
              )}
            </div>
          </div>

          <div className="flex items-start gap-3">
            <div className="flex-shrink-0 w-6 h-6 rounded-full bg-primary-100 text-primary-600 flex items-center justify-center text-sm font-medium">
              3
            </div>
            <div>
              <h4 className="font-medium">Enter the 6-digit code</h4>
              <form onSubmit={handleSubmit(onSubmit)} className="mt-2">
                <Input
                  type="text"
                  inputMode="numeric"
                  pattern="\d{6}"
                  maxLength={6}
                  placeholder="123 456"
                  className="text-center text-2xl tracking-widest"
                  {...register('totpCode')}
                  disabled={verifyMutation.isPending}
                />
                {errors.totpCode && (
                  <p className="text-sm text-red-600 mt-1">
                    {errors.totpCode.message}
                  </p>
                )}
                {verifyMutation.error && (
                  <p className="text-sm text-red-600 mt-1">
                    {verifyMutation.error.message}
                  </p>
                )}
                <TOTPCountdown />
                <Button
                  type="submit"
                  variant="primary"
                  className="w-full mt-4"
                  disabled={verifyMutation.isPending}
                >
                  {verifyMutation.isPending ? 'Verifying...' : 'Verify and Continue'}
                </Button>
              </form>
            </div>
          </div>
        </div>
      </div>
    </Card>
  )
}
```

### 3. Recovery Page (Backup Code Login)

```tsx
// routes/recovery.tsx
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { recoverySchema, type RecoveryInput } from '@/schemas/auth'
import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card } from '@/components/ui/card'
import { Key, AlertCircle, Loader2, ArrowLeft } from 'lucide-react'

export const Route = createFileRoute('/recovery')({
  component: RecoveryPage,
})

function RecoveryPage() {
  const navigate = useNavigate()
  const { loginWithBackupCode, requestRecovery } = useAuth()
  const [mode, setMode] = useState<'backup-code' | 'request-help'>('backup-code')

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RecoveryInput>({
    resolver: zodResolver(recoverySchema),
    defaultValues: {
      tenantId: '',
      identifier: '',
      backupCode: '',
    },
  })

  const backupCodeMutation = useMutation({
    mutationFn: (data: RecoveryInput) => loginWithBackupCode(data),
    onSuccess: () => {
      navigate({ to: '/account/totp' })
    },
  })

  const requestHelpMutation = useMutation({
    mutationFn: (data: { tenantId: string; whatsappNumber: string }) =>
      requestRecovery(data),
    onSuccess: () => {
      setMode('success')
    },
  })

  const onSubmit = async (data: RecoveryInput) => {
    if (mode === 'backup-code') {
      await backupCodeMutation.mutateAsync(data)
    }
  }

  if (mode === 'success') {
    return (
      <Card className="p-6 max-w-md mx-auto">
        <div className="text-center space-y-4">
          <AlertCircle className="h-12 w-12 mx-auto text-blue-600" />
          <h2 className="text-xl font-bold">Recovery Request Sent</h2>
          <p className="text-gray-600">
            Check your WhatsApp for recovery instructions.
          </p>
          <Button onClick={() => navigate({ to: '/login' })} variant="primary">
            Back to Login
          </Button>
        </div>
      </Card>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <Card className="w-full max-w-md p-6">
        <Button
          type="button"
          variant="ghost"
          onClick={() => navigate({ to: '/login' })}
          className="mb-4"
        >
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Login
        </Button>

        <div className="text-center mb-6">
          <Key className="h-12 w-12 mx-auto text-primary-600 mb-4" />
          <h2 className="text-2xl font-bold">Account Recovery</h2>
          <p className="text-gray-600">
            Lost access to your authenticator app?
          </p>
        </div>

        <div className="flex rounded-md border mb-6">
          <button
            onClick={() => setMode('backup-code')}
            className={`flex-1 py-2 text-sm font-medium rounded-l-md ${
              mode === 'backup-code'
                ? 'bg-primary-50 text-primary-700'
                : 'bg-white text-gray-600 hover:bg-gray-50'
            }`}
          >
            Use Backup Code
          </button>
          <button
            onClick={() => setMode('request-help')}
            className={`flex-1 py-2 text-sm font-medium rounded-r-md ${
              mode === 'request-help'
                ? 'bg-primary-50 text-primary-700'
                : 'bg-white text-gray-600 hover:bg-gray-50'
            }`}
          >
            Request Help
          </button>
        </div>

        {mode === 'backup-code' && (
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div>
              <Label htmlFor="tenantId">Tenant ID</Label>
              <Input id="tenantId" {...register('tenantId')} />
              {errors.tenantId && (
                <p className="text-sm text-red-600">{errors.tenantId.message}</p>
              )}
            </div>

            <div>
              <Label htmlFor="identifier">Email or WhatsApp Number</Label>
              <Input id="identifier" {...register('identifier')} />
              {errors.identifier && (
                <p className="text-sm text-red-600">{errors.identifier.message}</p>
              )}
            </div>

            <div>
              <Label htmlFor="backupCode">Backup Code</Label>
              <Input
                id="backupCode"
                type="text"
                placeholder="A7B9-C2D4"
                className="uppercase"
                {...register('backupCode')}
              />
              {errors.backupCode && (
                <p className="text-sm text-red-600">{errors.backupCode.message}</p>
              )}
              <p className="text-xs text-gray-500 mt-1">
                Enter one of your 10 backup codes
              </p>
            </div>

            <Button
              type="submit"
              variant="primary"
              className="w-full"
              disabled={isSubmitting || backupCodeMutation.isPending}
            >
              {backupCodeMutation.isPending ? 'Verifying...' : 'Use Backup Code'}
            </Button>

            {backupCodeMutation.error && (
              <div className="flex items-start gap-2 p-3 bg-red-50 rounded-md">
                <AlertCircle className="h-5 w-5 text-red-600 flex-shrink-0" />
                <p className="text-sm text-red-700">
                  {backupCodeMutation.error.message || 'Invalid backup code'}
                </p>
              </div>
            )}
          </form>
        )}

        {mode === 'request-help' && (
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="p-4 bg-blue-50 rounded-md">
              <p className="text-sm text-blue-800">
                Your admin will reset your two-factor authentication and send you
                a new setup link via WhatsApp.
              </p>
            </div>

            <div>
              <Label htmlFor="tenantId">Tenant ID</Label>
              <Input id="tenantId" {...register('tenantId')} />
              {errors.tenantId && (
                <p className="text-sm text-red-600">{errors.tenantId.message}</p>
              )}
            </div>

            <div>
              <Label htmlFor="whatsappNumber">WhatsApp Number</Label>
              <Input
                id="whatsappNumber"
                type="tel"
                placeholder="+1234567890"
                {...register('whatsappNumber')}
              />
              {errors.whatsappNumber && (
                <p className="text-sm text-red-600">
                  {errors.whatsappNumber.message}
                </p>
              )}
            </div>

            <Button
              type="submit"
              variant="primary"
              className="w-full"
              disabled={isSubmitting || requestHelpMutation.isPending}
            >
              {requestHelpMutation.isPending ? 'Sending...' : 'Request Admin Help'}
            </Button>
          </form>
        )}
      </Card>
    </div>
  )
}
```

---

## Hooks & Query Configuration

### Auth Hook with TanStack Query

```tsx
// hooks/useAuth.ts
import { createContext, useContext, useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import apiClient from '@/lib/apiClient'

interface AuthContextType {
  isAuthenticated: boolean
  isLoading: boolean
  user: User | null
  login: (data: LoginInput) => Promise<void>
  loginWithBackupCode: (data: RecoveryInput) => Promise<void>
  logout: () => Promise<void>
  requestRecovery: (data: RecoveryRequestInput) => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [isLoading, setIsLoading] = useState(true)

  const { data: meData } = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: async () => {
      const response = await apiClient.get('/dashboard/api/me')
      return response.data
    },
    retry: false,
    staleTime: 5 * 60 * 1000,
  })

  useEffect(() => {
    setIsAuthenticated(!!meData?.user)
    setIsLoading(false)
  }, [meData])

  const loginMutation = useMutation({
    mutationFn: async (data: LoginInput) => {
      await apiClient.post('/dashboard/api/login', data)
      const meResponse = await apiClient.get('/dashboard/api/me')
      return meResponse.data
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['auth', 'me'], data)
      setIsAuthenticated(true)
    },
  })

  const logoutMutation = useMutation({
    mutationFn: async () => {
      await apiClient.post('/dashboard/api/logout')
    },
    onSuccess: () => {
      queryClient.clear()
      setIsAuthenticated(false)
      navigate({ to: '/login' })
    },
  })

  const value: AuthContextType = {
    isAuthenticated,
    isLoading,
    user: meData?.user || null,
    login: async (data: LoginInput) => {
      await loginMutation.mutateAsync(data)
    },
    loginWithBackupCode: async (data: RecoveryInput) => {
      await apiClient.post('/dashboard/api/login/backup-code', data)
      const meResponse = await apiClient.get('/dashboard/api/me')
      queryClient.setQueryData(['auth', 'me'], meResponse.data)
      setIsAuthenticated(true)
    },
    logout: async () => {
      await logoutMutation.mutateAsync()
    },
    requestRecovery: async (data: RecoveryRequestInput) => {
      await apiClient.post('/dashboard/api/recovery/request', data)
    },
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}
```

### Query Client Configuration

```ts
// lib/queryClient.ts
import { QueryClient } from '@tanstack/react-query'

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
      staleTime: 5 * 60 * 1000,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    },
    mutations: {
      retry: 0,
      onError: (error) => {
        console.error('Mutation error:', error)
      },
    },
  },
})
```

---

## Zod Schemas

```ts
// schemas/auth.ts
import { z } from 'zod'

export const loginSchema = z.object({
  tenantId: z.string().uuid('Invalid tenant ID format'),
  identifier: z.string().min(1, 'Email or WhatsApp is required'),
  totpCode: z.string().regex(/^\d{6}$/, 'Must be 6 digits'),
})

export type LoginInput = z.infer<typeof loginSchema>

export const recoverySchema = z.object({
  tenantId: z.string().uuid('Invalid tenant ID format'),
  identifier: z.string().min(1, 'Email or WhatsApp is required'),
  backupCode: z.string().regex(/^[A-Z0-9]{8}$/, 'Invalid backup code format'),
  whatsappNumber: z.string().optional(),
})

export type RecoveryInput = z.infer<typeof recoverySchema>

export const totpSetupSchema = z.object({
  totpCode: z.string().regex(/^\d{6}$/, 'Must be 6 digits'),
})

export const signupSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters'),
  email: z.string().email('Invalid email address').optional().or(z.literal('')),
  whatsappNumber: z.string().min(10, 'Invalid WhatsApp number'),
  password: z.never(), // No passwords!
})

export type SignupInput = z.infer<typeof signupSchema>
```

---

## Implementation Phases

### Phase 1: Setup & Infrastructure (Week 1)

**Tasks:**
- [ ] Install TanStack Router + React Query dependencies
- [ ] Configure TanStack Router with file-based routing
- [ ] Set up QueryClient with auth-aware defaults
- [ ] Create Zod schemas for TOTP authentication
- [ ] Update apiClient.ts with TOTP login interceptors
- [ ] Create AuthContext with TanStack Query integration

**Files to create:**
```
frontend/src/
├── lib/queryClient.ts
├── schemas/auth.ts
├── hooks/useAuth.ts (update)
└── routes/__root.tsx
```

### Phase 2: Login & Recovery UI (Week 2)

**Tasks:**
- [ ] Create Login page with TOTP input
- [ ] Create TOTPCountdown component (30-second timer)
- [ ] Create Recovery page with backup code + request help modes
- [ ] Add route guards for protected routes
- [ ] Test login flow with backend TOTP endpoint
- [ ] Test backup code recovery flow

**Files to create:**
```
frontend/src/
├── routes/login.tsx
├── routes/recovery.tsx
├── components/auth/LoginForm.tsx
├── components/auth/RecoveryForm.tsx
└── components/auth/TOTPCountdown.tsx
```

### Phase 3: TOTP Setup UI (Week 3)

**Tasks:**
- [ ] Create TOTPSetup component with QR code
- [ ] Add manual entry fallback for TOTP secret
- [ ] Create BackupCodes component with download/copy
- [ ] Integrate TOTP setup into signup flow
- [ ] Integrate TOTP setup into invitation acceptance flow
- [ ] Test QR code generation with backend
- [ ] Test TOTP verification flow

**Files to create:**
```
frontend/src/
├── components/auth/TOTPSetup.tsx
├── components/auth/BackupCodes.tsx
├── components/auth/TOTPCountdown.tsx
├── routes/signup/tenant.tsx
├── routes/signup/operator.tsx
└── routes/invitation/$token.tsx
```

### Phase 4: Invitation Flow (Week 4)

**Tasks:**
- [ ] Create invitation acceptance page with token validation
- [ ] Implement operator signup via WhatsApp invitation
- [ ] Add TOTP setup step to invitation flow
- [ ] Add WhatsApp number input and validation
- [ ] Integrate with backend invitation acceptance endpoint
- [ ] Test complete invitation-to-login flow

**Files to create:**
```
frontend/src/
├── routes/invitation/$token.tsx
├── schemas/invitation.ts
└── hooks/useInvitation.ts
```

### Phase 5: Account Management (Week 5)

**Tasks:**
- [ ] Create account settings page
- [ ] Implement TOTP regeneration flow
- [ ] Add backup code regeneration endpoint integration
- [ ] Create WhatsApp number management UI
- [ ] Add session management (view active sessions, logout all)
- [ ] Implement passwordless account security UI

**Files to create:**
```
frontend/src/
├── routes/account/totp.tsx
├── routes/account/security.tsx
├── components/account/TOTPRegenerate.tsx
└── components/account/BackupCodesRegenerate.tsx
```

### Phase 6: Polish & Testing (Week 6)

**Tasks:**
- [ ] Add loading states to all auth flows
- [ ] Implement error boundaries for auth errors
- [ ] Add toast notifications for auth events
- [ ] Create comprehensive E2E tests (Playwright/Cypress)
- [ ] Accessibility audit (WCAG 2.1 AA)
- [ ] Mobile responsiveness testing
- [ ] Performance optimization (code splitting, lazy loading)

**Test coverage:**
- [ ] TOTP login success/failure
- [ ] Backup code login success/failure
- [ ] TOTP setup with QR code
- [ ] TOTP setup with manual entry
- [ ] Backup code download
- [ ] Recovery request flow
- [ ] Invitation acceptance flow
- [ ] Route guards (authenticated vs unauthenticated)

---

## Component Library Recommendations

### UI Components (shadcn/ui)

**Recommended**: Use [shadcn/ui](https://ui.shadcn.com/) for consistent, accessible components

```bash
npx shadcn-ui@latest init
npx shadcn-ui@latest add button input label card alert dialog
```

**Benefits**:
- ✅ Built with Tailwind CSS
- ✅ Radix UI primitives (accessible)
- ✅ Fully customizable
- ✅ TypeScript support
- ✅ No runtime dependency (copy-paste components)

### Alternative: Headless UI

```bash
npm install @headlessui/react @heroicons/react
```

**Benefits**:
- ✅ Completely unstyled
- ✅ Built by Tailwind CSS team
- ✅ Fully accessible
- ✅ React-only (no framework lock-in)

---

## Key UI Components Reference

### TOTPCountdown Component

```tsx
// components/auth/TOTPCountdown.tsx
import { useEffect, useState } from 'react'

export function TOTPCountdown() {
  const [seconds, setSeconds] = useState(30 - (Math.floor(Date.now() / 1000) % 30))

  useEffect(() => {
    const interval = setInterval(() => {
      setSeconds(30 - (Math.floor(Date.now() / 1000) % 30))
    }, 1000)

    return () => clearInterval(interval)
  }, [])

  const isExpiring = seconds <= 5

  return (
    <div className="flex items-center justify-between text-sm mt-2">
      <span className="text-gray-600">Code expires in:</span>
      <span className={`font-mono font-medium ${
        isExpiring ? 'text-red-600' : 'text-gray-700'
      }`}>
        {seconds.toString().padStart(2, '0')}s
      </span>
    </div>
  )
}
```

### BackupCodes Component

```tsx
// components/auth/BackupCodes.tsx
import { Download, Copy, Check } from 'lucide-react'

interface BackupCodesProps {
  codes: string[]
  onAcknowledge: () => void
  acknowledged: boolean
}

export function BackupCodes({ codes, onAcknowledge, acknowledged }: BackupCodesProps) {
  const [copied, setCopied] = useState(false)

  const handleDownload = () => {
    const content = `Backup Codes\n\n${codes.join('\n')}`
    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'backup-codes.txt'
    a.click()
    URL.revokeObjectURL(url)
  }

  const handleCopy = async () => {
    await navigator.clipboard.writeText(codes.join('\n'))
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="space-y-4">
      <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4">
        <p className="text-sm text-yellow-800">
          <strong>Important:</strong> Save these codes in a safe place. Each code can only be used once.
        </p>
      </div>

      <div className="bg-gray-50 rounded-md p-4 font-mono text-sm">
        <div className="grid grid-cols-2 gap-2">
          {codes.map((code, index) => (
            <div key={index} className="text-center py-2 bg-white rounded border">
              {code}
            </div>
          ))}
        </div>
      </div>

      <div className="flex gap-2">
        <button onClick={handleDownload} className="flex-1 flex items-center justify-center gap-2 px-4 py-2 border rounded-md hover:bg-gray-50">
          <Download className="h-4 w-4" />
          Download
        </button>
        <button onClick={handleCopy} className="flex-1 flex items-center justify-center gap-2 px-4 py-2 border rounded-md hover:bg-gray-50">
          {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
          {copied ? 'Copied!' : 'Copy'}
        </button>
      </div>

      <label className="flex items-center gap-2">
        <input
          type="checkbox"
          checked={acknowledged}
          onChange={() => onAcknowledge()}
          className="h-4 w-4"
        />
        <span className="text-sm">I have saved these backup codes securely</span>
      </label>
    </div>
  )
}
```

---

## API Integration Patterns

### Query Keys

```ts
// lib/queryKeys.ts
export const queryKeys = {
  auth: {
    me: ['auth', 'me'],
    session: ['auth', 'session'],
  },
  totp: {
    setup: (operatorId: string) => ['totp', operatorId, 'setup'],
    backupCodes: (operatorId: string) => ['totp', operatorId, 'backup-codes'],
  },
  invitations: {
    byToken: (token: string) => ['invitations', token],
  },
}
```

### Mutation Hooks Pattern

```ts
// hooks/useTOTP.ts
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import apiClient from '@/lib/apiClient'
import { queryKeys } from '@/lib/queryKeys'

export function useTOTPSetup() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async () => {
      const response = await apiClient.post('/dashboard/api/totp/setup')
      return response.data
    },
    onSuccess: (data) => {
      queryClient.setQueryData(queryKeys.totp.setup('me'), data)
    },
  })
}

export function useTOTPVerify() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (totpCode: string) => {
      const response = await apiClient.post('/dashboard/api/totp/verify', { totpCode })
      return response.data
    },
    onSuccess: (data) => {
      // Invalidate auth query to refresh user state
      queryClient.invalidateQueries({ queryKey: queryKeys.auth.me })
      // Invalidate backup codes to show new codes
      queryClient.invalidateQueries({ queryKey: queryKeys.totp.backupCodes('me') })
    },
  })
}
```

---

## Testing Strategy

### Unit Tests (Vitest + React Testing Library)

```ts
// __tests__/TOTPSetup.test.tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { TOTPSetup } from '@/components/auth/TOTPSetup'

describe('TOTPSetup', () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      mutations: { retry: false },
    },
  })

  const renderWithProviders = (ui: React.ReactElement) => {
    return render(
      <QueryClientProvider client={queryClient}>
        {ui}
      </QueryClientProvider>
    )
  }

  it('shows QR code on mount', () => {
    renderWithProviders(
      <TOTPSetup
        otpauthUrl="otpauth://totp/test"
        secret="TESTSECRET"
        accountName="test@example.com"
        issuer="Test App"
        onVerify={async () => {}}
        onSuccess={() => {}}
        onError={() => {}}
      />
    )

    expect(screen.getByText('Set Up Two-Factor Authentication')).toBeInTheDocument()
    expect(screen.getByAltText(/qr code/i)).toBeInTheDocument()
  })

  it('shows manual entry when requested', async () => {
    renderWithProviders(
      <TOTPSetup
        otpauthUrl="otpauth://totp/test"
        secret="TESTSECRET"
        accountName="test@example.com"
        issuer="Test App"
        onVerify={async () => {}}
        onSuccess={() => {}}
        onError={() => {}}
      />
    )

    fireEvent.click(screen.getByText(/can't scan\?/i))
    await waitFor(() => {
      expect(screen.getByText(/manual entry key/i)).toBeInTheDocument()
    })
  })

  it('calls onVerify with TOTP code', async () => {
    const onVerify = vi.fn()
    const onSuccess = vi.fn()
    
    renderWithProviders(
      <TOTPSetup
        otpauthUrl="otpauth://totp/test"
        secret="TESTSECRET"
        accountName="test@example.com"
        issuer="Test App"
        onVerify={onVerify}
        onSuccess={onSuccess}
        onError={() => {}}
      />
    )

    fireEvent.change(screen.getByPlaceholderText(/123 456/i), {
      target: { value: '123456' },
    })
    fireEvent.click(screen.getByText(/verify and continue/i))

    await waitFor(() => {
      expect(onVerify).toHaveBeenCalledWith('123456')
    })
  })
})
```

### E2E Tests (Playwright)

```ts
// e2e/totp-login.spec.ts
import { test, expect } from '@playwright/test'

test.describe('TOTP Login', () => {
  test('successful TOTP login', async ({ page }) => {
    await page.goto('/login')

    // Fill login form
    await page.fill('[name="tenantId"]', '550e8400-e29b-41d4-a716-446655440000')
    await page.fill('[name="identifier"]', 'operator@example.com')
    await page.fill('[name="totpCode"]', '123456')

    // Submit
    await page.click('button[type="submit"]')

    // Wait for navigation to dashboard
    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('h1')).toContainText('Inbox')
  })

  test('invalid TOTP code shows error', async ({ page }) => {
    await page.goto('/login')

    await page.fill('[name="tenantId"]', '550e8400-e29b-41d4-a716-446655440000')
    await page.fill('[name="identifier"]', 'operator@example.com')
    await page.fill('[name="totpCode"]', '000000')

    await page.click('button[type="submit"]')

    // Wait for error message
    await expect(page.locator('[role="alert"]')).toContainText('Invalid authentication code')
  })

  test('navigates to recovery page', async ({ page }) => {
    await page.goto('/login')

    await page.click('text=Lost access to authenticator?')

    await expect(page).toHaveURL('/recovery')
    await expect(page.locator('h2')).toContainText('Account Recovery')
  })
})

// e2e/totp-setup.spec.ts
import { test, expect } from '@playwright/test'

test.describe('TOTP Setup', () => {
  test('complete TOTP setup flow', async ({ page }) => {
    // Navigate to invitation link
    await page.goto('/invitation/test-token-123')

    // Fill operator details
    await page.fill('[name="name"]', 'Test Operator')
    await page.fill('[name="whatsappNumber"]', '+1234567890')
    await page.click('button[type="submit"]')

    // Wait for TOTP setup page
    await expect(page.locator('h3')).toContainText('Set Up Two-Factor Authentication')

    // Verify QR code is present
    await expect(page.locator('svg[data-testid="qr-code"]')).toBeVisible()

    // Enter TOTP code (mocked backend will accept any code in test)
    await page.fill('[name="totpCode"]', '123456')
    await page.click('button[type="submit"]')

    // Wait for backup codes
    await expect(page.locator('text=Backup Codes')).toBeVisible()

    // Acknowledge backup codes
    await page.click('text=I have saved these backup codes securely')
    await page.click('text=Continue to Dashboard')

    // Should be on dashboard
    await expect(page).toHaveURL('/dashboard')
  })
})
```

---

## Accessibility Checklist

- [ ] All form inputs have associated `<label>` elements
- [ ] Error messages are announced by screen readers (`aria-live="polite"`)
- [ ] QR code has descriptive `alt` text
- [ ] TOTP input has `inputMode="numeric"` for mobile keyboards
- [ ] TOTP input has `aria-describedby` for instructions
- [ ] Backup codes can be navigated with keyboard
- [ ] Loading states use `aria-busy="true"`
- [ ] Color is not the only indicator of state (e.g., error states have icons + text)
- [ ] Focus states are visible and high-contrast
- [ ] All interactive elements are at least 44x44px touch target

---

## Performance Optimizations

### Code Splitting with TanStack Router

```tsx
// routes/__root.tsx
import { createRootRoute, lazyRouteComponent } from '@tanstack/react-router'

export const Route = createRootRoute({
  component: () => (
    <>
      <Outlet />
      <TanStackRouterDevtools position="bottom-right" />
    </>
  ),
})

// Lazy load routes
const loginRoute = createFileRoute('/login')({
  component: lazyRouteComponent(() => import('./routes/login')),
})

const recoveryRoute = createFileRoute('/recovery')({
  component: lazyRouteComponent(() => import('./routes/recovery')),
})
```

### Query Optimization

```ts
// lib/queryClient.ts
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Cache auth data for 5 minutes
      staleTime: 5 * 60 * 1000,
      // Only retry once to avoid long delays on auth failures
      retry: 1,
      // Don't refetch on window focus for auth queries
      refetchOnWindowFocus: false,
      // Structural sharing for faster updates
      structuralSharing: true,
    },
    mutations: {
      // No retries on mutations (auth operations should fail fast)
      retry: 0,
    },
  },
})
```

---

## Security Considerations

### Frontend Security Checklist

- [ ] Never log TOTP codes or backup codes
- [ ] Use HTTPS for all API calls (enforced in production)
- [ ] Set `withCredentials: true` in Axios for session cookies
- [ ] Implement CSRF token handling (if not using SameSite cookies)
- [ ] Sanitize all user inputs (Zod validation)
- [ ] Implement rate limiting on frontend (prevent accidental spam)
- [ ] Clear sensitive data on logout
- [ ] Use `rel="noopener noreferrer"` for external links
- [ ] Implement Content Security Policy (CSP)
- [ ] Mask WhatsApp numbers in UI (show last 4 digits only)

### Cookie Configuration

```ts
// lib/apiClient.ts
const apiClient = axios.create({
  baseURL: '/dashboard/api',
  withCredentials: true, // Send cookies
  headers: {
    'Content-Type': 'application/json',
  },
})
```

Backend must set cookies with:
```http
Set-Cookie: session=...; HttpOnly; Secure; SameSite=Lax; Path=/
```

---

## Migration Checklist

### From Password to TOTP

**Before Migration**:
- [ ] Backup current user data
- [ ] Test TOTP flow in staging environment
- [ ] Prepare rollback plan
- [ ] Update user documentation
- [ ] Train support team on TOTP recovery

**During Migration**:
- [ ] Deploy backend TOTP endpoints first
- [ ] Deploy frontend TOTP UI
- [ ] Monitor error rates and user feedback
- [ ] Keep password login available in parallel (optional)

**After Migration**:
- [ ] Verify all users can login with TOTP
- [ ] Remove password fields from database (after grace period)
- [ ] Remove password reset endpoints
- [ ] Update all documentation to reflect TOTP
- [ ] Monitor support tickets for TOTP issues

---

## Resources

### Documentation

- [TanStack Router](https://tanstack.com/router/latest/docs/framework/react/overview)
- [TanStack Query](https://tanstack.com/query/latest/docs/framework/react/overview)
- [React Hook Form](https://react-hook-form.com/)
- [Zod](https://zod.dev/)
- [shadcn/ui](https://ui.shadcn.com/)
- [TOTP RFC 6238](https://tools.ietf.org/html/rfc6238)

### Libraries

```bash
# Core
npm install @tanstack/react-router @tanstack/react-query
npm install react-hook-form @hookform/resolvers zod

# UI
npm install lucide-react qrcode.react
npx shadcn-ui@latest init

# Development
npm install -D @types/node
npm install -D vitest @testing-library/react @testing-library/jest-dom
npm install -D @playwright/test
```

---

## Summary

This UI migration plan provides a comprehensive roadmap for implementing TOTP-based authentication using TanStack Router and React Query. The implementation is divided into 6 phases over 6 weeks, with clear deliverables and testing requirements.

**Key Benefits**:
- ✅ Type-safe routing with TanStack Router
- ✅ Efficient server state management with React Query
- ✅ Robust form handling with React Hook Form + Zod
- ✅ Modern, accessible UI components
- ✅ Comprehensive testing strategy
- ✅ Clear migration path from passwords

**Start with Phase 1 and you'll have a production-ready TOTP authentication system in 6 weeks**! 🚀
