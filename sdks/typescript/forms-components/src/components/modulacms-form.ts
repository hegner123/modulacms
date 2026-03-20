import type {
  FormDefinition,
  FormFieldDefinition,
  FieldValidationError,
  SubmitResponse,
} from "../types.js";
import { FormsApiClient } from "../api-client.js";
import { createShadowRoot, html, escapeHtml, $ } from "../utils/dom.js";
import { dispatch } from "../utils/events.js";
import { validateField, validateForm } from "../utils/validation.js";
import { generateId } from "../utils/aria.js";
import { CSS_CUSTOM_PROPERTIES } from "../utils/styles.js";
import { baseStyles } from "../styles/base.css.js";
import { formStyles } from "../styles/form.css.js";

const OBSERVED_ATTRIBUTES = [
  "form-id",
  "api-url",
  "api-key",
  "submit-label",
  "success-message",
  "redirect-url",
  "loading-text",
] as const;

type ObservedAttribute = (typeof OBSERVED_ATTRIBUTES)[number];

const enum Phase {
  Idle = "idle",
  Loading = "loading",
  Ready = "ready",
  Submitting = "submitting",
  Success = "success",
  Error = "error",
}

export class ModulaCMSForm extends HTMLElement {
  static get observedAttributes(): string[] {
    return [...OBSERVED_ATTRIBUTES];
  }

  private shadow: ShadowRoot | null = null;
  private api: FormsApiClient | null = null;
  private formDef: FormDefinition | null = null;
  private phase: Phase = Phase.Idle;
  private fieldIds: Map<string, string> = new Map();
  private recaptchaWidgetId: number | null = null;
  private abortController: AbortController | null = null;

  // -------------------------------------------------------------------------
  // Lifecycle
  // -------------------------------------------------------------------------

  connectedCallback(): void {
    if (!this.shadow) {
      this.shadow = this.attachShadow({ mode: "open" });
    }
    this.renderInitial();

    const formId = this.getAttribute("form-id");
    const apiUrl = this.getAttribute("api-url");
    if (formId && apiUrl) {
      this.loadForm();
    }
  }

  disconnectedCallback(): void {
    this.abortController?.abort();
    this.abortController = null;
  }

  attributeChangedCallback(
    name: string,
    oldValue: string | null,
    newValue: string | null,
  ): void {
    if (oldValue === newValue) {
      return;
    }
    if (name === "form-id" || name === "api-url" || name === "api-key") {
      const formId = this.getAttribute("form-id");
      const apiUrl = this.getAttribute("api-url");
      if (formId && apiUrl && this.isConnected) {
        this.loadForm();
      }
    }
  }

  // -------------------------------------------------------------------------
  // Public methods
  // -------------------------------------------------------------------------

  reset(): void {
    if (!this.shadow || !this.formDef) {
      return;
    }

    this.phase = Phase.Ready;

    const successEl = $(this.shadow, "[part='success']") as HTMLElement | null;
    if (successEl) {
      successEl.hidden = true;
    }

    const errorStateEl = $(this.shadow, "[part='error-state']") as HTMLElement | null;
    if (errorStateEl) {
      errorStateEl.hidden = true;
    }

    const formEl = $(this.shadow, "form") as HTMLFormElement | null;
    if (formEl) {
      formEl.hidden = false;
      formEl.reset();
    }

    // Restore default values
    for (const field of this.formDef.fields) {
      if (field.default_value !== null && field.default_value !== "") {
        this.setFieldInputValue(field, field.default_value);
      }
    }

    this.clearAllErrors();
  }

  validate(): FieldValidationError[] {
    if (!this.formDef) {
      return [];
    }
    const errors = validateForm(this.formDef.fields, this.collectFieldValues());
    this.displayValidationErrors(errors);
    return errors;
  }

  async submit(): Promise<void> {
    if (!this.formDef || !this.api) {
      return;
    }
    await this.handleSubmit();
  }

  setFieldValue(name: string, value: string): void {
    if (!this.shadow || !this.formDef) {
      return;
    }
    const field = this.formDef.fields.find((f) => f.name === name);
    if (!field) {
      return;
    }
    this.setFieldInputValue(field, value);
    dispatch(this, "field:change", { name, value });
  }

