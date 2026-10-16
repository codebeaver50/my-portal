import type { Metadata } from "next";
import { PageContainer, PageHeader } from "@/components/ui";
import { RolePermissionsEditor } from "./_components/RolePermissionsEditor";

export const metadata: Metadata = {
  title: "アクセス管理 | DemoPortal | Code Beaver",
};

export default function AccessControlPage() {
  return (
    <PageContainer>
      <PageHeader
        title="アクセス管理"
        description="ロールごとに、各デモアプリとアクセス管理画面に対する権限（なし・閲覧・編集）を設定します。"
      />
      <RolePermissionsEditor />
    </PageContainer>
  );
}
