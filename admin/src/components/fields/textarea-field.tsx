import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import type { FieldComponentProps } from '@/components/fields/field-renderer'

export function TextareaField({ field, value, onChange, error }: FieldComponentProps) {
  return (
    <div className="space-y-2">
      <Label htmlFor={field.field_id}>{field.label}</Label>
      <Textarea
        id={field.field_id}
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  )
}