  getFieldValue(name: string): string | undefined {
    if (!this.shadow || !this.formDef) {
      return undefined;
    }
    const field = this.formDef.fields.find((f) => f.name === name);
    if (!field) {
      return undefined;
    }
    return this.getFieldInputValue(field);
  }

  // -------------------------------------------------------------------------
  // Form loading
  // -------------------------------------------------------------------------

  private async loadForm(): Promise<void> {
    const formId = this.getAttribute("form-id");
    const apiUrl = this.getAttribute("api-url");
    if (!formId || !apiUrl) {
      return;
    }

    this.abortController?.abort();
    this.abortController = new AbortController();

    const apiKey = this.getAttribute("api-key") ?? undefined;
    this.api = new FormsApiClient(apiUrl, apiKey);

    this.phase = Phase.Loading;
    this.renderLoading();

    try {
      const formDef = await this.api.getPublicForm(formId);
      // Check if we were aborted while awaiting
      if (this.abortController.signal.aborted) {
        return;
      }
      this.formDef = formDef;
      this.phase = Phase.Ready;
      this.renderForm();
      dispatch(this, "form:loaded", { form: formDef });
    } catch (err) {
      if (this.abortController.signal.aborted) {
        return;
      }
      this.phase = Phase.Error;
      const message = err instanceof Error ? err.message : "Failed to load form";
      this.renderErrorState(message);
      dispatch(this, "form:error", { error: message });
    }
  }

  // -------------------------------------------------------------------------
  // Rendering
  // -------------------------------------------------------------------------

  private renderInitial(): void {
    if (!this.shadow) {
      return;
    }
    this.shadow.innerHTML = `<style>${CSS_CUSTOM_PROPERTIES}${baseStyles}${formStyles}</style><div part="loading" aria-live="polite"></div>`;
  }

  private renderLoading(): void {
    if (!this.shadow) {
      return;
    }
    const loadingText = this.getAttribute("loading-text") ?? "Loading form...";
    this.shadow.innerHTML =
      `<style>${CSS_CUSTOM_PROPERTIES}${baseStyles}${formStyles}</style>` +
      `<div part="loading" role="status" aria-live="polite">${escapeHtml(loadingText)}</div>`;
  }

  private renderErrorState(message: string): void {
    if (!this.shadow) {
      return;
    }
    this.shadow.innerHTML =
      `<style>${CSS_CUSTOM_PROPERTIES}${baseStyles}${formStyles}</style>` +
      `<div part="error-state" role="alert">${escapeHtml(message)}</div>`;
  }

  private renderForm(): void {
    if (!this.shadow || !this.formDef) {
      return;
    }

    this.fieldIds.clear();
    const formDef = this.formDef;

    const submitLabel =
      this.getAttribute("submit-label") ?? formDef.submit_label ?? "Submit";

    const fieldsMarkup = formDef.fields
      .slice()
      .sort((a, b) => a.position - b.position)
      .map((field) => this.renderFieldMarkup(field))
      .join("");

    const honeypotMarkup =
      '<input name="_hp" type="text" tabindex="-1" autocomplete="off" ' +
      'aria-hidden="true" style="position:absolute;left:-9999px;opacity:0">';

    const captchaMarkup = this.renderCaptchaPlaceholder();

    const formHtml =
      `<style>${CSS_CUSTOM_PROPERTIES}${baseStyles}${formStyles}</style>` +
      `<form part="form" novalidate>` +
      fieldsMarkup +
      honeypotMarkup +
      captchaMarkup +
      `<button type="submit" part="submit">${escapeHtml(submitLabel)}</button>` +
      `</form>` +
      `<div part="loading" role="status" aria-live="polite" hidden></div>` +
      `<div part="success" role="status" aria-live="polite" hidden></div>` +
      `<div part="error-state" role="alert" hidden></div>`;

    this.shadow.innerHTML = formHtml;

    this.attachFormListeners();
    this.setDefaultValues();
    this.initCaptcha();
  }

