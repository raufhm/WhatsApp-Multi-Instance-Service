import { createRoute, Navigate, Outlet } from '@tanstack/react-router'
import { lazy } from 'react'
import { useAuth } from '@/hooks/useAuth'
import Layout from '@/components/Layout'
import { rootRoute } from './root'

// Lazy page components
const Inbox = lazy(() => import('@/pages/Inbox'))
const Contacts = lazy(() => import('@/pages/Contacts'))
const ContactDetail = lazy(() => import('@/pages/ContactDetail'))
const Accounts = lazy(() => import('@/pages/Accounts'))
const Team = lazy(() => import('@/pages/Team'))
const BotRules = lazy(() => import('@/pages/BotRules'))
const UploadJobs = lazy(() => import('@/pages/UploadJobs'))
const PipelinesSettings = lazy(() => import('@/pages/PipelinesSettings'))
const SetupWizard = lazy(() => import('@/pages/SetupWizard'))
const TotpSettings = lazy(() => import('@/pages/TotpSettings'))

function AuthenticatedLayout() {
  const { isAuthenticated, isLoading } = useAuth()

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary-500" />
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/" />
  }

  return (
    <Layout>
      <Outlet />
    </Layout>
  )
}

export const authenticatedRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: '_authenticated',
  component: AuthenticatedLayout,
})

// Authenticated child routes
export const indexRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/inbox',
  component: Inbox,
})

export const conversationRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/conversations/$id',
  component: Inbox,
})

export const contactsRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/contacts',
  component: Contacts,
})

export const contactDetailRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/contacts/$id',
  component: ContactDetail,
})

export const settingsPipelinesRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/settings/pipelines',
  component: PipelinesSettings,
})

export const accountsRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/accounts',
  component: Accounts,
})

export const teamRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/team',
  component: Team,
})

export const invitationsManageRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/invitations',
  component: Team,
})

export const botRulesRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/bot-rules',
  component: BotRules,
})

export const uploadJobsRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/upload-jobs',
  component: UploadJobs,
})

export const setupWizardRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/setup',
  component: SetupWizard,
})

export const totpSettingsRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/account/totp',
  component: TotpSettings,
})
