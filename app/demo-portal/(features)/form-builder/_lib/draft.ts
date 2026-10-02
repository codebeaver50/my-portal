import { FIELD_TYPE_LABEL, hasOptions } from "./fieldTypes";
import { GRID_COLUMNS, toRows } from "./layout";
import type { FieldType, Form, FormInput } from "./types";

// FieldDraft はフォーム構築画面で編集中の項目。key はReactのkeyとドラッグ対象の識別用
// （保存済み項目のIDとは別）、選択肢は1行1つのテキストとして編集する。
export type FieldDraft = {
  key: string;
  id?: number;
  label: string;
  type: FieldType;
  required: boolean;
  optionsText: string;
  // 1行を12等分した単位での幅。
  width: number;
};

export const DEFAULT_OPTIONS_TEXT = "選択肢1\n選択肢2";

export function parseOptions(optionsText: string): string[] {
  return optionsText
    .split("\n")
    .map((option) => option.trim())
    .filter((option) => option !== "");
}

// newDraft はパーツから追加した直後の項目。そのまま保存してもエラーにならない初期値にする。
// 幅は1行全体とし、既存の行に入れた場合は insertItem が行に収まるよう調整する。
export function newDraft(key: string, type: FieldType): FieldDraft {
  return {
    key,
    label: FIELD_TYPE_LABEL[type],
    type,
    required: false,
    optionsText: hasOptions(type) ? DEFAULT_OPTIONS_TEXT : "",
    width: GRID_COLUMNS,
  };
}

export function draftRowsFromForm(form: Form): FieldDraft[][] {
  return toRows(form.fields).map((row) =>
    row.map((field) => ({
      key: `field-${field.id}`,
      id: field.id,
      label: field.label,
      type: field.type,
      required: field.required,
      optionsText: field.options.join("\n"),
      width: field.width,
    })),
  );
}

export function toFieldInput(draft: FieldDraft, row: number): FormInput["fields"][number] {
  return {
    id: draft.id,
    label: draft.label,
    type: draft.type,
    required: draft.required,
    options: hasOptions(draft.type) ? parseOptions(draft.optionsText) : [],
    row,
    width: draft.width,
  };
}

// toFieldInputs は行ごとの配列を、APIに送る表示順の平らな配列にする。
export function toFieldInputs(rows: FieldDraft[][]): FormInput["fields"] {
  return rows.flatMap((row, rowIndex) => row.map((draft) => toFieldInput(draft, rowIndex)));
}
