"use client";

import { useEffect, useState } from "react";
import { fetchSession } from "./auth";
import type { Session } from "./auth";

// ログイン・ログアウトしたことを、同じページ内のほかのコンポーネント（各featureの書き込みボタン等）に知らせるイベント。
const sessionChangeEvent = "portal:session-change";

export function notifySessionChange() {
  window.dispatchEvent(new Event(sessionChangeEvent));
}

// ログイン状態と権限を取得する。取得中は undefined、取得に失敗した場合は null を返す。
// ヘッダーからログアウトしたとき（notifySessionChange）にも取得し直す。
export function useSession(): Session | null | undefined {
  const [session, setSession] = useState<Session | null | undefined>(undefined);

  useEffect(() => {
    let ignore = false;
    function load() {
      fetchSession().then((result) => {
        if (!ignore) setSession(result);
      });
    }
    load();
    window.addEventListener(sessionChangeEvent, load);
    return () => {
      ignore = true;
      window.removeEventListener(sessionChangeEvent, load);
    };
  }, []);

  return session;
}

// resource（feature-slug）への書き込み（保存・削除等）ができない理由を返す。write 権限があれば undefined。
// action は「保存」「削除」のような操作の名前で、「保存するにはログインが必要です」のような文になる。
// Button の blockedReason に渡して、押せないボタンとツールチップで理由を出す想定。
// 画面の表示を切り替えるためのもので、実際の権限チェックはGo API側で行う。
export function useWriteBlockedReason(resource: string, action: string): string | undefined {
  const session = useSession();
  if (session === undefined) return "権限を確認しています";
  if (session === null) return "権限を確認できませんでした";
  if (session.permissions[resource] === "write") return undefined;
  return session.user ? `${action}する権限がありません` : `${action}するにはログインが必要です`;
}
