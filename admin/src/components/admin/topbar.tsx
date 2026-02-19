import { useNavigate, useMatches } from '@tanstack/react-router'
import { LogOut } from 'lucide-react'
import { useAuthContext } from '@/lib/auth'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

function getPageTitle(matches: ReturnType<typeof useMatches>): string {
  const lastMatch = matches[matches.length - 1]
  if (!lastMatch) {
    return 'Dashboard'
  }

  const path = lastMatch.pathname
  const segment = path.split('/').filter(Boolean).pop()
  if (!segment) {
    return 'Dashboard'
  }

  return segment.charAt(0).toUpperCase() + segment.slice(1)
}

export function Topbar() {
  const { user, logout } = useAuthContext()
  const navigate = useNavigate()
  const matches = useMatches()

  const pageTitle = getPageTitle(matches)

  const handleLogout = async () => {
    await logout()
    await navigate({ to: '/login' })
  }

  return (
    <header className="flex h-14 items-center justify-between border-b px-6">
      <div className="text-sm font-medium text-muted-foreground">
        {pageTitle}
      </div>

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <button
            type="button"
            className="flex items-center gap-2 rounded-full outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <Avatar className="h-8 w-8">
              <AvatarFallback className="text-xs">
                {user ? user.email.slice(0, 2).toUpperCase() : '??'}
              </AvatarFallback>
            </Avatar>
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-48">
          <DropdownMenuLabel className="font-normal text-sm text-muted-foreground">
            {user?.name ?? 'Unknown'}
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={handleLogout}>
            <LogOut className="mr-2 h-4 w-4" />
            Logout
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </header>
  )
}