  private renderFieldMarkup(field: FormFieldDefinition): string {
    const inputId = generateId(field.name);
    this.fieldIds.set(field.name, inputId);

    const errorId = inputId + "-error";
    const helpId = field.help_text ? inputId + "-help" : "";

    const describedBy = [helpId, errorId].filter(Boolean).join(" ");
    const describedByAttr = describedBy
      ? ` aria-describedby="${escapeHtml(describedBy)}"`
      : "";
    const requiredAttr = field.required ? " required" : "";
    const requiredIndicator = field.required
      ? ' <span aria-hidden="true">*</span>'
      : "";

    const placeholderAttr =
      field.placeholder !== null && field.placeholder !== ""
        ? ` placeholder="${escapeHtml(field.placeholder)}"`
        : "";

    if (field.field_type === "hidden") {
      const defaultVal = field.default_value ?? "";
      return (
        `<input type="hidden" name="${escapeHtml(field.name)}" ` +
        `id="${escapeHtml(inputId)}" value="${escapeHtml(defaultVal)}">`
      );
    }

    let inputMarkup: string;

    switch (field.field_type) {
      case "textarea":
        inputMarkup = this.renderTextarea(
          field,
          inputId,
          describedByAttr,
          requiredAttr,
          placeholderAttr,
        );
        break;
      case "select":
        inputMarkup = this.renderSelect(
          field,
          inputId,
          describedByAttr,
          requiredAttr,
        );
        break;
      case "radio":
        inputMarkup = this.renderRadioGroup(
          field,
          inputId,
          describedByAttr,
          requiredAttr,
        );
        break;
      case "checkbox":
        inputMarkup = this.renderCheckbox(
          field,
          inputId,
          describedByAttr,
          requiredAttr,
        );
        break;
      case "file":
        inputMarkup = this.renderFileInput(
          field,
          inputId,
          describedByAttr,
          requiredAttr,
        );
        break;
      default:
        inputMarkup = this.renderStandardInput(
          field,
          inputId,
          describedByAttr,
          requiredAttr,
          placeholderAttr,
        );
        break;
    }

    // Checkbox has its own label arrangement (label wraps the input)
    const labelMarkup =
      field.field_type === "checkbox"
        ? ""
        : `<label part="label" for="${escapeHtml(inputId)}">${escapeHtml(field.label)}${requiredIndicator}</label>`;

    const helpMarkup =
      field.help_text !== null && field.help_text !== ""
        ? `<span part="help-text" id="${escapeHtml(helpId)}">${escapeHtml(field.help_text)}</span>`
        : "";

    const errorMarkup =
      `<span part="error" id="${escapeHtml(errorId)}" role="alert" aria-live="polite" hidden></span>`;

    return (
      `<div part="field" data-field-name="${escapeHtml(field.name)}" data-field-type="${escapeHtml(field.field_type)}">` +
      labelMarkup +
      inputMarkup +
      helpMarkup +
      errorMarkup +
      `</div>`
    );
  }

  private renderStandardInput(
    field: FormFieldDefinition,
    inputId: string,
    describedByAttr: string,
    requiredAttr: string,
    placeholderAttr: string,
  ): string {
    const type = field.field_type === "datetime" ? "datetime-local" : field.field_type;
    let extraAttrs = "";

    if (field.validation_rules) {
      if (field.validation_rules.min_length !== undefined) {
        extraAttrs += ` minlength="${field.validation_rules.min_length}"`;
      }
      if (field.validation_rules.max_length !== undefined) {
        extraAttrs += ` maxlength="${field.validation_rules.max_length}"`;
      }
    }

    return (
      `<input part="input" type="${escapeHtml(type)}" ` +
      `name="${escapeHtml(field.name)}" ` +
      `id="${escapeHtml(inputId)}"` +
      describedByAttr +
      requiredAttr +
      placeholderAttr +
      extraAttrs +
      `>`
    );
  }

  private renderTextarea(
    field: FormFieldDefinition,
    inputId: string,
    describedByAttr: string,
    requiredAttr: string,
    placeholderAttr: string,
  ): string {
    let extraAttrs = "";

    if (field.validation_rules) {
      if (field.validation_rules.min_length !== undefined) {
        extraAttrs += ` minlength="${field.validation_rules.min_length}"`;
      }
      if (field.validation_rules.max_length !== undefined) {
        extraAttrs += ` maxlength="${field.validation_rules.max_length}"`;
      }
    }

    return (
      `<textarea part="input" ` +
      `name="${escapeHtml(field.name)}" ` +
      `id="${escapeHtml(inputId)}"` +
      describedByAttr +
      requiredAttr +
      placeholderAttr +
      extraAttrs +
      `></textarea>`
    );
  }

