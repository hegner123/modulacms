import { useState, type ReactNode } from 'react'
import { Link, useMatchRoute } from '@tanstack/react-router'
import type { LucideIcon } from 'lucide-react'
import {
  LayoutDashboard,
  FileText,
  Blocks,
  FormInput,
  Image,
  Users,
  Shield,
  Key,
  Globe,
  Puzzle,
  Upload,
  History,
  Settings,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipTrigger,
  TooltipContent,
} from '@/components/ui/tooltip'

type SidebarProps = {
  collapsed: boolean
  onToggle: () => void
}

function NavItem({
  to,
  icon: Icon,
  label,
  collapsed,
}: {
  to: string
  icon: LucideIcon
  label: string
  collapsed: boolean
}) {
  const linkContent = (
    <Link
      to={to}
      activeProps={{
        className: 'bg-sidebar-accent text-sidebar-accent-foreground',
      }}
      className={cn(
        'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-sidebar-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
        collapsed && 'justify-center px-2',
      )}
    >
      <Icon className="h-4 w-4 shrink-0" />
      {!collapsed && <span>{label}</span>}
    </Link>
  )

  if (collapsed) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>{linkContent}</TooltipTrigger>
        <TooltipContent side="right">{label}</TooltipContent>
      </Tooltip>
    )
  }

  return linkContent
}

function SectionHeader({
  label,
  collapsed,
}: {
  label: string
  collapsed: boolean
}) {
  if (collapsed) {
    return <div className="mx-auto my-2 h-px w-4 bg-sidebar-border" />
  }

  return (
    <div className="px-3 py-2 text-xs font-semibold uppercase tracking-wider text-sidebar-foreground/50">
      {label}
    </div>
  )
}

function NavGroup({
  to,
  icon: Icon,
  label,
  collapsed,
  defaultOpen = true,
  children,
}: {
  to: string
  icon: LucideIcon
  label: string
  collapsed: boolean
  defaultOpen?: boolean
  children: ReactNode
}) {
  const [open, setOpen] = useState(defaultOpen)
  const matchRoute = useMatchRoute()
  const isActive = !!matchRoute({ to, fuzzy: true })

  if (collapsed) {
    return (
      <>
        <NavItem to={to} icon={Icon} label={label} collapsed={collapsed} />
        {children}
      </>
    )
  }

  return (
    <div>
      <div
        className={cn(
          'group flex rounded-md text-sidebar-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
          isActive && 'bg-sidebar-accent text-sidebar-accent-foreground',
        )}
      >
        <Link
          to={to}
          className="flex flex-1 items-center gap-3 px-3 py-2 text-sm font-medium"
        >
          <Icon className="h-4 w-4 shrink-0" />
          <span>{label}</span>
        </Link>
        <button
          type="button"
          onClick={() => setOpen((prev) => !prev)}
          className="flex items-center py-2 px-1.5 text-current/50 group-hover:text-current"
        >
          <ChevronDown
            className={cn(
              'h-4 w-4 transition-transform duration-200',
              !open && '-rotate-90',
            )}
          />
        </button>
      </div>
      {open && <div className="mt-0.5 ml-4 space-y-0.5">{children}</div>}
    </div>
  )
}

export function Sidebar({ collapsed, onToggle }: SidebarProps) {
  return (
    <aside
      className={cn(
        'fixed left-0 top-0 z-30 flex h-screen flex-col border-r border-sidebar-border bg-sidebar-background transition-all duration-200',
        collapsed ? 'w-16' : 'w-64',
      )}
    >
      <div
        className={cn(
          'flex h-14 items-center border-b border-sidebar-border px-4',
          collapsed ? 'justify-center px-2' : 'justify-between',
        )}
      >
        {collapsed ? (
          <Button
            variant="ghost"
            size="icon"
            onClick={onToggle}
            className="text-sidebar-foreground"
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        ) : (
          <>
            <span className="text-lg font-bold text-sidebar-primary">
              Modula
            </span>
            <Button
              variant="ghost"
              size="icon"
              onClick={onToggle}
              className="text-sidebar-foreground"
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
          </>
        )}
      </div>

      <nav className="flex-1 space-y-1 overflow-y-auto p-2">
        <NavItem
          to="/"
          icon={LayoutDashboard}
          label="Dashboard"
          collapsed={collapsed}
        />
        <NavItem
          to="/content"
          icon={FileText}
          label="Content"
          collapsed={collapsed}
        />

        <SectionHeader label="Schema" collapsed={collapsed} />
        <NavItem
          to="/schema/datatypes"
          icon={Blocks}
          label="Datatypes"
          collapsed={collapsed}
        />
        <NavItem
          to="/schema/fields"
          icon={FormInput}
          label="Fields"
          collapsed={collapsed}
        />

        <NavItem
          to="/media"
          icon={Image}
          label="Media"
          collapsed={collapsed}
        />

        <NavGroup
          to="/users"
          icon={Users}
          label="Users"
          collapsed={collapsed}
        >
          <NavItem
            to="/users/roles"
            icon={Shield}
            label="Roles"
            collapsed={collapsed}
          />
          <NavItem
            to="/users/tokens"
            icon={Key}
            label="API Tokens"
            collapsed={collapsed}
          />
        </NavGroup>

        <SectionHeader label="Routes" collapsed={collapsed} />
        <NavItem
          to="/routes"
          icon={Globe}
          label="Routes"
          collapsed={collapsed}
        />

        <SectionHeader label="Plugins" collapsed={collapsed} />
        <NavItem
          to="/plugins"
          icon={Puzzle}
          label="Plugins"
          collapsed={collapsed}
        />

        <NavItem
          to="/import"
          icon={Upload}
          label="Import"
          collapsed={collapsed}
        />
        <NavItem
          to="/audit"
          icon={History}
          label="Audit Log"
          collapsed={collapsed}
        />
        <NavItem
          to="/settings"
          icon={Settings}
          label="Settings"
          collapsed={collapsed}
        />
      </nav>

    </aside>
  )
}
