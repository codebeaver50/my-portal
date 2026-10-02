import type { Metadata } from "next";
import { FormBuilder } from "../_components/FormBuilder";

export const metadata: Metadata = {
  title: "フォームを作成 | カスタムフォームビルダー | My Portal",
};

export default function NewFormPage() {
  return <FormBuilder />;
}
