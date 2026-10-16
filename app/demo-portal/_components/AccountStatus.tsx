"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Badge } from "@/components/ui";
import { fetchSession, logout } from "../_lib/auth";
import type { Session } from "../_lib/auth";
import { notifySessionChange } from "../_lib/useSession";

const loginPath = "/demo-portal/login";
const accessControlPath = "/demo-portal/access-control";

// デモポータルのヘッダーに出す、ログイン中のユーザーとロール（未ログインならゲスト）とログイン/ログアウトの導線。
export function AccountStatus() {
  const router = useRouter();
  const pathname = usePathname();
  const [session, setSession] = useState<Session | null>(null);
  const [isPending, setIsPending] = useState(false);

  useEffect(() => {
    // ログインページからはソフトナビゲーション（router.push）で戻ってくるためマウントされ直さない。
    // pathname依存にして遷移のたびに再取得する。
    let ignore = false;
    fetchSession().then((result) => {
      if (!ignore) setSession(result);
    });
    return () => {
      ignore = true;
    };
  }, [pathname]);

  async function handleLogout() {
    setIsPending(true);
    try {
      await logout();
      // 遷移が終わるまでのあいだも、表示中のページの書き込みボタン等をゲストの状態にする
      notifySessionChange();
      // ログアウト後はログイン画面へ。ログインし直したら元のページへ戻れるよう redirect を付ける
      // （ヘッダーのログイン状態は pathname の変化で取得し直される）
      router.push(pathname === loginPath ? loginPath : `${loginPath}?redirect=${encodeURIComponent(pathname)}`);
    } finally {
      setIsPending(false);
    }
  }

  // 取得前・取得失敗時は何も出さない（ゲストと誤表示しないため）
  if (!session) return null;

  const linkClassName = "text-sm text-muted-foreground transition-colors hover:text-foreground hover:underline";

  if (!session.user) {
    return (
      <div className="flex items-center gap-3">
        <Badge variant="neutral">ゲスト</Badge>
        {pathname !== loginPath ? (
          <Link href={`${loginPath}?redirect=${encodeURIComponent(pathname)}`} className={linkClassName}>
            ログイン
          </Link>
        ) : null}
      </div>
    );
  }

  return (
    <div className="flex items-center gap-3">
      {session.permissions["access-control"] !== "none" && pathname !== accessControlPath ? (
        <Link href={accessControlPath} className={linkClassName}>
          アクセス管理
        </Link>
      ) : null}
      <div className="flex min-w-0 flex-col items-end gap-1 sm:flex-row sm:items-center sm:gap-2">
        <span className="truncate text-sm font-semibold">{session.user.displayName}</span>
        <span className="flex gap-1">
          {session.user.roles.map((role) => (
            <Badge key={role.code} variant="accent" dot={false}>
              {role.name}
            </Badge>
          ))}
        </span>
      </div>
      <button type="button" onClick={handleLogout} disabled={isPending} className={`${linkClassName} disabled:opacity-60`}>
        {isPending ? "..." : "ログアウト"}
      </button>
    </div>
  );
}
