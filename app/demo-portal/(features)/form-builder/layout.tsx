import { PageContainer, PageHeader } from "@/components/ui";
import { Providers } from "./_components/Providers";
import { TabNav } from "./_components/TabNav";

export default function FormBuilderLayout({ children }: LayoutProps<"/demo-portal/form-builder">) {
  return (
    <Providers>
      <PageContainer>
        <PageHeader
          title="カスタムフォームビルダー"
          description="項目を組み合わせてフォームを作成し、集まった回答をデータベースとして一覧できるデモアプリです。画面の操作は誰でも試せますが、保存・削除には編集権限のあるアカウントでのログインが必要です。"
        />
        <TabNav
          items={[
            { href: "/demo-portal/form-builder", label: "フォーム一覧" },
            { href: "/demo-portal/form-builder/new", label: "フォームを作成" },
          ]}
        />
        {children}
      </PageContainer>
    </Providers>
  );
}
