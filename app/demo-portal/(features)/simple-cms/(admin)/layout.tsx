import type { ReactNode } from "react";
import { AdminHeader } from "../_components/AdminHeader";

// 記事投稿・カテゴリ管理の画面は誰でも開いて操作を試せる。保存・削除のボタンだけを、
// simple-cms の write 権限がない場合に押せなくする（各コンポーネントの useWriteBlockedReason）。
export default function SimpleCmsAdminLayout({ children }: { children: ReactNode }) {
  return (
    <>
      <AdminHeader />
      {children}
    </>
  );
}
