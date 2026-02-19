import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import type { FieldComponentProps } from '@/components/fields/field-renderer'

export function BooleanField({ field, value, onChange, error }: FieldComponentProps) {
  const checked = value === 'true'

  function handleCheckedChange(checkedState: boolean | 'indeterminate') {
    if (checkedState === 'indeterminate') {
      onChange('false')
      return
    }
    onChange(checkedState ? 'true' : 'false')
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <Checkbox
          id={field.field_id}
          checked={checked}
          onCheckedChange={handleCheckedChange}
        />
        <Label htmlFor={field.field_id}>{field.label}</Label>
      </div>
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  )
}
