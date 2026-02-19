import { createFileRoute, Outlet } from '@tanstack/react-router'
import { AuthProvider } from '@/lib/auth'

export const Route = createFileRoute('/_auth')({
  component: AuthLayout,
})

function AuthLayout() {
  return (
    <AuthProvider>
      <div className="flex min-h-screen items-center justify-center bg-muted/40 p-4">
        <div className="w-full max-w-md">
          <Outlet />
        </div>
      </div>
    </AuthProvider>
  )
}
