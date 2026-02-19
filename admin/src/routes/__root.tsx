import { createRootRoute, Outlet, Link, useRouter } from '@tanstack/react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { TooltipProvider } from '@/components/ui/tooltip'
import { Toaster } from '@/components/ui/sonner'
import { Button } from '@/components/ui/button'
import '@/globals.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,
      retry: 1,
    },
  },
})

export const Route = createRootRoute({
  component: RootLayout,
  notFoundComponent: NotFound,
})

function NotFound() {
  const router = useRouter()

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-6 bg-background text-foreground">
      <div className="text-center">
        <p className="text-7xl font-bold tracking-tight">404</p>
        <p className="mt-2 text-lg text-muted-foreground">
          This page doesn't exist.
        </p>
      </div>
      <div className="flex gap-3">
        <Button variant="outline" onClick={() => router.history.back()}>
          Go back
        </Button>
        <Button asChild>
          <Link to="/">Dashboard</Link>
        </Button>
      </div>
    </div>
  )
}

function RootLayout() {
  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <Outlet />
        <Toaster />
      </TooltipProvider>
    </QueryClientProvider>
  )
}
