import { PageContainer, PageHeader } from "@/components/ui";
import { Providers } from "./_components/Providers";
import { TabNav } from "./_components/TabNav";

export default function FormBuilderLayout({ children }: LayoutProps<"/form-builder">) {
  return (
    <Providers>
      <PageContainer>
        <PageHeader
          title="カスタムフォームビルダー"
          description="項目を組み合わせてフォームを作成し、集まった回答をデータベースとして一覧できるデモアプリです。閲覧中のフォームと回答は誰でも編集・削除できる共有データです。"
        />
        <TabNav
          items={[
            { href: "/form-builder", label: "フォーム一覧" },
            { href: "/form-builder/new", label: "フォームを作成" },
          ]}
        />
        {children}
      </PageContainer>
    </Providers>
  );
}
