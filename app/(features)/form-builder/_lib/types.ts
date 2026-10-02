// Go API（server/internal/features/formbuilder/dto）のリクエスト/レスポンス型と1対1で対応させる。

export const FIELD_TYPES = [
  "text",
  "textarea",
  "number",
  "email",
  "date",
  "select",
  "radio",
  "checkbox",
] as const;

export type FieldType = (typeof FIELD_TYPES)[number];

export type FormField = {
  id: number;
  label: string;
  type: FieldType;
  required: boolean;
  options: string[];
  // 表示する行（0始まり）。同じ行の項目は横並びで表示する。
  row: number;
  // 1行を12等分した単位での幅（3〜12）。同じ行の合計は12以内。
  width: number;
};

export type Form = {
  id: number;
  title: string;
  description: string;
  fields: FormField[];
  createdAt: string;
  updatedAt: string;
};

export type FormSummary = {
  id: number;
  title: string;
  description: string;
  fieldCount: number;
  recordCount: number;
  createdAt: string;
  updatedAt: string;
};

export type FormFieldInput = {
  // 既存項目の更新時のみ指定する。未指定の項目は新規作成される。
  id?: number;
  label: string;
  type: FieldType;
  required: boolean;
  options: string[];
  row: number;
  width: number;
};

// fields は表示順（行ごとに左から右）に並べ、row は先頭から昇順にする。
export type FormInput = {
  title: string;
  description: string;
  fields: FormFieldInput[];
};

// text/textarea/email/date/select/radio は string、number は number、checkbox は string[]。
export type RecordValue = string | number | string[];

// キーは項目IDの文字列。未入力の任意項目はキー自体が存在しない。
export type RecordValues = Record<string, RecordValue>;

export type FormRecord = {
  id: number;
  formId: number;
  values: RecordValues;
  createdAt: string;
};

export type RecordInput = {
  values: RecordValues;
};

export type RecordsPage = {
  records: FormRecord[];
  total: number;
  page: number;
  pageSize: number;
  hasNextPage: boolean;
};