  private renderSelect(
    field: FormFieldDefinition,
    inputId: string,
    describedByAttr: string,
    requiredAttr: string,
  ): string {
    const options = field.options ?? [];
    const optionsMarkup =
      `<option value="">-- Select --</option>` +
      options
        .map(
          (opt) =>
            `<option value="${escapeHtml(opt)}">${escapeHtml(opt)}</option>`,
        )
        .join("");

    return (
      `<select part="input" ` +
      `name="${escapeHtml(field.name)}" ` +
      `id="${escapeHtml(inputId)}"` +
      describedByAttr +
      requiredAttr +
      `>${optionsMarkup}</select>`
    );
  }

  private renderRadioGroup(
    field: FormFieldDefinition,
    inputId: string,
    describedByAttr: string,
    requiredAttr: string,
  ): string {
    const options = field.options ?? [];
    const groupName = field.name;

    const radiosMarkup = options
      .map((opt, idx) => {
        const radioId = inputId + "-" + idx;
        return (
          `<label>` +
          `<input type="radio" ` +
          `name="${escapeHtml(groupName)}" ` +
          `id="${escapeHtml(radioId)}" ` +
          `value="${escapeHtml(opt)}"` +
          requiredAttr +
          (idx === 0 ? describedByAttr : "") +
          `> ${escapeHtml(opt)}</label>`
        );
      })
      .join("");

    return (
      `<div part="input" role="radiogroup" ` +
      `aria-labelledby="${escapeHtml(inputId)}-grouplabel">` +
      radiosMarkup +
      `</div>`
    );
  }

  private renderCheckbox(
    field: FormFieldDefinition,
    inputId: string,
    describedByAttr: string,
    requiredAttr: string,
  ): string {
    const requiredIndicator = field.required
      ? ' <span aria-hidden="true">*</span>'
      : "";

    return (
      `<label part="label" for="${escapeHtml(inputId)}">` +
      `<input part="input" type="checkbox" ` +
      `name="${escapeHtml(field.name)}" ` +
      `id="${escapeHtml(inputId)}" ` +
      `value="true"` +
      describedByAttr +
      requiredAttr +
      `> ${escapeHtml(field.label)}${requiredIndicator}</label>`
    );
  }

  private renderFileInput(
    field: FormFieldDefinition,
    inputId: string,
    describedByAttr: string,
    requiredAttr: string,
  ): string {
    let acceptAttr = "";
    if (field.config && typeof field.config["accept"] === "string") {
      acceptAttr = ` accept="${escapeHtml(field.config["accept"])}"`;
    }

    let maxSizeAttr = "";
    if (field.validation_rules?.max_file_size !== undefined) {
      maxSizeAttr = ` data-max-size="${field.validation_rules.max_file_size}"`;
    }

    return (
      `<input part="input" type="file" ` +
      `name="${escapeHtml(field.name)}" ` +
      `id="${escapeHtml(inputId)}"` +
      describedByAttr +
      requiredAttr +
      acceptAttr +
      maxSizeAttr +
      `>`
    );
  }

  private renderCaptchaPlaceholder(): string {
    if (!this.formDef?.captcha_config) {
      return "";
    }
    return '<div data-captcha-container></div>';
  }

  // -------------------------------------------------------------------------
  // Event listeners
  // -------------------------------------------------------------------------

  private attachFormListeners(): void {
    if (!this.shadow) {
      return;
    }

    const formEl = $(this.shadow, "form") as HTMLFormElement | null;
    if (!formEl) {
      return;
    }

    formEl.addEventListener("submit", (e: Event) => {
      e.preventDefault();
      this.handleSubmit();
    });

    formEl.addEventListener("input", (e: Event) => {
      const target = e.target;
      if (!(target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement || target instanceof HTMLSelectElement)) {
        return;
      }
      const name = target.name;
      if (!name || name === "_hp") {
        return;
      }

      const value = target instanceof HTMLInputElement && target.type === "checkbox"
        ? target.checked ? "true" : ""
        : target.value;

      dispatch(this, "field:change", { name, value });

      // Live validation: clear error on valid input, show on invalid
      const field = this.formDef?.fields.find((f) => f.name === name);
      if (field) {
        const fieldError = validateField(field, value);
        this.setFieldError(name, fieldError);
        dispatch(this, "field:validate", {
          name,
          error: fieldError,
        });
      }
    });

    // Handle file inputs: convert to base64 on change
    formEl.addEventListener("change", (e: Event) => {
      const target = e.target;
      if (!(target instanceof HTMLInputElement) || target.type !== "file") {
        return;
      }
      this.handleFileChange(target);
    });
  }

