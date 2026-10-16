import type { ReactNode } from "react";
import { redirect } from "next/navigation";
import { Card } from "@/components/ui";
import { fetchServerSession } from "@/app/demo-portal/_lib/serverSession";
import { AdminHeader } from "../_components/AdminHeader";

// 記事投稿・カテゴリ管理には、アクセス管理（RBAC）で simple-cms の write 権限が必要。
// 未ログインで直接URLを叩かれたら、デモポータルのログインページへ、戻り先を記事一覧にして送る。
// ログイン済みで権限がない場合は、その旨を表示する。
export default async function SimpleCmsAdminLayout({ children }: { children: ReactNode }) {
  const session = await fetchServerSession();
  if (!session?.user) {
    redirect("/demo-portal/login?redirect=/demo-portal/simple-cms");
  }
  if (session.permissions["simple-cms"] !== "write") {
    return (
      <Card>
        <p className="text-sm text-muted-foreground">記事の投稿・カテゴリの管理を行う権限がありません。</p>
      </Card>
    );
  }

  return (
    <>
      <AdminHeader />
      {children}
    </>
  );
}
