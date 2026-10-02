import type { Form, FormInput, FormRecord, FormSummary, RecordInput, RecordsPage } from "./types";

// Next.js and the Go API run in separate Docker containers, so both
// server- and client-side calls go through this same public base URL
// (Caddy reverse-proxies /api/* on that host to the api container).
export function getApiBaseUrl(): string {
  return process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
}

function formsUrl(path = ""): string {
  return `${getApiBaseUrl()}/api/form-builder/forms${path}`;
}

// ApiError はAPIのエラーレスポンス（{ error, fieldErrors? }）を保持する。
// fieldErrors のキーは項目IDの文字列。
export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly fieldErrors: Record<string, string> = {},
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function toApiError(response: Response, fallback: string): Promise<ApiError> {
  const body: unknown = await response.json().catch(() => null);
  if (body && typeof body === "object") {
    const { error, fieldErrors } = body as { error?: unknown; fieldErrors?: unknown };
    return new ApiError(
      typeof error === "string" && error ? error : fallback,
      response.status,
      fieldErrors && typeof fieldErrors === "object" ? (fieldErrors as Record<string, string>) : {},
    );
  }
  return new ApiError(fallback, response.status);
}

export async function fetchForms(init?: RequestInit): Promise<FormSummary[]> {
  const response = await fetch(formsUrl(), init);
  if (!response.ok) {
    throw await toApiError(response, "フォームの取得に失敗しました");
  }
  const data: { forms: FormSummary[] } = await response.json();
  return data.forms;
}

// fetchForm は存在しないフォームの場合 null を返す。
export async function fetchForm(id: number, init?: RequestInit): Promise<Form | null> {
  const response = await fetch(formsUrl(`/${id}`), init);
  if (response.status === 404) {
    return null;
  }
  if (!response.ok) {
    throw await toApiError(response, "フォームの取得に失敗しました");
  }
  return response.json();
}

export async function createForm(input: FormInput): Promise<Form> {
  const response = await fetch(formsUrl(), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw await toApiError(response, "フォームの作成に失敗しました");
  }
  return response.json();
}

export async function updateForm(id: number, input: FormInput): Promise<Form> {
  const response = await fetch(formsUrl(`/${id}`), {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw await toApiError(response, "フォームの更新に失敗しました");
  }
  return response.json();
}

export async function deleteForm(id: number): Promise<void> {
  const response = await fetch(formsUrl(`/${id}`), { method: "DELETE" });
  if (!response.ok) {
    throw await toApiError(response, "フォームの削除に失敗しました");
  }
}

export async function fetchRecords(
  formId: number,
  page: number,
  pageSize: number,
  init?: RequestInit,
): Promise<RecordsPage> {
  const params = new URLSearchParams({ page: String(page), pageSize: String(pageSize) });
  const response = await fetch(formsUrl(`/${formId}/records?${params.toString()}`), init);
  if (!response.ok) {
    throw await toApiError(response, "回答の取得に失敗しました");
  }
  return response.json();
}

export async function createRecord(formId: number, input: RecordInput): Promise<FormRecord> {
  const response = await fetch(formsUrl(`/${formId}/records`), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw await toApiError(response, "回答の送信に失敗しました");
  }
  return response.json();
}

export async function deleteRecord(formId: number, recordId: number): Promise<void> {
  const response = await fetch(formsUrl(`/${formId}/records/${recordId}`), { method: "DELETE" });
  if (!response.ok) {
    throw await toApiError(response, "回答の削除に失敗しました");
  }
}