  // -------------------------------------------------------------------------
  // Submit flow
  // -------------------------------------------------------------------------

  private async handleSubmit(): Promise<void> {
    if (!this.shadow || !this.formDef || !this.api) {
      return;
    }

    if (this.phase === Phase.Submitting) {
      return;
    }

    // 1. Validate
    const values = this.collectFieldValues();
    const errors = validateForm(this.formDef.fields, values);
    this.displayValidationErrors(errors);

    if (errors.length > 0) {
      const message = "Please correct the errors above.";
      dispatch(this, "form:error", { error: message });
      // Focus the first field with an error
      this.focusFirstError(errors);
      return;
    }

    // 2. Dispatch cancelable form:submit
    const allowed = dispatch(this, "form:submit", { data: values }, { cancelable: true });
    if (!allowed) {
      return;
    }

    // 3. Show loading
    this.phase = Phase.Submitting;
    this.setLoadingVisible(true);
    this.setSubmitDisabled(true);

    // 4. Collect all data including honeypot
    const data: Record<string, unknown> = { ...values };
    const hpInput = $(this.shadow, 'input[name="_hp"]') as HTMLInputElement | null;
    if (hpInput) {
      data["_hp"] = hpInput.value;
    }

    // Include reCAPTCHA token if present
    const recaptchaResponse = this.getRecaptchaResponse();
    if (recaptchaResponse) {
      data["g-recaptcha-response"] = recaptchaResponse;
    }

    // 5. Submit
    try {
      const formId = this.getAttribute("form-id");
      if (!formId) {
        throw new Error("Missing form-id attribute");
      }
      const response = await this.api.submitForm(formId, data);

      this.phase = Phase.Success;
      this.setLoadingVisible(false);

      // 6. Success: redirect or show message
      const redirectUrl =
        this.getAttribute("redirect-url") ??
        response.redirect_url ??
        this.formDef.redirect_url;

      if (redirectUrl) {
        dispatch<{ response: SubmitResponse }>(this, "form:success", { response });
        window.location.href = redirectUrl;
        return;
      }

      this.showSuccessMessage(response);
      dispatch<{ response: SubmitResponse }>(this, "form:success", { response });
    } catch (err) {
      this.phase = Phase.Ready;
      this.setLoadingVisible(false);
      this.setSubmitDisabled(false);

      const message = err instanceof Error ? err.message : "Submission failed";
      this.showErrorState(message);
      dispatch(this, "form:error", { error: message });
    }
  }

  // -------------------------------------------------------------------------
  // Field value helpers
  // -------------------------------------------------------------------------

  private collectFieldValues(): Record<string, string> {
    if (!this.shadow || !this.formDef) {
      return {};
    }

    const values: Record<string, string> = {};
    for (const field of this.formDef.fields) {
      const value = this.getFieldInputValue(field);
      if (value !== undefined) {
        values[field.name] = value;
      }
    }
    return values;
  }

  private getFieldInputValue(field: FormFieldDefinition): string | undefined {
    if (!this.shadow) {
      return undefined;
    }

    const inputId = this.fieldIds.get(field.name);
    if (!inputId) {
      return undefined;
    }

    if (field.field_type === "radio") {
      const checked = $(
        this.shadow,
        `input[name="${CSS.escape(field.name)}"]:checked`,
      ) as HTMLInputElement | null;
      return checked?.value ?? "";
    }

    if (field.field_type === "checkbox") {
      const cb = $(this.shadow, `#${CSS.escape(inputId)}`) as HTMLInputElement | null;
      return cb?.checked ? "true" : "";
    }

    if (field.field_type === "file") {
      // File value is stored as a data attribute after base64 conversion
      const fileInput = $(this.shadow, `#${CSS.escape(inputId)}`) as HTMLInputElement | null;
      return fileInput?.dataset.base64Value ?? "";
    }

    const el = $(this.shadow, `#${CSS.escape(inputId)}`) as
      | HTMLInputElement
      | HTMLTextAreaElement
      | HTMLSelectElement
      | null;
    return el?.value ?? undefined;
  }

