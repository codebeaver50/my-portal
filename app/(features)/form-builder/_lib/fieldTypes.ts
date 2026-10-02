import type { FieldType } from "./types";

export const FIELD_TYPE_LABEL: Record<FieldType, string> = {
  text: "テキスト（1行）",
  textarea: "テキスト（複数行）",
  number: "数値",
  email: "メールアドレス",
  date: "日付",
  select: "プルダウン",
  radio: "ラジオボタン",
  checkbox: "チェックボックス",
};

// Go側の models.FieldType.HasOptions と対応する。
export function hasOptions(type: FieldType): boolean {
  return type === "select" || type === "radio" || type === "checkbox";
}
