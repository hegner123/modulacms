import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { FieldComponentProps } from '@/components/fields/field-renderer'

export function MediaField({ field, value, onChange, error }: FieldComponentProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [inputValue, setInputValue] = useState(value)

  function handleSelect() {
    if (isEditing) {
      onChange(inputValue)
      setIsEditing(false)
      return
    }
    setInputValue(value)
    setIsEditing(true)
  }

  function handleClear() {
    onChange('')
    setInputValue('')
    setIsEditing(false)
  }

  function handleCancel() {
    setInputValue(value)
    setIsEditing(false)
  }

  return (
    <div className="space-y-2">
      <Label htmlFor={field.field_id}>{field.label}</Label>
      {isEditing ? (
        <div className="space-y-2">
          <Input
            id={field.field_id}
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            placeholder="Enter media ID or URL"
          />
          <div className="flex items-center gap-2">
            <Button type="button" size="sm" onClick={handleSelect}>
              Save
            </Button>
            <Button type="button" variant="outline" size="sm" onClick={handleCancel}>
              Cancel
            </Button>
          </div>
        </div>
      ) : (
        <div className="space-y-2">
          {value ? (
            <p className="text-sm text-muted-foreground truncate rounded-md border px-3 py-2">
              {value}
            </p>
          ) : (
            <p className="text-sm text-muted-foreground italic rounded-md border px-3 py-2">
              No media selected
            </p>
          )}
          <div className="flex items-center gap-2">
            <Button type="button" size="sm" variant="outline" onClick={handleSelect}>
              Select Media
            </Button>
            {value && (
              <Button type="button" size="sm" variant="outline" onClick={handleClear}>
                Clear
              </Button>
            )}
          </div>
        </div>
      )}
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  )
}
