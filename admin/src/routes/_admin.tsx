import { createFileRoute, Outlet, Navigate } from '@tanstack/react-router'
import { AuthProvider, useAuthContext } from '@/lib/auth'
import { Sidebar } from '@/components/admin/sidebar'
import { Topbar } from '@/components/admin/topbar'
import { useState } from 'react'

export const Route = createFileRoute('/_admin')({
  component: AdminLayout,
})

function AdminLayout() {
  return (
    <AuthProvider>
      <AdminShell />
    </AuthProvider>
  )
}

function AdminShell() {
  const { user, isLoading } = useAuthContext()
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    )
  }

  if (!user) {
    return <Navigate to="/login" />
  }

  return (
    <div className="flex min-h-screen">
      <Sidebar
        collapsed={sidebarCollapsed}
        onToggle={() => setSidebarCollapsed((prev) => !prev)}
      />
      <div
        className="flex flex-1 flex-col transition-all duration-200"
        style={{ marginLeft: sidebarCollapsed ? 64 : 256 }}
      >
        <Topbar />
        <main className="flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