  private setFieldInputValue(field: FormFieldDefinition, value: string): void {
    if (!this.shadow) {
      return;
    }

    const inputId = this.fieldIds.get(field.name);
    if (!inputId) {
      return;
    }

    if (field.field_type === "radio") {
      const radios = this.shadow.querySelectorAll(
        `input[name="${CSS.escape(field.name)}"]`,
      );
      for (const radio of Array.from(radios)) {
        if (radio instanceof HTMLInputElement) {
          radio.checked = radio.value === value;
        }
      }
      return;
    }

    if (field.field_type === "checkbox") {
      const cb = $(this.shadow, `#${CSS.escape(inputId)}`) as HTMLInputElement | null;
      if (cb) {
        cb.checked = value === "true";
      }
      return;
    }

    // file inputs cannot have their value set programmatically
    if (field.field_type === "file") {
      return;
    }

    const el = $(this.shadow, `#${CSS.escape(inputId)}`) as
      | HTMLInputElement
      | HTMLTextAreaElement
      | HTMLSelectElement
      | null;
    if (el) {
      el.value = value;
    }
  }

  private setDefaultValues(): void {
    if (!this.formDef) {
      return;
    }
    for (const field of this.formDef.fields) {
      if (field.default_value !== null && field.default_value !== "") {
        this.setFieldInputValue(field, field.default_value);
      }
    }
  }

  // -------------------------------------------------------------------------
  // File handling
  // -------------------------------------------------------------------------

  private handleFileChange(input: HTMLInputElement): void {
    const file = input.files?.[0];
    if (!file) {
      input.dataset.base64Value = "";
      return;
    }

    // Check max file size
    const maxSize = input.dataset.maxSize ? parseInt(input.dataset.maxSize, 10) : 0;
    if (maxSize > 0 && file.size > maxSize) {
      const name = input.name;
      const maxMB = (maxSize / (1024 * 1024)).toFixed(1);
      this.setFieldError(name, `File exceeds maximum size of ${maxMB} MB`);
      input.value = "";
      input.dataset.base64Value = "";
      return;
    }

    const reader = new FileReader();
    reader.onload = () => {
      if (typeof reader.result === "string") {
        input.dataset.base64Value = reader.result;
      }
    };
    reader.onerror = () => {
      input.dataset.base64Value = "";
      this.setFieldError(input.name, "Failed to read file");
    };
    reader.readAsDataURL(file);
  }

  // -------------------------------------------------------------------------
  // Validation display
  // -------------------------------------------------------------------------

  private displayValidationErrors(errors: FieldValidationError[]): void {
    this.clearAllErrors();
    for (const err of errors) {
      this.setFieldError(err.field, err.message);
    }
  }

  private setFieldError(fieldName: string, message: string | null): void {
    if (!this.shadow) {
      return;
    }

    const inputId = this.fieldIds.get(fieldName);
    if (!inputId) {
      return;
    }

    const errorId = inputId + "-error";
    const errorEl = $(this.shadow, `#${CSS.escape(errorId)}`) as HTMLElement | null;
    if (!errorEl) {
      return;
    }

    if (message) {
      errorEl.textContent = message;
      errorEl.hidden = false;
      // Mark the input as invalid
      const inputEl = $(this.shadow, `#${CSS.escape(inputId)}`) as HTMLElement | null;
      if (inputEl) {
        inputEl.setAttribute("aria-invalid", "true");
      }
    } else {
      errorEl.textContent = "";
      errorEl.hidden = true;
      const inputEl = $(this.shadow, `#${CSS.escape(inputId)}`) as HTMLElement | null;
      if (inputEl) {
        inputEl.removeAttribute("aria-invalid");
      }
    }
  }

  private clearAllErrors(): void {
    if (!this.shadow) {
      return;
    }
    for (const [fieldName] of this.fieldIds) {
      this.setFieldError(fieldName, null);
    }
  }

  private focusFirstError(errors: FieldValidationError[]): void {
    if (!this.shadow || errors.length === 0) {
      return;
    }
    const firstFieldName = errors[0]!.field;
    const inputId = this.fieldIds.get(firstFieldName);
    if (!inputId) {
      return;
    }
    const inputEl = $(this.shadow, `#${CSS.escape(inputId)}`) as HTMLElement | null;
    if (inputEl) {
      inputEl.focus();
    }
  }

