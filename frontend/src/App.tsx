import {
  createRouter,
  RouterProvider,
  Outlet,
  createRootRoute,
  createRoute,
  Navigate,
} from '@tanstack/react-router'
import { useAuth, AuthProvider } from '@/hooks/useAuth'
import Layout from '@/components/Layout'
import Login from '@/pages/Login'
import SignupChoice from '@/pages/SignupChoice'
import SignupTenant from '@/pages/SignupTenant'
import VerifyEmail from '@/pages/VerifyEmail'
import OperatorInvitation from '@/pages/OperatorInvitation'
import JoinWithCode from '@/pages/JoinWithCode'
import Recovery from '@/pages/Recovery'
import SetupWizard from '@/pages/SetupWizard'
import TotpSettings from '@/pages/TotpSettings'
import Team from '@/pages/Team'
import Inbox from '@/pages/Inbox'
import ConversationDetail from '@/pages/ConversationDetail'
import Contacts from '@/pages/Contacts'
import ContactDetail from '@/pages/ContactDetail'
import Accounts from '@/pages/Accounts'
import BotRules from '@/pages/BotRules'
import UploadJobs from '@/pages/UploadJobs'

const rootRoute = createRootRoute({
  component: () => <Outlet />,
})

function AuthenticatedLayout() {
  const { isAuthenticated, isLoading } = useAuth()

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary-500"></div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" />
  }

  return (
    <Layout>
      <Outlet />
    </Layout>
  )
}

function LoginRoute() {
  const { isAuthenticated } = useAuth()

  if (isAuthenticated) {
    return <Navigate to="/" />
  }

  return <Login />
}

const authenticatedRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: '_authenticated',
  component: AuthenticatedLayout,
})

// Public Routes
const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: LoginRoute,
})

const signupChoiceRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/signup',
  component: SignupChoice,
})

const signupTenantRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/signup/tenant',
  component: SignupTenant,
})

const verifyEmailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/verify-email',
  component: VerifyEmail,
})

const signupOperatorRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/signup/operator',
  component: OperatorInvitation,
})

const invitationTokenRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/invitation/$token',
  component: OperatorInvitation,
})

const invitationRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/invitation',
  component: OperatorInvitation,
})

const joinRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/join',
  component: JoinWithCode,
})

const recoveryRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/recovery',
  component: Recovery,
})

// Authenticated Routes
const indexRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/',
  component: Inbox,
})

const conversationRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/conversations/$id',
  component: ConversationDetail,
})

const contactsRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/contacts',
  component: Contacts,
})

const contactDetailRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/contacts/$id',
  component: ContactDetail,
})

const accountsRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/accounts',
  component: Accounts,
})

const teamRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/team',
  component: Team,
})

const invitationsManageRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/invitations',
  component: Team,
})

const botRulesRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/bot-rules',
  component: BotRules,
})

const uploadJobsRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/upload-jobs',
  component: UploadJobs,
})

const setupWizardRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/setup',
  component: SetupWizard,
})

const totpSettingsRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/account/totp',
  component: TotpSettings,
})

const routeTree = rootRoute.addChildren([
  loginRoute,
  signupChoiceRoute,
  signupTenantRoute,
  verifyEmailRoute,
  signupOperatorRoute,
  invitationTokenRoute,
  invitationRoute,
  joinRoute,
  recoveryRoute,
  authenticatedRoute.addChildren([
    indexRoute,
    conversationRoute,
    contactsRoute,
    contactDetailRoute,
    accountsRoute,
    teamRoute,
    invitationsManageRoute,
    botRulesRoute,
    uploadJobsRoute,
    setupWizardRoute,
    totpSettingsRoute,
  ]),
])

const router = createRouter({
  routeTree,
  basepath: '/dashboard',
  defaultPreload: 'intent',
})

// Register the router type for the module
// eslint-disable-next-line @typescript-eslint/no-empty-object-type
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

function App() {
  return (
    <AuthProvider>
      <RouterProvider router={router} />
    </AuthProvider>
  )
}

export default App
