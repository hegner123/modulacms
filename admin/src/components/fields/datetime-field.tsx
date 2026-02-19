import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { FieldComponentProps } from '@/components/fields/field-renderer'

export function DatetimeField({ field, value, onChange, error }: FieldComponentProps) {
  // Convert ISO datetime string to datetime-local format (YYYY-MM-DDTHH:MM)
  function toDatetimeLocalValue(isoString: string): string {
    if (!isoString) return ''
    // If already in datetime-local format, use as-is
    if (/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/.test(isoString)) return isoString
    const date = new Date(isoString)
    if (Number.isNaN(date.getTime())) return isoString
    // Format as YYYY-MM-DDTHH:MM for datetime-local input
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    return `${year}-${month}-${day}T${hours}:${minutes}`
  }

  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    onChange(e.target.value)
  }

  return (
    <div className="space-y-2">
      <Label htmlFor={field.field_id}>{field.label}</Label>
      <Input
        id={field.field_id}
        type="datetime-local"
        value={toDatetimeLocalValue(value)}
        onChange={handleChange}
      />
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  )
}
