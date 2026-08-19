import React, { useState, useEffect, useMemo } from 'react'
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
  ShieldCheck,
  Sparkles,
  UserCheck,
  GitBranch,
  Settings,
  ChevronDown,
  ChevronRight,
  PanelLeftClose,
  PanelLeftOpen,
} from 'lucide-react'

interface NavChild {
  name: string
  href: string
  icon: React.ElementType
  adminOnly?: boolean
}

interface NavItem {
  name: string
  href: string
  icon: React.ElementType
  children?: NavChild[]
}

const allNavigation: NavItem[] = [
  { name: 'Inbox', href: '/inbox', icon: MessageSquare },
  { name: 'Contacts', href: '/contacts', icon: Users },
  { name: 'Channels', href: '/accounts', icon: Smartphone },
  {
    name: 'Settings',
    href: '/settings',
    icon: Settings,
    children: [
      { name: 'Pipelines & Stages', href: '/settings/pipelines', icon: GitBranch, adminOnly: true },
      { name: 'Security & TOTP', href: '/account/totp', icon: ShieldCheck, adminOnly: true },
      { name: 'Team & Invites', href: '/team', icon: UserCheck, adminOnly: true },
      { name: 'Setup Wizard', href: '/setup', icon: Sparkles, adminOnly: true },
    ],
  },
]

