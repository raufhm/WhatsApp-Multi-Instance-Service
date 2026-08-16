import React from 'react'
import { Link, useNavigate } from '@tanstack/react-router'
import { Card } from '@/components/ui/card'
import Button from '@/components/ui/button'
import { MessageCircle, Building2, UserPlus, ArrowRight, ShieldCheck } from 'lucide-react'

const SignupChoice: React.FC = () => {
  const navigate = useNavigate()

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md text-center">
        <div className="inline-flex items-center justify-center h-14 w-14 rounded-2xl bg-primary-600 text-white shadow-lg mb-4">
          <MessageCircle className="h-8 w-8" />
        </div>
        <h1 className="text-3xl font-extrabold text-gray-900 tracking-tight">
          Join WhatsApp Multi-Instance
        </h1>
        <p className="mt-2 text-sm text-gray-600 max-w-sm mx-auto">
          Choose how you would like to get started with our passwordless, TOTP-secured platform.
        </p>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-xl px-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Card 1: New Organization */}
          <Card className="p-6 flex flex-col justify-between hover:border-primary-500 hover:shadow-md transition-all border-2 border-transparent">
            <div>
              <div className="h-12 w-12 rounded-xl bg-primary-100 text-primary-700 flex items-center justify-center mb-4">
                <Building2 className="h-6 w-6" />
              </div>
              <h2 className="text-lg font-bold text-gray-900">New Organization</h2>
              <p className="mt-2 text-sm text-gray-600 leading-relaxed">
                Register a new company or tenant as an Administrator. Connect WhatsApp instances, set up
                bot rules, and manage operators.
              </p>
            </div>
            <div className="mt-6">
              <Button
                variant="primary"
                size="md"
                className="w-full justify-center group"
                onClick={() => navigate({ to: '/signup/tenant' })}
              >
                <span>Register Tenant</span>
                <ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-0.5 transition-transform" />
              </Button>
            </div>
          </Card>

          {/* Card 2: Operator Invitation */}
          <Card className="p-6 flex flex-col justify-between hover:border-primary-500 hover:shadow-md transition-all border-2 border-transparent">
            <div>
              <div className="h-12 w-12 rounded-xl bg-green-100 text-green-700 flex items-center justify-center mb-4">
                <UserPlus className="h-6 w-6" />
              </div>
              <h2 className="text-lg font-bold text-gray-900">Operator Invitation</h2>
              <p className="mt-2 text-sm text-gray-600 leading-relaxed">
                Invited by your team admin via WhatsApp or Email? Accept your invite, scan your TOTP QR
                code, and join your workspace.
              </p>
            </div>
            <div className="mt-6">
              <Button
                variant="secondary"
                size="md"
                className="w-full justify-center group"
                onClick={() => navigate({ to: '/join' })}
              >
                <span>Join with Code</span>
                <ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-0.5 transition-transform" />
              </Button>
            </div>
          </Card>
        </div>

        {/* Security Highlight */}
        <div className="mt-8 bg-primary-50/70 border border-primary-100 rounded-xl p-4 flex items-center gap-3 text-xs text-primary-900">
          <ShieldCheck className="h-5 w-5 text-primary-600 flex-shrink-0" />
          <span>
            <strong>Passwordless Security:</strong> We use Time-based One-Time Passwords (TOTP) and backup
            codes instead of vulnerable static passwords.
          </span>
        </div>

        {/* Sign in footer */}
        <div className="mt-6 text-center text-sm text-gray-600">
          Already have an account?{' '}
          <Link to="/login" className="font-semibold text-primary-600 hover:text-primary-500 underline">
            Sign in here
          </Link>
        </div>
      </div>
    </div>
  )
}

export default SignupChoice
