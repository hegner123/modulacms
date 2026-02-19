import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

import { BooleanField } from '@/components/fields/boolean-field'
import { DateField } from '@/components/fields/date-field'
import { DatetimeField } from '@/components/fields/datetime-field'
import { EmailField } from '@/components/fields/email-field'
import { JsonField } from '@/components/fields/json-field'
import { MediaField } from '@/components/fields/media-field'
import { NumberField } from '@/components/fields/number-field'
import { RelationField } from '@/components/fields/relation-field'
import { RichtextField } from '@/components/fields/richtext-field'
import { SelectField } from '@/components/fields/select-field'
import { SlugField } from '@/components/fields/slug-field'
import { TextField } from '@/components/fields/text-field'
import { TextareaField } from '@/components/fields/textarea-field'
import { UrlField } from '@/components/fields/url-field'

export type FieldType =
  | 'text'
  | 'textarea'
  | 'number'
  | 'date'
  | 'datetime'
  | 'boolean'
  | 'select'
  | 'media'
  | 'relation'
  | 'json'
  | 'richtext'
  | 'slug'
  | 'email'
  | 'url'

export type FieldComponentProps = {
  field: {
    field_id: string
    label: string
    type: FieldType
    validation: string
    ui_config: string
    data: string
  }
  value: string
  onChange: (value: string) => void
  error?: string
}

export function FieldRenderer(props: FieldComponentProps) {
  switch (props.field.type) {
    case 'text':
      return <TextField {...props} />
    case 'textarea':
      return <TextareaField {...props} />
    case 'number':
      return <NumberField {...props} />
    case 'date':
      return <DateField {...props} />
    case 'datetime':
      return <DatetimeField {...props} />
    case 'boolean':
      return <BooleanField {...props} />
    case 'select':
      return <SelectField {...props} />
    case 'media':
      return <MediaField {...props} />
    case 'relation':
      return <RelationField {...props} />
    case 'json':
      return <JsonField {...props} />
    case 'richtext':
      return <RichtextField {...props} />
    case 'slug':
      return <SlugField {...props} />
    case 'email':
      return <EmailField {...props} />
    case 'url':
      return <UrlField {...props} />
    default: {
      const unknownType = (props.field as { type: string }).type
      return (
        <div className="space-y-2">
          <Label htmlFor={props.field.field_id}>{props.field.label}</Label>
          <div className="flex items-center gap-2">
            <Input
              id={props.field.field_id}
              value={props.value}
              onChange={(e) => props.onChange(e.target.value)}
            />
            <Badge variant="destructive">Unknown field type: {unknownType}</Badge>
          </div>
          {props.error && <p className="text-sm text-destructive">{props.error}</p>}
        </div>
      )
    }
  }
}
