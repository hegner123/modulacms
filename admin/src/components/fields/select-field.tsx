import { useMemo } from 'react'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { FieldComponentProps } from '@/components/fields/field-renderer'

type SelectOption = {
  label: string
  value: string
}

type SelectFieldData = {
  options: SelectOption[]
}

export function SelectField({ field, value, onChange, error }: FieldComponentProps) {
  const { options, parseError } = useMemo(() => {
    if (!field.data) return { options: [] as SelectOption[], parseError: null }
    try {
      const parsed = JSON.parse(field.data) as SelectFieldData
      if (!parsed.options || !Array.isArray(parsed.options)) {
        return { options: [] as SelectOption[], parseError: null }
      }
      return { options: parsed.options, parseError: null }
    } catch {
      return { options: [] as SelectOption[], parseError: 'Failed to parse field options' }
    }
  }, [field.data])

  return (
    <div className="space-y-2">
      <Label htmlFor={field.field_id}>{field.label}</Label>
      <Select value={value} onValueChange={onChange} disabled={options.length === 0}>
        <SelectTrigger id={field.field_id} className="w-full">
          <SelectValue placeholder={options.length === 0 ? 'No options available' : 'Select an option'} />
        </SelectTrigger>
        <SelectContent>
          {options.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {parseError && <p className="text-sm text-destructive">{parseError}</p>}
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  )
}
