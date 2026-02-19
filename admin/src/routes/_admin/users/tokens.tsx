import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import type { ColumnDef } from '@tanstack/react-table'
import type { Token } from '@modulacms/admin-sdk'
import { Key, Plus, Trash2 } from 'lucide-react'
import { DataTable } from '@/components/data-table/data-table'
import { DataTableColumnHeader } from '@/components/data-table/column-header'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { EmptyState } from '@/components/shared/empty-state'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useTokens, useCreateToken, useDeleteToken } from '@/queries/users'
import { formatDateTime } from '@/lib/utils'

export const Route = createFileRoute('/_admin/users/tokens')({
  component: ApiTokensPage,
})

function maskToken(token: string): string {
  if (token.length <= 8) return token
  return token.slice(0, 8) + '...'
}

function ApiTokensPage() {
  const { data: tokens, isLoading } = useTokens()
  const createToken = useCreateToken()
  const deleteToken = useDeleteToken()

  const [createOpen, setCreateOpen] = useState(false)
  const [createTokenType, setCreateTokenType] = useState('')
  const [createTokenValue, setCreateTokenValue] = useState('')
  const [createExpiresAt, setCreateExpiresAt] = useState('')

  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<Token | null>(null)

  function handleCreate() {
    const now = new Date().toISOString()
    createToken.mutate(
      {
        user_id: null,
        token_type: createTokenType,
        token: createTokenValue,
        issued_at: now,
        expires_at: createExpiresAt
          ? new Date(createExpiresAt).toISOString()
          : new Date(Date.now() + 365 * 24 * 60 * 60 * 1000).toISOString(),
        revoked: false,
      },
      {
        onSuccess: () => {
          setCreateOpen(false)
          setCreateTokenType('')
          setCreateTokenValue('')
          setCreateExpiresAt('')
        },
      },
    )
  }

  function handleRevoke() {
    if (!deleteTarget) return
    deleteToken.mutate(deleteTarget.id, {
      onSuccess: () => {
        setDeleteOpen(false)
        setDeleteTarget(null)
      },
    })
  }

  const columns: ColumnDef<Token>[] = [
    {
      accessorKey: 'token',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Token" />
      ),
      cell: ({ row }) => (
        <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-sm">
          {maskToken(row.original.token)}
        </code>
      ),
    },
    {
      accessorKey: 'token_type',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Type" />
      ),
    },
    {
      accessorKey: 'issued_at',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Issued" />
      ),
      cell: ({ row }) => formatDateTime(row.original.issued_at),
    },
    {
      accessorKey: 'expires_at',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Expires" />
      ),
      cell: ({ row }) => formatDateTime(row.original.expires_at),
    },
    {
      accessorKey: 'revoked',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Status" />
      ),
      cell: ({ row }) => (
        <Badge variant={row.original.revoked ? 'destructive' : 'default'}>
          {row.original.revoked ? 'Revoked' : 'Active'}
        </Badge>
      ),
    },
    {
      id: 'actions',
      enableHiding: false,
      cell: ({ row }) => {
        const token = row.original
        return (
          <Button
            variant="ghost"
            size="sm"
            className="text-destructive hover:text-destructive"
            onClick={() => {
              setDeleteTarget(token)
              setDeleteOpen(true)
            }}
          >
            <Trash2 className="mr-2 h-4 w-4" />
            Revoke
          </Button>
        )
      },
    },
  ]

  if (isLoading) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">API Tokens</h1>
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">API Tokens</h1>
          <p className="text-muted-foreground">
            Create and manage API tokens for programmatic access to the CMS.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Token
        </Button>
      </div>

      {tokens && tokens.length > 0 ? (
        <DataTable
          columns={columns}
          data={tokens}
          searchKey="token_type"
          searchPlaceholder="Search by token type..."
        />
      ) : (
        <EmptyState
          icon={Key}
          title="No tokens yet"
          description="Create an API token for programmatic CMS access."
          action={
            <Button onClick={() => setCreateOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Create Token
            </Button>
          }
        />
      )}

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Token</DialogTitle>
            <DialogDescription>
              Generate a new API token for programmatic access.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="create-token-type">Token Type</Label>
              <Input
                id="create-token-type"
                placeholder="e.g. access, api-key"
                value={createTokenType}
                onChange={(e) => setCreateTokenType(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="create-token-value">Token Value</Label>
              <Input
                id="create-token-value"
                placeholder="Token string"
                value={createTokenValue}
                onChange={(e) => setCreateTokenValue(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="create-expires-at">Expires At</Label>
              <Input
                id="create-expires-at"
                type="datetime-local"
                value={createExpiresAt}
                onChange={(e) => setCreateExpiresAt(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">
                Leave empty for a 1-year expiration.
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setCreateOpen(false)}
              disabled={createToken.isPending}
            >
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={
                !createTokenType.trim() ||
                !createTokenValue.trim() ||
                createToken.isPending
              }
            >
              {createToken.isPending ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Revoke Token"
        description="Are you sure you want to revoke this token? Any applications using it will lose access."
        onConfirm={handleRevoke}
        loading={deleteToken.isPending}
        variant="destructive"
      />
    </div>
  )
}
