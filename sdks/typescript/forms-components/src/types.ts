export interface FormDefinition {
  id: string;
  name: string;
  description: string;
  submit_label: string;
  success_message: string;
  redirect_url: string | null;
  captcha_config: CaptchaConfig | null;
  version: number;
  enabled: boolean;
  fields: FormFieldDefinition[];
}

export interface CaptchaConfig {
  provider: string;
  site_key: string;
}

export interface FormFieldDefinition {
  id: string;
  form_id: string;
  name: string;
  label: string;
  field_type: FieldType;
  placeholder: string | null;
  default_value: string | null;
  help_text: string | null;
  required: boolean;
  validation_rules: ValidationRules | null;
  options: string[] | null;
  position: number;
  config: Record<string, unknown> | null;
}

export type FieldType =
  | "text"
  | "textarea"
  | "email"
  | "number"
  | "tel"
  | "url"
  | "date"
  | "time"
  | "datetime"
  | "select"
  | "radio"
  | "checkbox"
  | "hidden"
  | "file";

export interface ValidationRules {
  min_length?: number;
  max_length?: number;
  max_file_size?: number;
}

export interface FormEntry {
  id: string;
  form_id: string;
  form_version: number;
  data: Record<string, unknown>;
  client_ip: string | null;
  user_agent: string | null;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface FormWebhook {
  id: string;
  form_id: string;
  url: string;
  method: string;
  headers: Record<string, string> | null;
  events: string;
  active: boolean;
  secret: string | null;
  created_at: string;
  updated_at: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface ExportResponse {
  items: FormEntry[];
  count: number;
  after: string | null;
  has_more: boolean;
}

export interface SubmitResponse {
  id: string;
  message: string;
  redirect_url: string | null;
}

export interface QueueInfo {
  pending: number;
  failed: number;
  recent_failures: QueueFailure[];
}

export interface QueueFailure {
  id: string;
  webhook_id: string | null;
  event: string;
  last_error: string | null;
  attempts: number;
  created_at: string;
}

export interface FieldValidationError {
  field: string;
  message: string;
}

export interface ApiError {
  error: string;
}
