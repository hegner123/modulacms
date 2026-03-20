import type {
  FormDefinition,
  FormFieldDefinition,
  FormEntry,
  FormWebhook,
  PaginatedResponse,
  ExportResponse,
  SubmitResponse,
  QueueInfo,
  ApiError,
} from "./types";

export class FormsApiClient {
  private baseUrl: string;
  private apiKey: string | null;

  constructor(apiUrl: string, apiKey?: string) {
    this.baseUrl = apiUrl.replace(/\/$/, "") + "/api/v1/plugins/forms";
    this.apiKey = apiKey ?? null;
  }

  // ---------------------------------------------------------------------------
  // Public endpoints
  // ---------------------------------------------------------------------------

  async getPublicForm(formId: string): Promise<FormDefinition> {
    return this.request<FormDefinition>("GET", `/public/${formId}`);
  }

  async submitForm(
    formId: string,
    data: Record<string, unknown>,
  ): Promise<SubmitResponse> {
    return this.request<SubmitResponse>("POST", `/public/${formId}/submit`, data);
  }

  // ---------------------------------------------------------------------------
  // Form CRUD (admin)
  // ---------------------------------------------------------------------------

  async listForms(
    params?: { limit?: number; offset?: number },
  ): Promise<PaginatedResponse<FormDefinition>> {
    const query = this.buildQuery(params);
    return this.request<PaginatedResponse<FormDefinition>>("GET", `/forms${query}`);
  }

  async getForm(
    formId: string,
  ): Promise<{ form: FormDefinition; fields: FormFieldDefinition[] }> {
    return this.request<{ form: FormDefinition; fields: FormFieldDefinition[] }>(
      "GET",
      `/forms/${formId}`,
    );
  }

  async createForm(data: Partial<FormDefinition>): Promise<FormDefinition> {
    return this.request<FormDefinition>("POST", "/forms", data);
  }

  async updateForm(
    formId: string,
    data: Partial<FormDefinition> & { version: number },
  ): Promise<FormDefinition> {
    return this.request<FormDefinition>("PUT", `/forms/${formId}`, data);
  }

  async deleteForm(formId: string): Promise<{ deleted: boolean }> {
    return this.request<{ deleted: boolean }>("DELETE", `/forms/${formId}`);
  }

  // ---------------------------------------------------------------------------
  // Field CRUD (admin)
  // ---------------------------------------------------------------------------

  async listFields(
    formId: string,
  ): Promise<{ items: FormFieldDefinition[] }> {
    return this.request<{ items: FormFieldDefinition[] }>(
      "GET",
      `/forms/${formId}/fields`,
    );
  }

  async createField(
    formId: string,
    data: Partial<FormFieldDefinition> & { version: number },
  ): Promise<FormFieldDefinition> {
    return this.request<FormFieldDefinition>(
      "POST",
      `/forms/${formId}/fields`,
      data,
    );
  }

  async updateField(
    fieldId: string,
    data: Partial<FormFieldDefinition> & { version: number },
  ): Promise<FormFieldDefinition> {
    return this.request<FormFieldDefinition>(
      "PUT",
      `/fields/${fieldId}`,
      data,
    );
  }

  async deleteField(
    fieldId: string,
    version: number,
  ): Promise<{ deleted: boolean; version: number }> {
    return this.request<{ deleted: boolean; version: number }>(
      "DELETE",
      `/fields/${fieldId}?version=${version}`,
    );
  }

  async reorderFields(
    formId: string,
    fieldIds: string[],
    version: number,
  ): Promise<{ version: number }> {
    return this.request<{ version: number }>(
      "POST",
      `/forms/${formId}/fields/reorder`,
      { field_ids: fieldIds, version },
    );
  }

  // ---------------------------------------------------------------------------
  // Entry operations (admin)
  // ---------------------------------------------------------------------------

  async listEntries(
    formId: string,
    params?: { limit?: number; offset?: number },
  ): Promise<PaginatedResponse<FormEntry>> {
    const query = this.buildQuery(params);
    return this.request<PaginatedResponse<FormEntry>>(
      "GET",
      `/forms/${formId}/entries${query}`,
    );
  }

  async getEntry(entryId: string): Promise<FormEntry> {
    return this.request<FormEntry>("GET", `/entries/${entryId}`);
  }

  async deleteEntry(entryId: string): Promise<{ deleted: boolean }> {
    return this.request<{ deleted: boolean }>("DELETE", `/entries/${entryId}`);
  }

  async exportEntries(
    formId: string,
    after?: string,
  ): Promise<ExportResponse> {
    const query = after ? `?after=${encodeURIComponent(after)}` : "";
    return this.request<ExportResponse>(
      "GET",
      `/forms/${formId}/entries/export${query}`,
    );
  }

  // ---------------------------------------------------------------------------
  // Webhook CRUD (admin)
  // ---------------------------------------------------------------------------

  async listWebhooks(
    formId: string,
  ): Promise<{ items: FormWebhook[] }> {
    return this.request<{ items: FormWebhook[] }>(
      "GET",
      `/forms/${formId}/webhooks`,
    );
  }

  async createWebhook(
    formId: string,
    data: Partial<FormWebhook>,
  ): Promise<FormWebhook> {
    return this.request<FormWebhook>(
      "POST",
      `/forms/${formId}/webhooks`,
      data,
    );
  }

  async updateWebhook(
    webhookId: string,
    data: Partial<FormWebhook>,
  ): Promise<FormWebhook> {
    return this.request<FormWebhook>(
      "PUT",
      `/webhooks/${webhookId}`,
      data,
    );
  }

  async deleteWebhook(webhookId: string): Promise<{ deleted: boolean }> {
    return this.request<{ deleted: boolean }>(
      "DELETE",
      `/webhooks/${webhookId}`,
    );
  }

  async getQueueInfo(formId: string): Promise<QueueInfo> {
    return this.request<QueueInfo>("GET", `/forms/${formId}/webhooks/queue`);
  }

  // ---------------------------------------------------------------------------
  // Internal helpers
  // ---------------------------------------------------------------------------

  private async request<T>(
    method: string,
    path: string,
    body?: unknown,
  ): Promise<T> {
    const url = this.baseUrl + path;

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };

    if (this.apiKey) {
      headers["X-API-Key"] = this.apiKey;
    }

    const init: RequestInit = { method, headers };

    if (body !== undefined) {
      init.body = JSON.stringify(body);
    }

    const response = await fetch(url, init);

    if (!response.ok) {
      let message = `HTTP ${response.status}`;
      try {
        const err: ApiError = await response.json();
        if (err.error) {
          message = err.error;
        }
      } catch {
        // Response body was not valid JSON; use the status message.
      }
      throw new Error(message);
    }

    return (await response.json()) as T;
  }

  private buildQuery(params?: { limit?: number; offset?: number }): string {
    if (!params) {
      return "";
    }
    const parts: string[] = [];
    if (params.limit !== undefined) {
      parts.push(`limit=${params.limit}`);
    }
    if (params.offset !== undefined) {
      parts.push(`offset=${params.offset}`);
    }
    return parts.length > 0 ? `?${parts.join("&")}` : "";
  }
}
