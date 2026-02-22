import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { FieldComponentProps } from '@/components/fields/field-renderer'

function toSlug(input: string): string {
  return input
    .toLowerCase()
    .trim()
    .replace(/\s+/g, '-')
    .replace(/[^a-z0-9-]/g, '')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
}

export function SlugField({ field, value, onChange, error }: FieldComponentProps) {
  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const formatted = toSlug(e.target.value)
    onChange(formatted)
  }

  return (
    <div className="space-y-2">
      <Label htmlFor={field.field_id}>{field.label}</Label>
      <Input
        id={field.field_id}
        type="text"
        value={value}
        onChange={handleChange}
        placeholder="enter-a-slug"
      />
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  )
}
