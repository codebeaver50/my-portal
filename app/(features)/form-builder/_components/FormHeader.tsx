import Link from "next/link";
import { TabNav } from "./TabNav";
import type { Form } from "../_lib/types";

// FormHeader は個別フォームのページ（回答・回答データ・編集）共通の見出しとタブ。
export function FormHeader({ form }: { form: Form }) {
  const base = `/form-builder/${form.id}`;

  return (
    <div className="flex flex-col gap-4">
      <Link href="/form-builder" className="text-sm text-muted-foreground hover:text-foreground">
        ← フォーム一覧に戻る
      </Link>
      <div className="flex flex-col gap-1">
        <h2 className="text-xl font-bold break-words">{form.title}</h2>
        {form.description ? <p className="text-sm text-muted-foreground">{form.description}</p> : null}
      </div>
      <TabNav
        items={[
          { href: base, label: "回答する" },
          { href: `${base}/records`, label: "回答データ" },
          { href: `${base}/edit`, label: "フォームを編集" },
        ]}
      />
    </div>
  );
}
