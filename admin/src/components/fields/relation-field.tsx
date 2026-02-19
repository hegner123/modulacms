import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { FieldComponentProps } from '@/components/fields/field-renderer'

export function RelationField({ field, value, onChange, error }: FieldComponentProps) {
  return (
    <div className="space-y-2">
      <Label htmlFor={field.field_id}>{field.label}</Label>
      <Input
        id={field.field_id}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Enter content ID"
      />
      {value && (
        <p className="text-xs text-muted-foreground">
          Related content ID: {value}
        </p>
      )}
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  )
}
