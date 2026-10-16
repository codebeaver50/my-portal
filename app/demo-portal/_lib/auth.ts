// デモポータル共通のアクセス管理（Go API の /api/auth）のクライアント。
// Client Componentから呼ぶ想定（ログインフォーム・ヘッダーのログイン状態表示等）。

export type PermissionLevel = "none" | "read" | "write";

export type SessionUser = {
  id: number;
  loginId: string;
  displayName: string;
  roles: { code: string; name: string }[];
};

// 未ログインのときは user が null で、permissions はゲストロールの権限になる。
// permissions のキーはリソース（feature-slug または access-control）。
export type Session = {
  user: SessionUser | null;
  permissions: Record<string, PermissionLevel>;
};

function getApiBaseUrl(): string {
  return process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
}

async function parseErrorMessage(response: Response, fallback: string): Promise<string> {
  const body = await response.json().catch(() => null);
  return (body && typeof body.error === "string" && body.error) || fallback;
}

export async function login(loginId: string, password: string): Promise<Session> {
  const response = await fetch(`${getApiBaseUrl()}/api/auth/login`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ loginId, password }),
  });
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "ログインに失敗しました"));
  }
  return response.json();
}

export async function logout(): Promise<void> {
  const response = await fetch(`${getApiBaseUrl()}/api/auth/logout`, {
    method: "POST",
    credentials: "include",
  });
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "ログアウトに失敗しました"));
  }
}

// ブラウザからログイン状態を取得する。取得に失敗した場合は null（ログイン状態が分からない）を返す。
export async function fetchSession(): Promise<Session | null> {
  try {
    const response = await fetch(`${getApiBaseUrl()}/api/auth/session`, {
      credentials: "include",
      cache: "no-store",
    });
    if (!response.ok) return null;
    return await response.json();
  } catch (error) {
    console.error("auth: failed to fetch session", error);
    return null;
  }
}

// ログイン後に戻る先として、外部URLへのオープンリダイレクトを避けるため "/" 始まりの相対パスのみ許可する。
export function resolveRedirectTo(value: string | string[] | undefined, fallback = "/demo-portal"): string {
  if (typeof value === "string" && value.startsWith("/") && !value.startsWith("//")) {
    return value;
  }
  return fallback;
}

// ログイン画面のフォームの下に載せる、訪問者が試せるデモ用アカウント（server/migrations のシードデータと揃える）。
export const demoAccount = { loginId: "demo-guest", password: "demo-pass" } as const;
