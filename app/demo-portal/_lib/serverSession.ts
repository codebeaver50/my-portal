import { cookies } from "next/headers";
import type { Session } from "./auth";

function getApiBaseUrl(): string {
  return process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
}

// Server Componentから呼ぶ、ログイン状態と権限の取得。ブラウザから来たCookieをGo APIの
// GET /api/auth/session へ中継する。取得に失敗した場合は null（権限が分からない）を返す。
// next/headers の cookies() を使うため、Client Component（"use client"）からは呼べない。
// Client Component からは ./auth の fetchSession や ./useSession を使う。
export async function fetchServerSession(): Promise<Session | null> {
  const cookieStore = await cookies();
  try {
    const response = await fetch(`${getApiBaseUrl()}/api/auth/session`, {
      headers: { Cookie: cookieStore.toString() },
      cache: "no-store",
    });
    if (!response.ok) return null;
    return await response.json();
  } catch (error) {
    console.error("auth: failed to fetch session on the server", error);
    return null;
  }
}
