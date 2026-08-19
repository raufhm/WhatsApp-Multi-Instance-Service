import { createRouter } from '@tanstack/react-router'
import { rootRoute } from './root'
import {
  landingRoute,
  loginRoute,
  signupChoiceRoute,
  signupTenantRoute,
  verifyEmailRoute,
  signupOperatorRoute,
  invitationTokenRoute,
  invitationRoute,
  joinRoute,
  recoveryRoute,
} from './public'
import {
  authenticatedRoute,
  indexRoute,
  conversationRoute,
  contactsRoute,
  contactDetailRoute,
  settingsPipelinesRoute,
  accountsRoute,
  teamRoute,
  invitationsManageRoute,
  botRulesRoute,
  uploadJobsRoute,
  setupWizardRoute,
  totpSettingsRoute,
} from './authenticated'

const routeTree = rootRoute.addChildren([
  landingRoute,
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
    settingsPipelinesRoute,
    accountsRoute,
    teamRoute,
    invitationsManageRoute,
    botRulesRoute,
    uploadJobsRoute,
    setupWizardRoute,
    totpSettingsRoute,
  ]),
])

export const router = createRouter({
  routeTree,
  basepath: '/dashboard',
  defaultPreload: 'intent',
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
