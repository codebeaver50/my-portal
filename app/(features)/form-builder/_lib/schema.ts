import { z } from "zod";
import { hasOptions } from "./fieldTypes";
import { MAX_FIELDS_PER_ROW, toRows } from "./layout";
import { FIELD_TYPES } from "./types";
import type { FormField, RecordValue, RecordValues } from "./types";

// 制約値は Go側（formbuilder/dto・service）のバリデーションと揃える。
export const MAX_FIELDS = 50;
const MAX_OPTIONS = 50;
const MAX_OPTION_LENGTH = 100;
const MAX_TEXT_LENGTH = 255;
const MAX_TEXTAREA_LENGTH = 2000;
const MAX_EMAIL_LENGTH = 254;
const MAX_NUMBER_ABS = 1e15;

const REQUIRED_MESSAGE = "この項目は必須です";

export const formFieldSchema = z
  .object({
    id: z.number().int().positive().optional(),
    label: z
      .string()
      .trim()
      .min(1, "項目名を入力してください")
      .max(100, "項目名は100文字以内で入力してください"),
    type: z.enum(FIELD_TYPES),
    required: z.boolean(),
    options: z.array(
      z.string().trim().min(1).max(MAX_OPTION_LENGTH, `選択肢は${MAX_OPTION_LENGTH}文字以内で入力してください`),
    ),
    row: z.number().int().min(0),
  })
  .superRefine((field, ctx) => {
    if (!hasOptions(field.type)) return;
    if (field.options.length === 0) {
      ctx.addIssue({ code: "custom", path: ["options"], message: "選択肢を1つ以上入力してください" });
    } else if (field.options.length > MAX_OPTIONS) {
      ctx.addIssue({ code: "custom", path: ["options"], message: `選択肢は${MAX_OPTIONS}個以内にしてください` });
    } else if (new Set(field.options).size !== field.options.length) {
      ctx.addIssue({ code: "custom", path: ["options"], message: "選択肢が重複しています" });
    }
  })
  .transform((field) => (hasOptions(field.type) ? field : { ...field, options: [] }));

export const formSchema = z.object({
  title: z
    .string()
    .trim()
    .min(1, "フォーム名を入力してください")
    .max(100, "フォーム名は100文字以内で入力してください"),
  description: z.string().trim().max(500, "説明は500文字以内で入力してください"),
  fields: z
    .array(formFieldSchema)
    .min(1, "項目を1つ以上追加してください")
    .max(MAX_FIELDS, `項目は${MAX_FIELDS}個以内にしてください`)
    .refine(
      (fields) => toRows(fields).every((row) => row.length <= MAX_FIELDS_PER_ROW),
      `1行に並べられる項目は${MAX_FIELDS_PER_ROW}個までです`,
    ),
});

// 回答画面の入力状態。checkbox は選択中の選択肢、それ以外は入力欄の文字列。
export type AnswerDraft = Record<string, string | string[]>;

function requiredError(fallback: string) {
  return (issue: { input?: unknown }) => (issue.input === undefined ? REQUIRED_MESSAGE : fallback);
}

// toRawValue は入力状態の値を検証前の値に変換する。未入力は undefined にする。
function toRawValue(field: FormField, draft: string | string[] | undefined): unknown {
  if (Array.isArray(draft)) {
    return draft.length > 0 ? draft : undefined;
  }
  if (draft === undefined || draft.trim() === "") {
    return undefined;
  }
  if (field.type === "number") {
    return Number(draft);
  }
  return field.type === "textarea" ? draft : draft.trim();
}

function valueSchema(field: FormField): z.ZodType<RecordValue> {
  switch (field.type) {
    case "text":
      return z
        .string({ error: requiredError("文字列を入力してください") })
        .max(MAX_TEXT_LENGTH, `${MAX_TEXT_LENGTH}文字以内で入力してください`);
    case "textarea":
      return z
        .string({ error: requiredError("文字列を入力してください") })
        .max(MAX_TEXTAREA_LENGTH, `${MAX_TEXTAREA_LENGTH}文字以内で入力してください`);
    case "email":
      return z
        .email({ error: requiredError("メールアドレスの形式が正しくありません") })
        .max(MAX_EMAIL_LENGTH, "メールアドレスの形式が正しくありません");
    case "date":
      return z.iso.date({ error: requiredError("日付の形式が正しくありません") });
    case "number":
      return z
        .number({ error: requiredError("数値を入力してください") })
        .refine((value) => Math.abs(value) <= MAX_NUMBER_ABS, "入力できる範囲を超えています");
    case "select":
    case "radio":
      return z.enum(field.options as [string, ...string[]], { error: requiredError("選択肢から選んでください") });
    case "checkbox":
      return z.array(z.enum(field.options as [string, ...string[]]), {
        error: requiredError("選択肢から選んでください"),
      });
  }
}

export type AnswerValidationResult =
  | { success: true; values: RecordValues }
  | { success: false; fieldErrors: Record<string, string> };

// validateAnswer はフォームの項目定義から回答のスキーマを組み立てて入力内容を検証し、
// APIに送る値（未入力の任意項目は除外）を返す。
export function validateAnswer(fields: FormField[], draft: AnswerDraft): AnswerValidationResult {
  const values: RecordValues = {};
  const fieldErrors: Record<string, string> = {};

  for (const field of fields) {
    const key = String(field.id);
    const raw = toRawValue(field, draft[key]);
    if (raw === undefined && !field.required) {
      continue;
    }

    const parsed = valueSchema(field).safeParse(raw);
    if (parsed.success) {
      values[key] = parsed.data;
    } else {
      fieldErrors[key] = parsed.error.issues[0]?.message ?? "入力内容を確認してください";
    }
  }

  return Object.keys(fieldErrors).length > 0 ? { success: false, fieldErrors } : { success: true, values };
}
