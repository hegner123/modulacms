import { useState } from 'react'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import type { FieldComponentProps } from '@/components/fields/field-renderer'

export function JsonField({ field, value, onChange, error }: FieldComponentProps) {
  const [jsonError, setJsonError] = useState<string | null>(null)

  function handleBlur() {
    if (!value.trim()) {
      setJsonError(null)
      return
    }
    try {
      JSON.parse(value)
      setJsonError(null)
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Invalid JSON'
      setJsonError(message)
    }
  }

  return (
    <div className="space-y-2">
      <Label htmlFor={field.field_id}>{field.label}</Label>
      <Textarea
        id={field.field_id}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onBlur={handleBlur}
        className="min-h-[120px] font-mono text-sm"
        placeholder='{"key": "value"}'
      />
      {jsonError && <p className="text-sm text-destructive">{jsonError}</p>}
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  )
}
