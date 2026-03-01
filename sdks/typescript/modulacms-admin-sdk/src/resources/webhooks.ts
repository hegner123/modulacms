/**
 * Webhooks resource providing CRUD operations for webhook management,
 * plus test delivery and delivery history/retry.
 *
 * @module resources/webhooks
 * @internal
 */

import type { HttpClient } from '../http.js'
import type { RequestOptions, WebhookID, WebhookDeliveryID } from '../types/common.js'
import type { Webhook, WebhookDelivery } from '@modulacms/types'

// ---------------------------------------------------------------------------
// Request/Response types
// ---------------------------------------------------------------------------

/** Parameters for creating a new webhook via `POST /admin/webhooks`. */
export type CreateWebhookParams = {
  /** Human-readable name. */
  name: string
  /** Delivery URL (must be HTTPS unless allow_http is enabled). */
  url: string
  /** HMAC secret for signing. Empty string to auto-generate. */
  secret?: string
  /** Event types to subscribe to. */
  events: string[]
  /** Whether the webhook is active. */
  is_active: boolean
  /** Custom headers to include in deliveries. */
  headers?: Record<string, string>
}

/** Parameters for updating a webhook via `PUT /admin/webhooks/`. */
export type UpdateWebhookParams = {
  /** ID of the webhook to update. */
  webhook_id: WebhookID
  /** Updated name. */
  name: string
  /** Updated delivery URL. */
  url: string
  /** Updated HMAC secret. */
  secret?: string
  /** Updated event subscriptions. */
  events: string[]
  /** Updated active status. */
  is_active: boolean
  /** Updated custom headers. */
  headers?: Record<string, string>
}

/** Response from a webhook test request. */
export type WebhookTestResponse = {
  /** Result status. */
  status: string
  /** HTTP status code from the test delivery. */
  status_code?: number
  /** Error message if the test failed. */
  error?: string
}

// ---------------------------------------------------------------------------
// Webhooks resource type
// ---------------------------------------------------------------------------

/** Webhook management operations available on `client.webhooks`. */
export type WebhooksResource = {
  /** List all webhooks. */
  list: (opts?: RequestOptions) => Promise<Webhook[]>
  /** Get a single webhook by ID. */
  get: (id: WebhookID, opts?: RequestOptions) => Promise<Webhook>
  /** Create a new webhook. */
  create: (params: CreateWebhookParams, opts?: RequestOptions) => Promise<Webhook>
  /** Update an existing webhook. */
  update: (params: UpdateWebhookParams, opts?: RequestOptions) => Promise<Webhook>
  /** Remove a webhook by ID. */
  remove: (id: WebhookID, opts?: RequestOptions) => Promise<void>
  /** Send a test event to a webhook. */
  test: (id: WebhookID, opts?: RequestOptions) => Promise<WebhookTestResponse>
  /** List deliveries for a webhook. */
  listDeliveries: (id: WebhookID, opts?: RequestOptions) => Promise<WebhookDelivery[]>
  /** Retry a failed delivery. */
  retryDelivery: (deliveryID: WebhookDeliveryID, opts?: RequestOptions) => Promise<void>
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

/**
 * Create the webhooks resource bound to the given HTTP client.
 * @param http - Configured HTTP client.
 * @returns A {@link WebhooksResource} with CRUD, test, and delivery operations.
 * @internal
 */
function createWebhooksResource(http: HttpClient): WebhooksResource {
  return {
    list(opts?: RequestOptions): Promise<Webhook[]> {
      return http.get<Webhook[]>('/admin/webhooks', undefined, opts)
    },

    get(id: WebhookID, opts?: RequestOptions): Promise<Webhook> {
      return http.get<Webhook>(`/admin/webhooks/${String(id)}`, undefined, opts)
    },

    create(params: CreateWebhookParams, opts?: RequestOptions): Promise<Webhook> {
      return http.post<Webhook>('/admin/webhooks', params as unknown as Record<string, unknown>, opts)
    },

    update(params: UpdateWebhookParams, opts?: RequestOptions): Promise<Webhook> {
      return http.put<Webhook>(`/admin/webhooks/${String(params.webhook_id)}`, params as unknown as Record<string, unknown>, opts)
    },

    remove(id: WebhookID, opts?: RequestOptions): Promise<void> {
      return http.del(`/admin/webhooks/${String(id)}`, undefined, opts)
    },

    test(id: WebhookID, opts?: RequestOptions): Promise<WebhookTestResponse> {
      return http.post<WebhookTestResponse>(`/admin/webhooks/${String(id)}/test`, {} as Record<string, unknown>, opts)
    },

    listDeliveries(id: WebhookID, opts?: RequestOptions): Promise<WebhookDelivery[]> {
      return http.get<WebhookDelivery[]>(`/admin/webhooks/${String(id)}/deliveries`, undefined, opts)
    },

    retryDelivery(deliveryID: WebhookDeliveryID, opts?: RequestOptions): Promise<void> {
      return http.post(`/admin/webhooks/deliveries/${String(deliveryID)}/retry`, {} as Record<string, unknown>, opts)
    },
  }
}

export { createWebhooksResource }
