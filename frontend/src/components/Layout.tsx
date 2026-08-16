import React, { useState } from 'react'
import { Link, useNavigate, useLocation } from '@tanstack/react-router'
import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import {
  Menu,
  X,
  LogOut,
  MessageSquare,
  Users,
  Smartphone,
  Bot as BotIcon,
  UploadCloud,
  ShieldCheck,
  Sparkles,
  UserCheck,
} from 'lucide-react'

const navigation = [
  { name: 'Inbox', href: '/', icon: MessageSquare },
  { name: 'Contacts', href: '/contacts', icon: Users },
  { name: 'Accounts', href: '/accounts', icon: Smartphone },
  { name: 'Team & Invites', href: '/team', icon: UserCheck },
  { name: 'Bot Rules', href: '/bot-rules', icon: BotIcon },
  { name: 'Upload Jobs', href: '/upload-jobs', icon: UploadCloud },
  { name: 'Setup Wizard', href: '/setup', icon: Sparkles },
  { name: 'Security & TOTP', href: '/account/totp', icon: ShieldCheck },
]

const Layout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()

  const handleLogout = async () => {
    try {
      await logout()
      navigate({ to: '/login' })
    } catch (error) {
      console.error('Logout failed:', error)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Mobile sidebar toggle button */}
      <button
        type="button"
        className="absolute top-4 left-4 z-20 p-2 rounded-md text-primary-600 focus:outline-none focus:ring-2 focus:ring-primary-500 lg:hidden"
        onClick={() => setSidebarOpen(true)}
      >
        <Menu className="h-6 w-6" />
      </button>

      {/* Sidebar (desktop) and Mobile sidebar */}
      <div className={`fixed inset-y-0 left-0 z-30 w-64 bg-white shadow-lg transform ${sidebarOpen ? 'translate-x-0' : '-translate-x-full'} transition-transform duration-300 ease-in-out lg:translate-x-0 lg:fixed lg:shadow-none`}>
        <div className="flex flex-col h-full">
          <div className="flex items-center justify-between p-4 border-b">
            <div className="flex items-center space-x-2">
              <div className="h-8 w-8 rounded-full bg-primary-600 flex items-center justify-center">
                <span className="text-white font-bold">W</span>
              </div>
              <span className="font-bold text-lg text-gray-900">WhatsApp Dashboard</span>
            </div>
            <button
              type="button"
              className="lg:hidden p-2 rounded-md text-gray-500 hover:text-gray-900 focus:outline-none focus:ring-2 focus:ring-primary-500"
              onClick={() => setSidebarOpen(false)}
            >
              <X className="h-6 w-6" />
            </button>
          </div>

          <nav className="flex-1 px-3 py-4 overflow-y-auto">
            <div className="space-y-1">
              {navigation.map((item) => {
                const Icon = item.icon
                const isActive = item.href === '/'
                  ? location.pathname === '/'
                  : location.pathname === item.href || location.pathname.startsWith(item.href + '/')
                return (
                  <Link
                    key={item.name}
                    to={item.href}
                    className={`w-full flex items-center px-3.5 py-2.5 text-left text-sm rounded-lg transition-colors ${
                      isActive
                        ? 'bg-primary-50 text-primary-700 font-semibold shadow-xs'
                        : 'text-gray-700 hover:bg-gray-100'
                    }`}
                    onClick={() => setSidebarOpen(false)}
                  >
                    <Icon className={`h-4.5 w-4.5 mr-3 ${isActive ? 'text-primary-600' : 'text-gray-400'}`} />
                    <span>{item.name}</span>
                  </Link>
                )
              })}
            </div>
          </nav>

          <div className="border-t p-4">
            <Link
              to="/account/totp"
              className="flex items-center space-x-3 p-1.5 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <div className="h-9 w-9 rounded-full bg-primary-100 text-primary-700 font-bold flex items-center justify-center text-sm">
                {user?.name?.charAt(0) || 'U'}
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-semibold text-gray-900 truncate">{user?.name || 'User'}</p>
                <p className="text-xs text-gray-500 truncate">{user?.email || user?.whatsapp_number || 'Operator'}</p>
              </div>
            </Link>

            <div className="mt-3 pt-2 border-t border-gray-100">
              <Button
                variant="ghost"
                size="sm"
                className="w-full justify-start text-xs text-gray-600 hover:text-red-600 hover:bg-red-50"
                onClick={handleLogout}
              >
                <LogOut className="h-3.5 w-3.5 mr-2" />
                Sign out
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Main content area */}
      <div className="lg:pl-64 min-h-screen flex flex-col">
        <header className="bg-white border-b sticky top-0 z-10">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex items-center justify-between h-16">
            <div className="flex items-center space-x-4 pl-8 lg:pl-0">
              <h1 className="text-lg sm:text-xl font-bold text-gray-900">
                WhatsApp Workspace
              </h1>
            </div>
            <div className="flex items-center space-x-3">
              <Link
                to="/account/totp"
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium bg-green-50 text-green-700 border border-green-200 hover:bg-green-100"
              >
                <ShieldCheck className="h-4 w-4 text-green-600" />
                <span>TOTP Secured</span>
              </Link>
            </div>
          </div>
        </header>

        <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 flex-1 w-full">
          {children}
        </main>
      </div>

      {/* Mobile sidebar backdrop */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black bg-opacity-50 z-20 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}
    </div>
  )
}

export default Layout
