import { createRoute, Navigate } from '@tanstack/react-router'
import { lazy } from 'react'
import { rootRoute } from './root'

const Landing = lazy(() => import('@/pages/Landing'))
const SignupChoice = lazy(() => import('@/pages/SignupChoice'))
const SignupTenant = lazy(() => import('@/pages/SignupTenant'))
const VerifyEmail = lazy(() => import('@/pages/VerifyEmail'))
const OperatorInvitation = lazy(() => import('@/pages/OperatorInvitation'))
const JoinWithCode = lazy(() => import('@/pages/JoinWithCode'))
const Recovery = lazy(() => import('@/pages/Recovery'))

function LandingRoute() {
  return <Landing />
}

export const landingRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: LandingRoute,
})

export const signupChoiceRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/signup',
  component: SignupChoice,
})

export const signupTenantRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/signup/tenant',
  component: SignupTenant,
})

export const verifyEmailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/verify-email',
  component: VerifyEmail,
})

export const signupOperatorRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/signup/operator',
  component: OperatorInvitation,
})

export const invitationTokenRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/invitation/$token',
  component: OperatorInvitation,
})

export const invitationRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/invitation',
  component: OperatorInvitation,
})

export const joinRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/join',
  component: JoinWithCode,
})

export const recoveryRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/recovery',
  component: Recovery,
})

function LoginRedirect() {
  return <Navigate to="/" />
}

// Legacy login route - redirects to landing where login modal is available
export const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: LoginRedirect,
})