  // -------------------------------------------------------------------------
  // UI state helpers
  // -------------------------------------------------------------------------

  private setLoadingVisible(visible: boolean): void {
    if (!this.shadow) {
      return;
    }

    const loadingEl = $(this.shadow, "[part='loading']") as HTMLElement | null;
    if (loadingEl) {
      const loadingText = this.getAttribute("loading-text") ?? "Submitting...";
      loadingEl.textContent = visible ? loadingText : "";
      loadingEl.hidden = !visible;
    }
  }

  private setSubmitDisabled(disabled: boolean): void {
    if (!this.shadow) {
      return;
    }
    const btn = $(this.shadow, "[part='submit']") as HTMLButtonElement | null;
    if (btn) {
      btn.disabled = disabled;
    }
  }

  private showSuccessMessage(response: SubmitResponse): void {
    if (!this.shadow) {
      return;
    }

    const successMessage =
      this.getAttribute("success-message") ??
      response.message ??
      "Thank you for your submission.";

    // Hide the form
    const formEl = $(this.shadow, "form") as HTMLFormElement | null;
    if (formEl) {
      formEl.hidden = true;
    }

    const successEl = $(this.shadow, "[part='success']") as HTMLElement | null;
    if (successEl) {
      successEl.textContent = successMessage;
      successEl.hidden = false;
    }
  }

  private showErrorState(message: string): void {
    if (!this.shadow) {
      return;
    }
    const errorStateEl = $(this.shadow, "[part='error-state']") as HTMLElement | null;
    if (errorStateEl) {
      errorStateEl.textContent = message;
      errorStateEl.hidden = false;
    }
  }

  // -------------------------------------------------------------------------
  // reCAPTCHA
  // -------------------------------------------------------------------------

  private initCaptcha(): void {
    if (!this.shadow || !this.formDef?.captcha_config) {
      return;
    }

    const config = this.formDef.captcha_config;
    if (config.provider !== "recaptcha") {
      return;
    }

    const container = $(this.shadow, "[data-captcha-container]") as HTMLElement | null;
    if (!container) {
      return;
    }

    // Load reCAPTCHA script if not already present
    this.loadRecaptchaScript(config.site_key, container);
  }

  private loadRecaptchaScript(siteKey: string, container: HTMLElement): void {
    const scriptId = "modulacms-recaptcha-script";
    if (document.getElementById(scriptId)) {
      // Script already loaded; render widget
      this.renderRecaptchaWidget(siteKey, container);
      return;
    }

    const script = document.createElement("script");
    script.id = scriptId;
    script.src = "https://www.google.com/recaptcha/api.js?render=explicit";
    script.async = true;
    script.defer = true;

    script.onload = () => {
      this.renderRecaptchaWidget(siteKey, container);
    };
    script.onerror = () => {
      container.textContent = "Failed to load CAPTCHA.";
    };

    document.head.appendChild(script);
  }

  private renderRecaptchaWidget(siteKey: string, container: HTMLElement): void {
    const grecaptcha = (window as unknown as Record<string, unknown>)[
      "grecaptcha"
    ] as {
      render: (
        el: HTMLElement,
        params: { sitekey: string },
      ) => number;
    } | undefined;

    if (!grecaptcha) {
      // API not ready yet; wait for it
      const checkInterval = setInterval(() => {
        const api = (window as unknown as Record<string, unknown>)[
          "grecaptcha"
        ] as { render: (el: HTMLElement, opts: { sitekey: string }) => number } | undefined;
        if (api) {
          clearInterval(checkInterval);
          this.recaptchaWidgetId = api.render(container, { sitekey: siteKey });
        }
      }, 100);

      // Stop checking after 10 seconds
      setTimeout(() => clearInterval(checkInterval), 10_000);
      return;
    }

    this.recaptchaWidgetId = grecaptcha.render(container, { sitekey: siteKey });
  }

  private getRecaptchaResponse(): string | null {
    if (this.recaptchaWidgetId === null) {
      return null;
    }

    const grecaptcha = (window as unknown as Record<string, unknown>)[
      "grecaptcha"
    ] as {
      getResponse: (widgetId: number) => string;
    } | undefined;

    if (!grecaptcha) {
      return null;
    }

    const response = grecaptcha.getResponse(this.recaptchaWidgetId);
    return response || null;
  }
}