const Layout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [collapsed, setCollapsed] = useState(() => {
    try {
      return localStorage.getItem('sidebar-collapsed') === 'true'
    } catch {
      return false
    }
  })
  const { user, tenantName, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()

  useEffect(() => {
    try {
      localStorage.setItem('sidebar-collapsed', String(collapsed))
    } catch {
      // ignore
    }
  }, [collapsed])

  const isAdmin = user?.role?.toUpperCase() === 'ADMIN'

  const navigation = useMemo(() => {
    if (isAdmin) return allNavigation
    return allNavigation.map((item) => {
      if (!item.children) return item
      const filtered = item.children.filter((child) => !child.adminOnly)
      if (filtered.length === 0) return null
      return { ...item, children: filtered }
    }).filter(Boolean) as NavItem[]
  }, [isAdmin])

  const activePath = location.pathname

  const isActive = (href: string) => {
    if (href === '/inbox') return activePath === '/inbox'
    return activePath === href || activePath.startsWith(href + '/')
  }

  const expandedGroups = useMemo(() => {
    return navigation
      .filter((item) => item.children && item.children.some((child) => isActive(child.href)))
      .map((item) => item.name)
  }, [activePath])

  const [openGroups, setOpenGroups] = useState<string[]>(expandedGroups)

  useEffect(() => {
    setOpenGroups((prev) => {
      const newSet = new Set([...prev, ...expandedGroups])
      return Array.from(newSet)
    })
  }, [expandedGroups])

  const handleLogout = async () => {
    try {
      await logout()
      navigate({ to: '/login' })
    } catch (error) {
      console.error('Logout failed:', error)
    }
  }

  const toggleGroup = (name: string) => {
    setOpenGroups((prev) => (prev.includes(name) ? prev.filter((n) => n !== name) : [...prev, name]))
  }

  const sidebarWidthClass = collapsed ? 'w-16' : 'w-56'
  const mainPaddingClass = collapsed ? 'lg:pl-16' : 'lg:pl-56'

  const renderNavItem = (item: NavItem) => {
    const Icon = item.icon
    const hasChildren = !!item.children
    const isGroupOpen = openGroups.includes(item.name)
    const anyChildActive = hasChildren && item.children!.some((child) => isActive(child.href))
    const itemActive = isActive(item.href) || anyChildActive

    return (
      <div key={item.name} className="relative">
        {hasChildren ? (
          <>
            <button
              type="button"
              onClick={() => toggleGroup(item.name)}
              className={`w-full flex items-center gap-2.5 px-2.5 py-1.5 text-left text-[13px] rounded-md transition-colors ${
                itemActive
                  ? 'bg-white/10 text-white font-semibold'
                  : 'text-gray-400 hover:text-gray-100 hover:bg-white/5'
              }`}
              title={item.name}
            >
              <Icon className={`h-4 w-4 shrink-0 ${itemActive ? 'text-primary-400' : 'text-gray-500'}`} />
              {!collapsed && (
                <>
                  <span className="flex-1">{item.name}</span>
                  {isGroupOpen ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                </>
              )}
            </button>
            {!collapsed && isGroupOpen && (
              <div className="mt-0.5 ml-6 space-y-0.5 border-l border-white/10 pl-2">
                {item.children!.map((child) => {
                  const ChildIcon = child.icon
                  const childActive = isActive(child.href)
                  return (
                    <Link
                      key={child.name}
                      to={child.href}
                      className={`w-full flex items-center gap-2.5 px-2.5 py-1.5 text-left text-[13px] rounded-md transition-colors ${
                        childActive
                          ? 'bg-white/10 text-white font-semibold'
                          : 'text-gray-400 hover:text-gray-100 hover:bg-white/5'
                      }`}
                      onClick={() => setSidebarOpen(false)}
                      title={child.name}
                    >
                      <ChildIcon className={`h-3.5 w-3.5 shrink-0 ${childActive ? 'text-primary-400' : 'text-gray-500'}`} />
                      <span>{child.name}</span>
                    </Link>
                  )
                })}
              </div>
            )}
          </>
        ) : (
          <Link
            to={item.href}
            className={`w-full flex items-center gap-2.5 px-2.5 py-1.5 text-left text-[13px] rounded-md transition-colors ${
              itemActive
                ? 'bg-white/10 text-white font-semibold'
                : 'text-gray-400 hover:text-gray-100 hover:bg-white/5'
            }`}
            onClick={() => setSidebarOpen(false)}
            title={item.name}
          >
            <Icon className={`h-4 w-4 shrink-0 ${itemActive ? 'text-primary-400' : 'text-gray-500'}`} />
            {!collapsed && <span>{item.name}</span>}
          </Link>
        )}
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-[#faf9f7]">
      {/* Mobile sidebar toggle button */}
      <button
        type="button"
        className="absolute top-3 left-3 z-20 p-1.5 rounded-md text-primary-600 focus:outline-none focus:ring-2 focus:ring-primary-500 lg:hidden"
        onClick={() => setSidebarOpen(true)}
      >
        <Menu className="h-5 w-5" />
      </button>

      {/* Sidebar (desktop) and Mobile sidebar */}
      <div
        className={`fixed inset-y-0 left-0 z-30 ${sidebarWidthClass} bg-[#1c1917] text-gray-400 border-r border-white/5 transform ${
          sidebarOpen ? 'translate-x-0' : '-translate-x-full'
        } transition-all duration-300 ease-in-out lg:translate-x-0`}
      >
        <div className="flex flex-col h-full">
          <div className="flex items-center justify-between px-3 py-3 border-b border-white/5">
            {!collapsed && (
              <div className="flex items-center space-x-2.5 min-w-0">
                <div className="h-7 w-7 rounded-lg bg-primary-500/90 flex items-center justify-center shadow-sm shrink-0">
                  <span className="text-white text-sm font-bold">w</span>
                </div>
                <span className="font-semibold text-sm text-gray-100 tracking-tight truncate">{tenantName || 'Dashboard'}</span>
              </div>
            )}
            {collapsed && (
              <div className="mx-auto">
                <div className="h-7 w-7 rounded-lg bg-primary-500/90 flex items-center justify-center shadow-sm">
                  <span className="text-white text-sm font-bold">w</span>
                </div>
              </div>
            )}
            <button
              type="button"
              className="lg:hidden p-1 rounded-md text-gray-500 hover:text-gray-200 hover:bg-white/5 focus:outline-none"
              onClick={() => setSidebarOpen(false)}
            >
              <X className="h-4 w-4" />
            </button>
          </div>

          <nav className="flex-1 px-2 py-3 overflow-y-auto">
            <div className="space-y-0.5">{navigation.map(renderNavItem)}</div>
          </nav>

          <div className="border-t border-white/5 p-3 space-y-2">
            {!collapsed && (
              <button
                type="button"
                onClick={() => setCollapsed(true)}
                className="hidden lg:flex w-full items-center gap-2 px-2.5 py-1.5 text-[13px] text-gray-400 hover:text-gray-100 hover:bg-white/5 rounded-md transition-colors"
                title="Collapse sidebar"
              >
                <PanelLeftClose className="h-4 w-4" />
                <span>Collapse</span>
              </button>
            )}
            {collapsed && (
              <button
                type="button"
                onClick={() => setCollapsed(false)}
                className="hidden lg:flex w-full justify-center p-1.5 text-gray-400 hover:text-gray-100 hover:bg-white/5 rounded-md transition-colors"
                title="Expand sidebar"
              >
                <PanelLeftOpen className="h-4 w-4" />
              </button>
            )}

            <Link
              to="/account/totp"
              className={`flex items-center ${collapsed ? 'justify-center' : 'space-x-2.5'} p-1.5 rounded-lg hover:bg-white/5 transition-colors`}
              title={user?.name || 'User'}
            >
              <div className="h-8 w-8 rounded-lg bg-white/10 text-gray-200 font-semibold flex items-center justify-center text-xs shrink-0">
                {user?.name?.charAt(0) || 'U'}
              </div>
              {!collapsed && (
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-gray-200 truncate">{user?.name || 'User'}</p>
                  <p className="text-xs text-gray-500 truncate">{user?.email || user?.whatsapp_number || 'Operator'}</p>
                </div>
              )}
            </Link>

            <div className="pt-2 border-t border-white/5">
              <Button
                variant="ghost"
                size="sm"
                className={`w-full text-xs text-gray-400 hover:text-red-400 hover:bg-white/5 ${collapsed ? 'justify-center px-1' : 'justify-start'}`}
                onClick={handleLogout}
                title="Sign out"
              >
                <LogOut className="h-3.5 w-3.5" />
                {!collapsed && <span className="ml-1.5">Sign out</span>}
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Main content area */}
      <div className={`${mainPaddingClass} min-h-screen flex flex-col transition-all duration-300`}>
        <header className="bg-white/90 backdrop-blur border-b border-gray-200 sticky top-0 z-10">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex items-center justify-between h-12">
            <div className="lg:hidden w-8" />
            <Link
              to="/account/totp"
              className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[11px] font-medium bg-emerald-50 text-emerald-700 border border-emerald-200 hover:bg-emerald-100 transition-colors"
            >
              <ShieldCheck className="h-3.5 w-3.5 text-emerald-600" />
              <span>TOTP Secured</span>
            </Link>
          </div>
        </header>

        <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-5 flex-1 w-full">{children}</main>
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
