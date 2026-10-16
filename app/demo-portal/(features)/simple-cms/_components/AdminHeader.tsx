import { LinkButton, PageHeader } from "@/components/ui";
import { FeatureNav } from "./FeatureNav";

export function AdminHeader() {
  return (
    <>
      <PageHeader
        title="Simple CMS"
        description="カテゴリ分けとリッチテキスト編集ができる記事投稿機能です。画面の操作は誰でも試せますが、保存・削除には編集権限のあるアカウントでのログインが必要です。"
        aside={
          <LinkButton href="/demo-portal/simple-cms/new" className="w-auto whitespace-nowrap">
            記事を投稿
          </LinkButton>
        }
      />
      <FeatureNav />
    </>
  );
}
