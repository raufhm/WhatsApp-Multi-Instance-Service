import { RouterProvider } from '@tanstack/react-router'
import { Suspense, useEffect } from 'react'
import { AuthProvider } from '@/hooks/useAuth'
import { router } from '@/routes'

function App() {
  useEffect(() => {
    document.documentElement.classList.add('dark')
  }, [])

  return (
    <AuthProvider>
      <Suspense
        fallback={
          <div className="min-h-screen flex items-center justify-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600" />
          </div>
        }
      >
        <RouterProvider router={router} />
      </Suspense>
    </AuthProvider>
  )
}

export default App
