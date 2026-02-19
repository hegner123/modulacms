import { useState } from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { useForm } from 'react-hook-form'
import type { Email, SshKeyListItem } from '@modulacms/admin-sdk'
import { ArrowLeft, Key, Plus, Trash2 } from 'lucide-react'
import { useUser, useUpdateUser, useSshKeys, useCreateSshKey, useDeleteSshKey } from '@/queries/users'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

export const Route = createFileRoute('/_admin/users/$id')({
  component: UserDetailPage,
})

type EditFormValues = {
  username: string
  name: string
  email: string
  role: string
}

function UserDetailPage() {
  const { id } = Route.useParams()
  const { data: user, isLoading } = useUser(id)
  const updateUser = useUpdateUser()

  const { data: sshKeys } = useSshKeys()
  const createSshKey = useCreateSshKey()
  const deleteSshKey = useDeleteSshKey()

  const [newPassword, setNewPassword] = useState('')
  const [passwordSaving, setPasswordSaving] = useState(false)

  const [addKeyOpen, setAddKeyOpen] = useState(false)
  const [keyLabel, setKeyLabel] = useState('')
  const [keyValue, setKeyValue] = useState('')
  const [deleteKeyOpen, setDeleteKeyOpen] = useState(false)
  const [deleteKeyTarget, setDeleteKeyTarget] = useState<SshKeyListItem | null>(null)

  const {
    register,
    handleSubmit,
    formState: { isDirty },
    reset,
  } = useForm<EditFormValues>({
    values: user
      ? {
          username: user.username,
          name: user.name,
          email: user.email,
          role: user.role,
        }
      : {
          username: '',
          name: '',
          email: '',
          role: '',
        },
  })

  function onSave(values: EditFormValues) {
    if (!user) return
    updateUser.mutate(
      {
        user_id: user.user_id,
        username: values.username,
        name: values.name,
        email: values.email as Email,
        role: values.role,
        date_created: user.date_created,
        date_modified: new Date().toISOString(),
      },
      {
        onSuccess: (updated) => {
          reset({
            username: updated.username,
            name: updated.name,
            email: updated.email,
            role: updated.role,
          })
        },
      },
    )
  }

  function handlePasswordChange() {
    if (!user || !newPassword.trim()) return
    setPasswordSaving(true)
    updateUser.mutate(
      {
        user_id: user.user_id,
        username: user.username,
        name: user.name,
        email: user.email as Email,
        password: newPassword,
        role: user.role,
        date_created: user.date_created,
        date_modified: new Date().toISOString(),
      },
      {
        onSuccess: () => {
          setNewPassword('')
          setPasswordSaving(false)
        },
        onError: () => {
          setPasswordSaving(false)
        },
      },
    )
  }

  function handleAddKey() {
    if (!keyValue.trim() || !keyLabel.trim()) return
    createSshKey.mutate(
      { public_key: keyValue.trim(), label: keyLabel.trim() },
      {
        onSuccess: () => {
          setAddKeyOpen(false)
          setKeyLabel('')
          setKeyValue('')
        },
      },
    )
  }

  function handleDeleteKey() {
    if (!deleteKeyTarget) return
    deleteSshKey.mutate(deleteKeyTarget.ssh_key_id, {
      onSuccess: () => {
        setDeleteKeyOpen(false)
        setDeleteKeyTarget(null)
      },
    })
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">User Detail</h1>
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  if (!user) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">User Not Found</h1>
        <p className="text-muted-foreground">
          The requested user could not be found.
        </p>
        <Button variant="outline" asChild>
          <Link to="/users">Back to Users</Link>
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" asChild>
          <Link to="/users">
            <ArrowLeft className="h-4 w-4" />
          </Link>
        </Button>
        <div>
          <h1 className="text-2xl font-bold">{user.name}</h1>
          <p className="text-sm text-muted-foreground">
            Edit user account details.
          </p>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Account Details</CardTitle>
          <CardDescription>
            Update the user's username, name, email, and role.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSave)} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="edit-username">Username</Label>
              <Input
                id="edit-username"
                {...register('username', { required: true })}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-name">Name</Label>
              <Input
                id="edit-name"
                {...register('name', { required: true })}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-email">Email</Label>
              <Input
                id="edit-email"
                type="email"
                {...register('email', { required: true })}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-role">Role</Label>
              <Input
                id="edit-role"
                {...register('role', { required: true })}
              />
            </div>

            <Button
              type="submit"
              disabled={!isDirty || updateUser.isPending}
            >
              {updateUser.isPending && !passwordSaving ? 'Saving...' : 'Save Changes'}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Change Password</CardTitle>
          <CardDescription>
            Set a new password for this user.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={(e) => { e.preventDefault(); handlePasswordChange() }} className="flex items-end gap-4">
            <div className="flex-1 space-y-2">
              <Label htmlFor="new-password">New Password</Label>
              <Input
                id="new-password"
                type="password"
                autoComplete="new-password"
                placeholder="Enter new password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                required
              />
            </div>
            <Button
              type="submit"
              disabled={!newPassword.trim() || passwordSaving}
            >
              {passwordSaving ? 'Saving...' : 'Update Password'}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>SSH Keys</CardTitle>
              <CardDescription>
                Manage SSH public keys for this user.
              </CardDescription>
            </div>
            <Button size="sm" onClick={() => setAddKeyOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Add Key
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {sshKeys && sshKeys.length > 0 ? (
            <div className="space-y-3">
              {sshKeys.map((key) => (
                <div
                  key={key.ssh_key_id}
                  className="group flex items-center justify-between rounded-md border px-4 py-3"
                >
                  <div className="flex items-center gap-3">
                    <Key className="h-4 w-4 text-muted-foreground" />
                    <div>
                      <p className="text-sm font-medium">{key.label}</p>
                      <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <Badge variant="secondary" className="text-xs">{key.key_type}</Badge>
                        <span className="font-mono">{key.fingerprint}</span>
                      </div>
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-muted-foreground opacity-0 hover:text-destructive group-hover:opacity-100"
                    onClick={() => {
                      setDeleteKeyTarget(key)
                      setDeleteKeyOpen(true)
                    }}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">
              No SSH keys registered. Add a public key to enable SSH access.
            </p>
          )}
        </CardContent>
      </Card>

      <Dialog open={addKeyOpen} onOpenChange={setAddKeyOpen}>
        <DialogContent className="sm:max-w-4xl">
          <DialogHeader>
            <DialogTitle>Add SSH Key</DialogTitle>
            <DialogDescription>
              Paste a public key (e.g. ssh-ed25519 or ssh-rsa) to register it for this user.
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={(e) => { e.preventDefault(); handleAddKey() }} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="ssh-key-label">Label</Label>
              <Input
                id="ssh-key-label"
                placeholder="e.g. Work Laptop"
                value={keyLabel}
                onChange={(e) => setKeyLabel(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="ssh-key-value">Public Key</Label>
              <Textarea
                id="ssh-key-value"
                placeholder="ssh-ed25519 AAAA..."
                value={keyValue}
                onChange={(e) => setKeyValue(e.target.value)}
                rows={6}
                className="break-all font-mono text-sm"
                required
              />
            </div>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setAddKeyOpen(false)}
                disabled={createSshKey.isPending}
              >
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={!keyLabel.trim() || !keyValue.trim() || createSshKey.isPending}
              >
                {createSshKey.isPending ? 'Adding...' : 'Add Key'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteKeyOpen}
        onOpenChange={setDeleteKeyOpen}
        title="Remove SSH Key"
        description={`Are you sure you want to remove the SSH key "${deleteKeyTarget?.label}"? This action cannot be undone.`}
        onConfirm={handleDeleteKey}
        loading={deleteSshKey.isPending}
        variant="destructive"
      />
    </div>
  )
}
