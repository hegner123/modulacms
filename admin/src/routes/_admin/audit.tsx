import { createFileRoute } from '@tanstack/react-router'
import { History } from 'lucide-react'
import { EmptyState } from '@/components/shared/empty-state'

export const Route = createFileRoute('/_admin/audit')({
  component: AuditLogPage,
})

function AuditLogPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Audit Log</h1>
        <p className="text-muted-foreground">
          Review a chronological record of all changes made in the admin panel.
        </p>
      </div>

      <EmptyState
        icon={History}
        title="Audit log not available"
        description="Audit log requires the activity/recent composed endpoint which is not yet implemented. This feature will be available in a future release."
      />
    </div>
  )
}
