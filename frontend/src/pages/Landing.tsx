import React, { useState, useEffect } from 'react'
import { Link, useNavigate } from '@tanstack/react-router'
import { useAuth } from '@/hooks/useAuth'
import {
  QrCode,
  Users,
  MessageSquare,
  Bot,
  Check,
  ArrowRight,
  Smartphone,
  Clock,
  TrendingUp,
  Building2,
  Mail,
  Phone,
} from 'lucide-react'
import Button from '@/components/ui/button'
import Login from './Login'
import { Modal } from '@/components/ui/Modal'

const Landing: React.FC = () => {
  const [showLoginModal, setShowLoginModal] = useState(false)
  const { isAuthenticated, isLoading } = useAuth()
  const navigate = useNavigate()

  // Redirect authenticated users to dashboard inbox
  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      navigate({ to: '/inbox' } as any)
    }
  }, [isAuthenticated, isLoading, navigate])

  // Don't render the landing page while checking auth — prevents flash
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600" />
      </div>
    )
  }

  // Don't render if authenticated (redirect is happening)
  if (isAuthenticated) {
    return null
  }

  return (
    <>
      {/* Navigation */}
      <nav className="fixed top-0 left-0 right-0 z-40 bg-white/95 backdrop-blur-sm border-b border-gray-100">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            {/* Logo */}
            <div className="flex items-center space-x-3">
              <div className="h-10 w-10 rounded-2xl bg-primary-600 flex items-center justify-center shadow-md">
                <span className="text-white text-xl font-bold">w</span>
              </div>
              <div>
                <span className="text-xl font-bold tracking-tight text-gray-900">whops</span>
                <span className="block text-xs text-gray-500 -mt-1">WhatsApp for your team</span>
              </div>
            </div>

            {/* CTA */}
            <div className="flex items-center space-x-4">
              <button
                onClick={() => setShowLoginModal(true)}
                className="text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors"
              >
                Sign In
              </button>
              <Link to="/signup">
                <Button variant="primary" size="sm">
                  Get Started
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="pt-32 pb-20 px-4 sm:px-6 lg:px-8 bg-gradient-to-br from-gray-50 via-white to-primary-50/30">
        <div className="max-w-7xl mx-auto">
          <div className="grid lg:grid-cols-2 gap-12 items-center">
            {/* Left: Copy */}
            <div className="space-y-8">
              <div>
                <h1 className="text-4xl sm:text-5xl lg:text-6xl font-extrabold text-gray-900 leading-tight tracking-tight">
                  Turn Your WhatsApp Into A{' '}
                  <span className="text-primary-600">Team Customer Hub</span>
                </h1>
                <p className="mt-6 text-lg sm:text-xl text-gray-600 leading-relaxed max-w-xl">
                  Share your personal or Business WhatsApp across your entire team. 
                  Track conversations, add notes, set follow-ups — all without Meta Business API complexity.
                </p>
              </div>

              {/* CTAs */}
              <div className="flex flex-col sm:flex-row gap-4">
                <Link to="/signup">
                  <Button variant="primary" size="xl" className="w-full sm:w-auto justify-center">
                    Get Started
                    <ArrowRight className="ml-2 h-5 w-5" />
                  </Button>
                </Link>
                <Button
                  onClick={() => setShowLoginModal(true)}
                  variant="secondary"
                  size="xl"
                  className="w-full sm:w-auto border-2 border-gray-200 text-gray-700"
                >
                  Sign In
                </Button>
              </div>

              {/* Trust indicators */}
              <div className="flex items-center space-x-6 text-sm text-gray-500">
                <div className="flex items-center space-x-2">
                  <Check className="h-5 w-5 text-green-500" />
                  <span>No credit card required</span>
                </div>
                <div className="flex items-center space-x-2">
                  <Check className="h-5 w-5 text-green-500" />
                  <span>Set up in 2 minutes</span>
                </div>
              </div>
            </div>

            {/* Right: Visual */}
            <div className="relative">
              <div className="bg-gradient-to-br from-primary-500 to-primary-700 rounded-3xl p-8 shadow-2xl">
                <div className="bg-white rounded-2xl p-6 shadow-lg">
                  {/* Mockup illustration */}
                  <div className="space-y-4">
                    <div className="flex items-center space-x-3 pb-4 border-b border-gray-100">
                      <div className="h-12 w-12 rounded-full bg-primary-100 flex items-center justify-center">
                        <Smartphone className="h-6 w-6 text-primary-600" />
                      </div>
                      <div>
                        <p className="font-semibold text-gray-900">Scan to Connect</p>
                        <p className="text-sm text-gray-500">One QR code, whole team ready</p>
                      </div>
                    </div>

                    <div className="space-y-3">
                      <div className="flex items-start space-x-3 p-3 bg-gray-50 rounded-lg">
                        <QrCode className="h-5 w-5 text-primary-600 mt-0.5" />
                        <div>
                          <p className="font-medium text-gray-900">Single scan setup</p>
                          <p className="text-sm text-gray-500">Connect your WhatsApp once</p>
                        </div>
                      </div>
                      <div className="flex items-start space-x-3 p-3 bg-gray-50 rounded-lg">
                        <Users className="h-5 w-5 text-primary-600 mt-0.5" />
                        <div>
                          <p className="font-medium text-gray-900">Instant team access</p>
                          <p className="text-sm text-gray-500">Multiple members, one number</p>
                        </div>
                      </div>
                      <div className="flex items-start space-x-3 p-3 bg-gray-50 rounded-lg">
                        <MessageSquare className="h-5 w-5 text-primary-600 mt-0.5" />
                        <div>
                          <p className="font-medium text-gray-900">Built-in CRM</p>
                          <p className="text-sm text-gray-500">Notes, tags, follow-ups included</p>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Decorative elements */}
              <div className="absolute -top-4 -right-4 w-24 h-24 bg-yellow-400/20 rounded-full blur-2xl" />
              <div className="absolute -bottom-4 -left-4 w-32 h-32 bg-primary-400/20 rounded-full blur-2xl" />
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 px-4 sm:px-6 lg:px-8 bg-white">
        <div className="max-w-7xl mx-auto">
          <div className="text-center max-w-3xl mx-auto mb-16">
            <h2 className="text-3xl sm:text-4xl font-extrabold text-gray-900 tracking-tight">
              Everything you need to manage WhatsApp as a team
            </h2>
            <p className="mt-4 text-lg text-gray-600">
              No developer required. No API keys. No monthly per-seat fees. Just scan and start collaborating.
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
            {/* Feature 1 */}
            <div className="group p-6 bg-gray-50 rounded-2xl hover:bg-primary-50 transition-colors duration-200">
              <div className="h-12 w-12 bg-primary-100 rounded-xl flex items-center justify-center mb-4 group-hover:bg-primary-600 transition-colors">
                <QrCode className="h-6 w-6 text-primary-600 group-hover:text-white transition-colors" />
              </div>
              <h3 className="text-lg font-bold text-gray-900 mb-2">One Scan, Whole Team</h3>
              <p className="text-gray-600 text-sm leading-relaxed">
                Connect your WhatsApp once via QR code. Every team member instantly accesses the same number from their own account.
              </p>
            </div>

            {/* Feature 2 */}
            <div className="group p-6 bg-gray-50 rounded-2xl hover:bg-primary-50 transition-colors duration-200">
              <div className="h-12 w-12 bg-primary-100 rounded-xl flex items-center justify-center mb-4 group-hover:bg-primary-600 transition-colors">
                <MessageSquare className="h-6 w-6 text-primary-600 group-hover:text-white transition-colors" />
              </div>
              <h3 className="text-lg font-bold text-gray-900 mb-2">Built-In CRM</h3>
              <p className="text-gray-600 text-sm leading-relaxed">
                Customer profiles, full conversation history, internal notes, and follow-up reminders — no separate software needed.
              </p>
            </div>

            {/* Feature 3 */}
            <div className="group p-6 bg-gray-50 rounded-2xl hover:bg-primary-50 transition-colors duration-200">
              <div className="h-12 w-12 bg-primary-100 rounded-xl flex items-center justify-center mb-4 group-hover:bg-primary-600 transition-colors">
                <Bot className="h-6 w-6 text-primary-600 group-hover:text-white transition-colors" />
              </div>
              <h3 className="text-lg font-bold text-gray-900 mb-2">Automated Responses</h3>
              <p className="text-gray-600 text-sm leading-relaxed">
                Set simple rules to auto-reply to common questions. Free up your team for conversations that matter.
              </p>
            </div>

            {/* Feature 4 */}
            <div className="group p-6 bg-gray-50 rounded-2xl hover:bg-primary-50 transition-colors duration-200">
              <div className="h-12 w-12 bg-primary-100 rounded-xl flex items-center justify-center mb-4 group-hover:bg-primary-600 transition-colors">
                <Clock className="h-6 w-6 text-primary-600 group-hover:text-white transition-colors" />
              </div>
              <h3 className="text-lg font-bold text-gray-900 mb-2">Never Miss a Message</h3>
              <p className="text-gray-600 text-sm leading-relaxed">
                Assign conversations, track response times, and ensure every customer gets a reply — even after hours.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Benefits / Use Cases Section */}
      <section className="py-20 px-4 sm:px-6 lg:px-8 bg-gray-50">
        <div className="max-w-7xl mx-auto">
          <div className="text-center max-w-3xl mx-auto mb-16">
            <h2 className="text-3xl sm:text-4xl font-extrabold text-gray-900 tracking-tight">
              Built for small businesses like yours
            </h2>
            <p className="mt-4 text-lg text-gray-600">
              See how teams use whops to simplify their customer communication
            </p>
          </div>

          <div className="grid md:grid-cols-2 gap-8">
            {/* Use Case 1 */}
            <div className="bg-white p-8 rounded-2xl shadow-sm border border-gray-100">
              <div className="flex items-start space-x-4">
                <div className="h-12 w-12 bg-green-100 rounded-xl flex items-center justify-center flex-shrink-0">
                  <Building2 className="h-6 w-6 text-green-600" />
                </div>
                <div>
                  <h3 className="text-xl font-bold text-gray-900 mb-2">E-commerce Stores</h3>
                  <p className="text-gray-600 leading-relaxed mb-4">
                    Customers message your business WhatsApp from multiple devices. With whops, your support team can respond together from a shared inbox, avoiding duplicate replies and missed messages.
                  </p>
                  <ul className="space-y-2">
                    <li className="flex items-start space-x-2 text-sm text-gray-600">
                      <Check className="h-5 w-5 text-green-500 flex-shrink-0 mt-0.5" />
                      <span>Track order inquiries and delivery updates</span>
                    </li>
                    <li className="flex items-start space-x-2 text-sm text-gray-600">
                      <Check className="h-5 w-5 text-green-500 flex-shrink-0 mt-0.5" />
                      <span>Add product notes directly to customer profiles</span>
                    </li>
                  </ul>
                </div>
              </div>
            </div>

            {/* Use Case 2 */}
            <div className="bg-white p-8 rounded-2xl shadow-sm border border-gray-100">
              <div className="flex items-start space-x-4">
                <div className="h-12 w-12 bg-blue-100 rounded-xl flex items-center justify-center flex-shrink-0">
                  <Phone className="h-6 w-6 text-blue-600" />
                </div>
                <div>
                  <h3 className="text-xl font-bold text-gray-900 mb-2">Service Businesses</h3>
                  <p className="text-gray-600 leading-relaxed mb-4">
                    Coordinating appointments and follow-ups becomes effortless. Your front desk, technicians, and managers all stay synchronized on customer conversations.
                  </p>
                  <ul className="space-y-2">
                    <li className="flex items-start space-x-2 text-sm text-gray-600">
                      <Check className="h-5 w-5 text-green-500 flex-shrink-0 mt-0.5" />
                      <span>Schedule callbacks with built-in reminders</span>
                    </li>
                    <li className="flex items-start space-x-2 text-sm text-gray-600">
                      <Check className="h-5 w-5 text-green-500 flex-shrink-0 mt-0.5" />
                      <span>Share service notes across team members</span>
                    </li>
                  </ul>
                </div>
              </div>
            </div>

            {/* Use Case 3 */}
            <div className="bg-white p-8 rounded-2xl shadow-sm border border-gray-100">
              <div className="flex items-start space-x-4">
                <div className="h-12 w-12 bg-purple-100 rounded-xl flex items-center justify-center flex-shrink-0">
                  <Mail className="h-6 w-6 text-purple-600" />
                </div>
                <div>
                  <h3 className="text-xl font-bold text-gray-900 mb-2">Support Teams</h3>
                  <p className="text-gray-600 leading-relaxed mb-4">
                    Stop switching between personal phones and trying to share accounts. Centralize all customer support in one place with proper handoff tools.
                  </p>
                  <ul className="space-y-2">
                    <li className="flex items-start space-x-2 text-sm text-gray-600">
                      <Check className="h-5 w-5 text-green-500 flex-shrink-0 mt-0.5" />
                      <span>Assign tickets to specific team members</span>
                    </li>
                    <li className="flex items-start space-x-2 text-sm text-gray-600">
                      <Check className="h-5 w-5 text-green-500 flex-shrink-0 mt-0.5" />
                      <span>Internal notes visible only to your team</span>
                    </li>
                  </ul>
                </div>
              </div>
            </div>

            {/* Use Case 4 */}
            <div className="bg-white p-8 rounded-2xl shadow-sm border border-gray-100">
              <div className="flex items-start space-x-4">
                <div className="h-12 w-12 bg-orange-100 rounded-xl flex items-center justify-center flex-shrink-0">
                  <TrendingUp className="h-6 w-6 text-orange-600" />
                </div>
                <div>
                  <h3 className="text-xl font-bold text-gray-900 mb-2">Sales Teams</h3>
                  <p className="text-gray-600 leading-relaxed mb-4">
                    Nurture leads through WhatsApp without expensive CRM software. Track where each prospect is in your sales pipeline directly from conversations.
                  </p>
                  <ul className="space-y-2">
                    <li className="flex items-start space-x-2 text-sm text-gray-600">
                      <Check className="h-5 w-5 text-green-500 flex-shrink-0 mt-0.5" />
                      <span>Tag leads by interest level or product</span>
                    </li>
                    <li className="flex items-start space-x-2 text-sm text-gray-600">
                      <Check className="h-5 w-5 text-green-500 flex-shrink-0 mt-0.5" />
                      <span>Set follow-up reminders for hot leads</span>
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Comparison Section */}
      <section className="py-20 px-4 sm:px-6 lg:px-8 bg-white">
        <div className="max-w-4xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-3xl sm:text-4xl font-extrabold text-gray-900 tracking-tight">
              Why choose whops over Meta Business API?
            </h2>
            <p className="mt-4 text-lg text-gray-600">
              Skip the complexity. Start in minutes, not weeks.
            </p>
          </div>

          <div className="grid md:grid-cols-2 gap-8">
            {/* Meta Business API */}
            <div className="p-6 bg-gray-50 rounded-2xl border-2 border-gray-200">
              <h3 className="text-lg font-bold text-gray-700 mb-4">Meta Business API</h3>
              <ul className="space-y-3">
                <li className="flex items-start space-x-2 text-sm text-gray-600">
                  <span className="text-red-500 font-bold">×</span>
                  <span>Requires business verification & documentation</span>
                </li>
                <li className="flex items-start space-x-2 text-sm text-gray-600">
                  <span className="text-red-500 font-bold">×</span>
                  <span>Needs developer to set up webhooks & APIs</span>
                </li>
                <li className="flex items-start space-x-2 text-sm text-gray-600">
                  <span className="text-red-500 font-bold">×</span>
                  <span>Per-conversation pricing adds up quickly</span>
                </li>
                <li className="flex items-start space-x-2 text-sm text-gray-600">
                  <span className="text-red-500 font-bold">×</span>
                  <span>Template messages require approval</span>
                </li>
                <li className="flex items-start space-x-2 text-sm text-gray-600">
                  <span className="text-red-500 font-bold">×</span>
                  <span>Setup takes weeks of back-and-forth</span>
                </li>
              </ul>
            </div>

            {/* whops */}
            <div className="p-6 bg-primary-50 rounded-2xl border-2 border-primary-200 relative overflow-hidden">
              <div className="absolute top-0 right-0 bg-primary-600 text-white text-xs font-bold px-3 py-1 rounded-bl-lg">
                SIMPLER
              </div>
              <h3 className="text-lg font-bold text-primary-900 mb-4">whops</h3>
              <ul className="space-y-3">
                <li className="flex items-start space-x-2 text-sm text-primary-900">
                  <Check className="h-5 w-5 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>Use your existing WhatsApp immediately</span>
                </li>
                <li className="flex items-start space-x-2 text-sm text-primary-900">
                  <Check className="h-5 w-5 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>No developer needed — scan QR and done</span>
                </li>
                <li className="flex items-start space-x-2 text-sm text-primary-900">
                  <Check className="h-5 w-5 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>Flat fee, no per-message charges</span>
                </li>
                <li className="flex items-start space-x-2 text-sm text-primary-900">
                  <Check className="h-5 w-5 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>Send any message, anytime</span>
                </li>
                <li className="flex items-start space-x-2 text-sm text-primary-900">
                  <Check className="h-5 w-5 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>Ready in under 2 minutes</span>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* ToS Disclaimer Section */}
      <section className="py-12 px-4 sm:px-6 lg:px-8 bg-amber-50 border-y border-amber-200">
        <div className="max-w-4xl mx-auto">
          <div className="bg-white rounded-xl p-6 border border-amber-200">
            <h3 className="text-lg font-bold text-gray-900 mb-3">Transparency first</h3>
            <p className="text-sm text-gray-700 leading-relaxed mb-3">
              Using linked WhatsApp accounts may violate WhatsApp's Terms of Service. We don't guarantee "ban-free" operation, but thousands of small businesses use similar solutions daily. By signing up, you acknowledge this risk and agree to use at your own discretion.
            </p>
            <p className="text-xs text-gray-500">
              Not affiliated with or endorsed by Meta/WhatsApp.
            </p>
          </div>
        </div>
      </section>

      {/* Final CTA Section */}
      <section className="py-20 px-4 sm:px-6 lg:px-8 bg-gradient-to-br from-primary-700 via-primary-800 to-primary-950 text-white">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="text-3xl sm:text-4xl font-extrabold tracking-tight mb-6">
            Ready to transform your WhatsApp into a team hub?
          </h2>
          <p className="text-lg text-primary-100 mb-8 max-w-2xl mx-auto">
            Join small businesses already using whops to improve customer service, boost sales, and streamline team collaboration.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link to="/signup">
              <Button variant="primary" size="xl" className="bg-white text-primary-700 hover:bg-primary-50 border-transparent">
                Get Started
                <ArrowRight className="ml-2 h-5 w-5" />
              </Button>
            </Link>
            <Button
              onClick={() => setShowLoginModal(true)}
              variant="secondary"
              size="xl"
              className="border-2 border-white/30 text-white bg-transparent hover:bg-white/10"
            >
              Sign In
            </Button>
          </div>
          <p className="mt-6 text-sm text-primary-200/80">
            No credit card • 2-minute setup • Cancel anytime
          </p>
        </div>
      </section>

      {/* Footer */}
      <footer className="py-12 px-4 sm:px-6 lg:px-8 bg-gray-900 text-gray-400">
        <div className="max-w-7xl mx-auto">
          <div className="grid md:grid-cols-4 gap-8 mb-8">
            {/* Brand */}
            <div className="col-span-2 md:col-span-1">
              <div className="flex items-center space-x-3 mb-4">
                <div className="h-10 w-10 rounded-2xl bg-primary-600 flex items-center justify-center">
                  <span className="text-white text-xl font-bold">w</span>
                </div>
                <span className="text-xl font-bold text-white">whops</span>
              </div>
              <p className="text-sm leading-relaxed">
                WhatsApp for your team. Simple, powerful, built-in CRM.
              </p>
            </div>

            {/* Product */}
            <div>
              <h4 className="font-semibold text-white mb-4">Product</h4>
              <ul className="space-y-2 text-sm">
                <li><a href="#" className="hover:text-white transition-colors">Features</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Pricing</a></li>
                <li><a href="#" className="hover:text-white transition-colors">FAQ</a></li>
              </ul>
            </div>

            {/* Legal */}
            <div>
              <h4 className="font-semibold text-white mb-4">Legal</h4>
              <ul className="space-y-2 text-sm">
                <li><a href="#" className="hover:text-white transition-colors">Privacy Policy</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Terms of Service</a></li>
                <li><a href="#" className="hover:text-white transition-colors">ToS Disclaimer</a></li>
              </ul>
            </div>

            {/* Contact */}
            <div>
              <h4 className="font-semibold text-white mb-4">Contact</h4>
              <ul className="space-y-2 text-sm">
                <li><a href="mailto:support@whops.io" className="hover:text-white transition-colors">support@whops.io</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Documentation</a></li>
              </ul>
            </div>
          </div>

          <div className="pt-8 border-t border-gray-800 flex flex-col sm:flex-row justify-between items-center gap-4">
            <p className="text-sm">© {new Date().getFullYear()} whops. All rights reserved.</p>
            <button
              onClick={() => setShowLoginModal(true)}
              className="text-sm text-gray-400 hover:text-white transition-colors"
            >
              Sign In
            </button>
          </div>
        </div>
      </footer>

      {/* Login Modal */}
      <Modal isOpen={showLoginModal} onClose={() => setShowLoginModal(false)} title="Welcome Back">
        <Login close={() => setShowLoginModal(false)} />
      </Modal>
    </>
  )
}

export default Landing
