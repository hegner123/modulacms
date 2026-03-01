import Foundation

/// Provides CRUD operations for webhooks, plus test delivery and delivery history.
public final class WebhookResource: Sendable {
    let http: HTTPClient

    init(http: HTTPClient) {
        self.http = http
    }

    // MARK: - CRUD

    /// List all webhooks.
    public func list() async throws -> [Webhook] {
        try await http.get(path: "/api/v1/admin/webhooks")
    }

    /// Get a single webhook by ID.
    public func get(id: WebhookID) async throws -> Webhook {
        try await http.get(path: "/api/v1/admin/webhooks/\(id.rawValue)")
    }

    /// Create a new webhook.
    public func create(params: CreateWebhookRequest) async throws -> Webhook {
        try await http.post(path: "/api/v1/admin/webhooks", body: params)
    }

    /// Update an existing webhook.
    public func update(params: UpdateWebhookRequest) async throws -> Webhook {
        try await http.put(path: "/api/v1/admin/webhooks/\(params.webhookID.rawValue)", body: params)
    }

    /// Delete a webhook by ID.
    public func delete(id: WebhookID) async throws {
        try await http.delete(path: "/api/v1/admin/webhooks/\(id.rawValue)")
    }

    // MARK: - Test

    /// Send a test event to a webhook and return the delivery result.
    public func test(id: WebhookID) async throws -> WebhookTestResponse {
        try await http.post(path: "/api/v1/admin/webhooks/\(id.rawValue)/test", body: EmptyBody())
    }

    // MARK: - Deliveries

    /// List delivery history for a webhook.
    public func listDeliveries(id: WebhookID) async throws -> [WebhookDelivery] {
        try await http.get(path: "/api/v1/admin/webhooks/\(id.rawValue)/deliveries")
    }

    /// Retry a failed delivery.
    public func retryDelivery(deliveryID: WebhookDeliveryID) async throws {
        let _: WebhookRetryResponse = try await http.post(
            path: "/api/v1/admin/webhooks/deliveries/\(deliveryID.rawValue)/retry",
            body: EmptyBody()
        )
    }
}

private struct WebhookRetryResponse: Decodable {}
