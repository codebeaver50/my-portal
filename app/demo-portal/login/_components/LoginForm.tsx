"use client";

import { useState } from "react";
import type { FormEvent } from "react";
import { useRouter } from "next/navigation";
import { Button, Card, Input } from "@/components/ui";
import { demoAccount, login } from "../../_lib/auth";

export function LoginForm({ redirectTo }: { redirectTo: string }) {
  const router = useRouter();
  const [loginId, setLoginId] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isPending, setIsPending] = useState(false);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setError(null);
    setIsPending(true);
    try {
      await login(loginId, password);
      // Server Componentがログイン状態を再判定できるよう、遷移後にキャッシュを破棄する。
      router.push(redirectTo);
      router.refresh();
    } catch (loginError) {
      setError(loginError instanceof Error ? loginError.message : "ログインに失敗しました");
      setIsPending(false);
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-sm flex-col gap-6">
      <Card>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <h1 className="text-lg font-bold">ログイン</h1>
          <Input
            id="login-id"
            label="ログインID"
            value={loginId}
            onChange={(e) => setLoginId(e.target.value)}
            autoComplete="username"
            maxLength={50}
            required
            autoFocus
          />
          <Input
            id="login-password"
            label="パスワード"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="current-password"
            maxLength={72}
            required
          />
          {error ? (
            <p role="alert" className="text-sm text-danger">
              {error}
            </p>
          ) : null}
          <Button type="submit" disabled={isPending}>
            {isPending ? "ログイン中..." : "ログイン"}
          </Button>
        </form>
      </Card>

      <section aria-labelledby="demo-accounts-heading" className="flex flex-col gap-2">
        <h2 id="demo-accounts-heading" className="text-sm font-bold">
          デモ用アカウント
        </h2>
        <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 rounded-lg border border-border bg-card px-4 py-3 text-sm">
          <dt className="text-xs font-semibold text-muted-foreground">ログインID</dt>
          <dd className="font-mono">{demoAccount.loginId}</dd>
          <dt className="text-xs font-semibold text-muted-foreground">パスワード</dt>
          <dd className="font-mono">{demoAccount.password}</dd>
        </dl>
      </section>
    </div>
  );
}
