import { createFileRoute, Link } from '@tanstack/react-router'
import {
  FileText,
  Blocks,
  FormInput,
  Globe,
  Image,
  Users,
  Shield,
  Key,
  Upload,
  History,
  ArrowRight,
} from 'lucide-react'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export const Route = createFileRoute('/_admin/')({
  component: DashboardPage,
})

type QuickLink = {
  title: string
  description: string
  icon: typeof FileText
  to: string
  section: string
}

const links: QuickLink[] = [
  {
    title: 'Content',
    description: 'Create and manage your content entries',
    icon: FileText,
    to: '/content',
    section: 'Content',
  },
  {
    title: 'Media',
    description: 'Upload and organize images and files',
    icon: Image,
    to: '/media',
    section: 'Content',
  },
  {
    title: 'Datatypes',
    description: 'Define content models and structures',
    icon: Blocks,
    to: '/schema/datatypes',
    section: 'Schema',
  },
  {
    title: 'Fields',
    description: 'Configure reusable field definitions',
    icon: FormInput,
    to: '/schema/fields',
    section: 'Schema',
  },
  {
    title: 'Routes',
    description: 'Set up URL routes for your content',
    icon: Globe,
    to: '/routes',
    section: 'Routes',
  },
  {
    title: 'Users',
    description: 'Manage user accounts and access',
    icon: Users,
    to: '/users',
    section: 'Users',
  },
  {
    title: 'Roles',
    description: 'Define roles and permissions',
    icon: Shield,
    to: '/users/roles',
    section: 'Users',
  },
  {
    title: 'API Tokens',
    description: 'Generate and manage API tokens',
    icon: Key,
    to: '/users/tokens',
    section: 'Users',
  },
  {
    title: 'Import',
    description: 'Import content from external sources',
    icon: Upload,
    to: '/import',
    section: 'Tools',
  },
  {
    title: 'Audit Log',
    description: 'Review activity and change history',
    icon: History,
    to: '/audit',
    section: 'Tools',
  },
]

const sections = ['Content', 'Schema', 'Routes', 'Users', 'Tools'] as const

function DashboardPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">Get Started</h1>
        <p className="mt-1 text-muted-foreground">
          Jump into any area of your CMS.
        </p>
      </div>

      {sections.map((section) => {
        const sectionLinks = links.filter((l) => l.section === section)
        if (sectionLinks.length === 0) return null
        return (
          <div key={section} className="space-y-3">
            <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
              {section}
            </h2>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {sectionLinks.map((link) => (
                <Link key={link.to} to={link.to}>
                  <Card className="group transition-colors hover:border-primary/50 hover:bg-muted/50">
                    <CardHeader className="flex flex-row items-center gap-3 pb-1">
                      <link.icon className="h-5 w-5 text-muted-foreground group-hover:text-primary" />
                      <CardTitle className="text-sm font-medium">
                        {link.title}
                      </CardTitle>
                      <ArrowRight className="ml-auto h-4 w-4 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100" />
                    </CardHeader>
                    <CardContent>
                      <p className="text-sm text-muted-foreground">
                        {link.description}
                      </p>
                    </CardContent>
                  </Card>
                </Link>
              ))}
            </div>
          </div>
        )
      })}
    </div>
  )
}
