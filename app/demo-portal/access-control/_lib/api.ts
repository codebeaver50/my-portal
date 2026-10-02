// アクセス管理画面のAPI（Go API の /api/access-control）のクライアント。Client Componentから呼ぶ。
import type { PermissionLevel } from "../../_lib/auth";

export type ResourceDefinition = {
  resource: string;
  // そのリソースに設定できる最大の権限（書き込み操作のない shop-css は read）
  maxLevel: Exclude<PermissionLevel, "none">;
};

export type Role = {
  id: number;
  code: string;
  name: string;
  description: string;
  isSystem: boolean;
  // false のロール（admin）は権限を変更できない
  editable: boolean;
  // 全リソースのキーを含む
  permissions: Record<string, PermissionLevel>;
};

export type RolesResponse = {
  resources: ResourceDefinition[];
  roles: Role[];
};

// ApiError はAPIのエラーレスポンス（{ error }）とHTTPステータスを保持する。
// 401（未ログイン）・403（権限なし）を画面側で出し分けるためにステータスを持たせる。
export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

function getApiBaseUrl(): string {
  return process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
}

async function toApiError(response: Response, fallback: string): Promise<ApiError> {
  const body = await response.json().catch(() => null);
  const message = body && typeof body.error === "string" && body.error ? body.error : fallback;
  return new ApiError(message, response.status);
}

export async function fetchRoles(): Promise<RolesResponse> {
  const response = await fetch(`${getApiBaseUrl()}/api/access-control/roles`, {
    credentials: "include",
    cache: "no-store",
  });
  if (!response.ok) {
    throw await toApiError(response, "ロールの取得に失敗しました");
  }
  return response.json();
}

// updateRolePermissions はロールの権限をまるごと置き換え、更新後のロールを返す。
export async function updateRolePermissions(roleId: number, permissions: Record<string, PermissionLevel>): Promise<Role> {
  const response = await fetch(`${getApiBaseUrl()}/api/access-control/roles/${roleId}/permissions`, {
    method: "PUT",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ permissions }),
  });
  if (!response.ok) {
    throw await toApiError(response, "権限の保存に失敗しました");
  }
  return response.json();
}
