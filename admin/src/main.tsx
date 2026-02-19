import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider, createRouter } from '@tanstack/react-router'
import { routeTree } from './routeTree.gen'

const router = createRouter({
  routeTree,
  defaultNotFoundComponent: () => {
    return (
      <div className="flex flex-col items-center justify-center gap-6 py-24 text-foreground">
        <div className="text-center">
          <p className="text-7xl font-bold tracking-tight">404</p>
          <p className="mt-2 text-lg text-muted-foreground">
            This page doesn't exist.
          </p>
        </div>
      </div>
    )
  },
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

const rootElement = document.getElementById('root')!
createRoot(rootElement).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
)
