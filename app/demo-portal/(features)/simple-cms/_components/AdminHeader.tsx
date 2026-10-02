import { LinkButton, PageHeader } from "@/components/ui";
import { FeatureNav } from "./FeatureNav";

export function AdminHeader() {
  return (
    <>
      <PageHeader
        title="Simple CMS"
        description="カテゴリ分けとリッチテキスト編集ができる記事投稿機能です。管理者としてログイン中です。"
        aside={
          <LinkButton href="/demo-portal/simple-cms/new" className="w-auto">
            記事を投稿
          </LinkButton>
        }
      />
      <FeatureNav />
    </>
  );
}
