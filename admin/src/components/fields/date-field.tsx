import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { FieldComponentProps } from '@/components/fields/field-renderer'

export function DateField({ field, value, onChange, error }: FieldComponentProps) {
  // Convert ISO date string (e.g. "2024-01-15T00:00:00Z") to YYYY-MM-DD for the input
  function toDateInputValue(isoString: string): string {
    if (!isoString) return ''
    // If already in YYYY-MM-DD format, use as-is
    if (/^\d{4}-\d{2}-\d{2}$/.test(isoString)) return isoString
    // Otherwise extract the date portion from ISO string
    const date = new Date(isoString)
    if (Number.isNaN(date.getTime())) return isoString
    return date.toISOString().split('T')[0]
  }

  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    onChange(e.target.value)
  }

  return (
    <div className="space-y-2">
      <Label htmlFor={field.field_id}>{field.label}</Label>
      <Input
        id={field.field_id}
        type="date"
        value={toDateInputValue(value)}
        onChange={handleChange}
      />
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  )
}
