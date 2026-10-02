import type { Metadata } from "next";
import { FormListView } from "./_components/FormListView";

export const metadata: Metadata = {
  title: "カスタムフォームビルダー | My Portal",
  description: "フォームを作成して回答を集め、データベースとして一覧できるデモアプリ",
};

export default function FormBuilderPage() {
  return <FormListView />;
}
