"use client";

import { Checkbox, Input, Radio, Select, Textarea } from "@/components/ui";
import type { FormField } from "../_lib/types";

const INPUT_TYPE: Partial<Record<FormField["type"], string>> = {
  text: "text",
  number: "number",
  email: "email",
  date: "date",
};

type FieldInputProps = {
  field: Pick<FormField, "label" | "type" | "required" | "options">;
  id: string;
  value: string | string[];
  onChange: (value: string | string[]) => void;
  error?: string;
};

function RequiredMark() {
  return (
    <span className="ml-1 text-danger" aria-label="必須">
      *
    </span>
  );
}

// FieldInput は項目定義に応じた入力欄を描画する。回答画面とフォーム構築画面の
// プレビューで共通に使う。
export function FieldInput({ field, id, value, onChange, error }: FieldInputProps) {
  const label = field.label.trim() || "（項目名未設定）";

  if (field.type === "radio" || field.type === "checkbox") {
    const selected = Array.isArray(value) ? value : [value];
    return (
      <fieldset className="flex flex-col gap-2" aria-describedby={error ? `${id}-error` : undefined}>
        <legend className="mb-1.5 text-sm font-medium">
          {label}
          {field.required ? <RequiredMark /> : null}
        </legend>
        <div className="flex flex-wrap gap-x-5 gap-y-2">
          {field.options.map((option) =>
            field.type === "radio" ? (
              <Radio
                key={option}
                name={id}
                label={option}
                checked={value === option}
                onChange={() => onChange(option)}
              />
            ) : (
              <Checkbox
                key={option}
                label={option}
                checked={selected.includes(option)}
                onChange={(event) =>
                  onChange(
                    event.target.checked
                      ? [...selected, option]
                      : selected.filter((current) => current !== option),
                  )
                }
              />
            ),
          )}
        </div>
        {error ? (
          <span id={`${id}-error`} className="text-xs text-danger">
            {error}
          </span>
        ) : null}
      </fieldset>
    );
  }

  const text = typeof value === "string" ? value : "";

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="text-sm font-medium">
        {label}
        {field.required ? <RequiredMark /> : null}
      </label>
      {field.type === "textarea" ? (
        <Textarea id={id} rows={4} value={text} error={error} onChange={(event) => onChange(event.target.value)} />
      ) : field.type === "select" ? (
        <Select id={id} value={text} error={error} onChange={(event) => onChange(event.target.value)}>
          <option value="">選択してください</option>
          {field.options.map((option) => (
            <option key={option} value={option}>
              {option}
            </option>
          ))}
        </Select>
      ) : (
        <Input
          id={id}
          type={INPUT_TYPE[field.type]}
          inputMode={field.type === "number" ? "decimal" : undefined}
          step={field.type === "number" ? "any" : undefined}
          value={text}
          error={error}
          onChange={(event) => onChange(event.target.value)}
        />
      )}
    </div>
  );
}
