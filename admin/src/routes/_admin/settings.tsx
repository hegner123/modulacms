import { createFileRoute } from '@tanstack/react-router'
import { Settings } from 'lucide-react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

export const Route = createFileRoute('/_admin/settings')({
  component: SettingsPage,
})

function SettingsPage() {
  const cmsUrl = import.meta.env.VITE_CMS_URL ?? 'http://localhost:8080'
  const isDev = import.meta.env.DEV

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-muted-foreground">
          Configure global CMS settings, integrations, and preferences.
        </p>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <Settings className="h-5 w-5 text-muted-foreground" />
            <div>
              <CardTitle>Configuration</CardTitle>
              <CardDescription>
                Current CMS configuration and environment details.
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex items-center justify-between rounded-lg border p-4">
              <div>
                <p className="text-sm font-medium">CMS URL</p>
                <p className="text-sm text-muted-foreground">
                  The backend API endpoint for this admin panel.
                </p>
              </div>
              <code className="rounded bg-muted px-2 py-1 font-mono text-sm">
                {cmsUrl}
              </code>
            </div>

            <div className="flex items-center justify-between rounded-lg border p-4">
              <div>
                <p className="text-sm font-medium">Environment</p>
                <p className="text-sm text-muted-foreground">
                  The current runtime environment.
                </p>
              </div>
              <Badge variant={isDev ? 'secondary' : 'default'}>
                {isDev ? 'Development' : 'Production'}
              </Badge>
            </div>

            <div className="rounded-lg border border-dashed p-4">
              <p className="text-sm text-muted-foreground">
                Additional settings will be available when the CMS exposes a settings API.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
