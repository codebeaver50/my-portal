import { z } from "zod";

const journalEntrySchema = z.object({
  chartOfAccountsId: z.number({ error: "勘定科目を選択してください" }).min(1, "勘定科目を選択してください"),
  type: z.enum(["debit", "credit"], { error: "借方または貸方を選択してください" }),
  amount: z.number().gt(0, "0より大きい金額を入力してください"),
  description: z.string().max(100, "説明は100文字以内で入力してください").default(""),
});

export const transactionSchema = z
  .object({
    date: z.string().min(1, "日付を入力してください"),
    description: z.string().max(100, "説明は100文字以内で入力してください").default(""),
    journalEntries: z
      .array(journalEntrySchema)
      .min(2, "取引には最低1つの借方明細と1つの貸方明細が必要です"),
  })
  .refine(
    (data) => data.journalEntries.some((entry) => entry.type === "debit"),
    { message: "借方の明細が必要です", path: ["journalEntries"] },
  )
  .refine(
    (data) => data.journalEntries.some((entry) => entry.type === "credit"),
    { message: "貸方の明細が必要です", path: ["journalEntries"] },
  )
  .refine(
    (data) => {
      const debitTotal = data.journalEntries
        .filter((entry) => entry.type === "debit")
        .reduce((sum, entry) => sum + entry.amount, 0);
      const creditTotal = data.journalEntries
        .filter((entry) => entry.type === "credit")
        .reduce((sum, entry) => sum + entry.amount, 0);
      return debitTotal === creditTotal;
    },
    { message: "複式簿記のルール: 借方合計と貸方合計が一致していません", path: ["journalEntries"] },
  );

export type TransactionFormData = z.infer<typeof transactionSchema>;
