import { z } from "zod";

export const contactSchema = z.object({
  name: z.string().min(1, "お名前を入力してください").max(100, "お名前は100文字以内で入力してください"),
  email: z
    .string()
    .min(1, "メールアドレスを入力してください")
    .email("メールアドレスの形式が正しくありません"),
  message: z
    .string()
    .min(1, "お問い合わせ内容を入力してください")
    .max(2000, "お問い合わせ内容は2000文字以内で入力してください"),
});

export type ContactFormData = z.infer<typeof contactSchema>;
