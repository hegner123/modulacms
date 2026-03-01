/**
 * Webhook entity types for event notification delivery.
 *
 * @module entities/webhook
 */

import type { WebhookID, WebhookDeliveryID, UserID } from '../ids.js'

/**
 * A webhook endpoint that receives event notifications.
 */
export type Webhook = {
  /** Unique identifier for this webhook. */
  webhook_id: WebhookID
  /** Human-readable name for this webhook. */
  name: string
  /** URL to deliver events to. */
  url: string
  /** HMAC secret for signing payloads. */
  secret: string
  /** List of event types this webhook subscribes to. */
  events: string[]
  /** Whether this webhook is active. */
  is_active: boolean
  /** Custom headers sent with each delivery. */
  headers: Record<string, string>
  /** ID of the user who created this webhook. */
  author_id: UserID
  /** ISO 8601 creation timestamp. */
  date_created: string
  /** ISO 8601 last-modified timestamp. */
  date_modified: string
}

/**
 * A single delivery attempt for a webhook event.
 */
export type WebhookDelivery = {
  /** Unique identifier for this delivery. */
  delivery_id: WebhookDeliveryID
  /** ID of the webhook this delivery belongs to. */
  webhook_id: WebhookID
  /** Event type that triggered this delivery. */
  event: string
  /** JSON payload sent to the webhook URL. */
  payload: string
  /** Current delivery status (pending, success, failed, retrying). */
  status: string
  /** Number of delivery attempts made. */
  attempts: number
  /** HTTP status code from the last attempt. */
  last_status_code: number
  /** Error message from the last failed attempt. */
  last_error: string
  /** ISO 8601 timestamp for the next retry attempt. */
  next_retry_at: string
  /** ISO 8601 creation timestamp. */
  created_at: string
  /** ISO 8601 timestamp when delivery completed (success or final failure). */
  completed_at: string
}
