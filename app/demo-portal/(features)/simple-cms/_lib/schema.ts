import { z } from "zod";

// slugはURLパスセグメントになるため、小文字英数字とハイフン区切りのみ許可する
// （Go API側のバリデーションと同じ形式）。
const slugSchema = z
  .string()
  .min(1, "URLスラッグを入力してください")
  .max(255, "URLスラッグは255文字以内で入力してください")
  .regex(/^[a-z0-9]+(-[a-z0-9]+)*$/, "小文字英数字とハイフンのみ使用できます（例: my-post）");

export const postSchema = z.object({
  categoryId: z.number({ error: "カテゴリを選択してください" }).min(1, "カテゴリを選択してください"),
  title: z.string().min(1, "タイトルを入力してください").max(255, "タイトルは255文字以内で入力してください"),
  slug: slugSchema,
  content: z.string().min(1, "本文を入力してください"),
});

export type PostFormData = z.infer<typeof postSchema>;

export const categorySchema = z.object({
  name: z.string().min(1, "カテゴリ名を入力してください").max(100, "カテゴリ名は100文字以内で入力してください"),
  slug: slugSchema.max(100, "URLスラッグは100文字以内で入力してください"),
});

export type CategoryFormData = z.infer<typeof categorySchema>;

// 新規記事のURLスラッグの初期値（例: 20260922-143512）。タイトルは日本語に
// なることが多くタイトルからの自動スラッグ化が機能しないため、日付+時刻を
// デフォルト値として入れておき、必要ならユーザーが上書きする。
export function generateDefaultPostSlug(now = new Date()): string {
  const pad = (n: number) => String(n).padStart(2, "0");
  const date = `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}`;
  const time = `${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}`;
  return `${date}-${time}`;
}
