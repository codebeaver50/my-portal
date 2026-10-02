import type { ReactNode } from "react";
import { redirect } from "next/navigation";
import { checkIsAdmin } from "@/app/_lib/adminAuth";
import { AdminHeader } from "../_components/AdminHeader";

// 記事投稿・カテゴリ管理は管理者専用。未ログインで直接URLを叩かれたら
// ポートフォリオ共通のログインページへ、戻り先をこのページにして送る。
export default async function SimpleCmsAdminLayout({ children }: { children: ReactNode }) {
  const isAdmin = await checkIsAdmin();
  if (!isAdmin) {
    redirect("/login?redirect=/demo-portal/simple-cms");
  }

  return (
    <>
      <AdminHeader />
      {children}
    </>
  );
}
