"use client";

import { useId, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { Button, Card, Checkbox, Input, Select, Textarea, useToast } from "@/components/ui";
import { FIELD_TYPE_LABEL, hasOptions } from "../_lib/fieldTypes";
import { formSchema } from "../_lib/schema";
import { useCreateForm, useUpdateForm } from "../_lib/useForms";
import { FIELD_TYPES } from "../_lib/types";
import { FieldInput } from "./FieldInput";
import { ArrowDownIcon, ArrowUpIcon, PlusIcon, TrashIcon } from "./icons";
import type { FieldType, Form, FormInput } from "../_lib/types";

// FieldDraft は編集中の項目。key はReactのkey用（保存済み項目のIDとは別）、
// 選択肢は1行1つのテキストとして編集する。
type FieldDraft = {
  key: string;
  id?: number;
  label: string;
  type: FieldType;
  required: boolean;
  optionsText: string;
};

type FieldErrors = { label?: string; options?: string };

type BuilderErrors = {
  title?: string;
  description?: string;
  fields?: string;
  byField: Record<string, FieldErrors>;
};

const EMPTY_ERRORS: BuilderErrors = { byField: {} };

function parseOptions(optionsText: string): string[] {
  return optionsText
    .split("\n")
    .map((option) => option.trim())
    .filter((option) => option !== "");
}

function toDraftInput(drafts: FieldDraft[]): FormInput["fields"] {
  return drafts.map((draft) => ({
    id: draft.id,
    label: draft.label,
    type: draft.type,
    required: draft.required,
    options: hasOptions(draft.type) ? parseOptions(draft.optionsText) : [],
  }));
}

type FormBuilderProps = {
  form?: Form;
};

export function FormBuilder({ form }: FormBuilderProps) {
  const isEdit = Boolean(form);
  const router = useRouter();
  const { showToast } = useToast();
  const createMutation = useCreateForm();
  const updateMutation = useUpdateForm();
  const idPrefix = useId();

  const nextKey = useRef(0);
  const newKey = () => `draft-${nextKey.current++}`;

  const [title, setTitle] = useState(form?.title ?? "");
  const [description, setDescription] = useState(form?.description ?? "");
  const [drafts, setDrafts] = useState<FieldDraft[]>(() =>
    form
      ? form.fields.map((field) => ({
          key: `field-${field.id}`,
          id: field.id,
          label: field.label,
          type: field.type,
          required: field.required,
          optionsText: field.options.join("\n"),
        }))
      : [{ key: "draft-initial", label: "", type: "text", required: false, optionsText: "" }],
  );
  const [errors, setErrors] = useState<BuilderErrors>(EMPTY_ERRORS);
  const [previewValues, setPreviewValues] = useState<Record<string, string | string[]>>({});

  const isPending = createMutation.isPending || updateMutation.isPending;

  function updateDraft(key: string, patch: Partial<FieldDraft>) {
    setDrafts((current) => current.map((draft) => (draft.key === key ? { ...draft, ...patch } : draft)));
    // 編集した項目のエラー表示は、次に保存するまで消しておく。
    setErrors((current) => {
      if (!current.byField[key]) return current;
      const byField = { ...current.byField };
      delete byField[key];
      return { ...current, byField };
    });
  }

  function moveDraft(index: number, offset: -1 | 1) {
    setDrafts((current) => {
      const target = index + offset;
      if (target < 0 || target >= current.length) return current;
      const next = [...current];
      [next[index], next[target]] = [next[target], next[index]];
      return next;
    });
  }

  function removeDraft(key: string) {
    setDrafts((current) => current.filter((draft) => draft.key !== key));
  }

  function addDraft() {
    setDrafts((current) => [
      ...current,
      { key: newKey(), label: "", type: "text", required: false, optionsText: "" },
    ]);
  }

  function handleSubmit() {
    const parsed = formSchema.safeParse({ title, description, fields: toDraftInput(drafts) });

    if (!parsed.success) {
      const nextErrors: BuilderErrors = { byField: {} };
      for (const issue of parsed.error.issues) {
        const [head, index, prop] = issue.path;
        if (head === "title" || head === "description") {
          nextErrors[head] ??= issue.message;
        } else if (head === "fields" && typeof index === "number" && (prop === "label" || prop === "options")) {
          const key = drafts[index]?.key;
          if (!key) continue;
          nextErrors.byField[key] ??= {};
          nextErrors.byField[key][prop] ??= issue.message;
        } else if (head === "fields") {
          nextErrors.fields ??= issue.message;
        }
      }
      setErrors(nextErrors);
      showToast("入力内容を確認してください", "error");
      return;
    }

    setErrors(EMPTY_ERRORS);
    const onSuccess = () => {
      showToast(isEdit ? "フォームを更新しました" : "フォームを作成しました");
      router.push("/form-builder");
      router.refresh();
    };
    const onError = (mutationError: Error) => {
      showToast(mutationError.message, "error");
    };

    if (form) {
      updateMutation.mutate({ id: form.id, input: parsed.data }, { onSuccess, onError });
    } else {
      createMutation.mutate(parsed.data, { onSuccess, onError });
    }
  }

  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
      <section className="flex flex-col gap-4" aria-label="フォームの設定">
        <Card className="justify-start">
          <Input
            id={`${idPrefix}-title`}
            label="フォーム名"
            value={title}
            maxLength={100}
            error={errors.title}
            onChange={(event) => {
              setTitle(event.target.value);
              setErrors((current) => ({ ...current, title: undefined }));
            }}
          />
          <Textarea
            id={`${idPrefix}-description`}
            label="説明（任意）"
            rows={2}
            value={description}
            maxLength={500}
            error={errors.description}
            onChange={(event) => {
              setDescription(event.target.value);
              setErrors((current) => ({ ...current, description: undefined }));
            }}
          />
        </Card>

        {isEdit ? (
          <p className="rounded-lg bg-warning-bg px-4 py-3 text-xs text-warning">
            項目を削除すると、その項目の回答は回答データに表示されなくなります。入力形式や選択肢を変更しても、既存の回答は書き換わりません。
          </p>
        ) : null}

        <ol className="flex flex-col gap-4">
          {drafts.map((draft, index) => {
            const fieldErrors = errors.byField[draft.key] ?? {};
            const fieldId = `${idPrefix}-${draft.key}`;
            return (
              <li key={draft.key}>
                <Card className="justify-start gap-3">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-sm font-semibold text-muted-foreground">項目 {index + 1}</span>
                    <div className="flex gap-1">
                      <Button
                        variant="secondary"
                        className="w-auto px-2 py-1.5"
                        aria-label="上へ移動"
                        disabled={index === 0}
                        onClick={() => moveDraft(index, -1)}
                      >
                        <ArrowUpIcon className="size-4" />
                      </Button>
                      <Button
                        variant="secondary"
                        className="w-auto px-2 py-1.5"
                        aria-label="下へ移動"
                        disabled={index === drafts.length - 1}
                        onClick={() => moveDraft(index, 1)}
                      >
                        <ArrowDownIcon className="size-4" />
                      </Button>
                      <Button
                        variant="secondary"
                        className="w-auto px-2 py-1.5 text-danger"
                        aria-label="項目を削除"
                        disabled={drafts.length === 1}
                        onClick={() => removeDraft(draft.key)}
                      >
                        <TrashIcon className="size-4" />
                      </Button>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-[minmax(0,1fr)_12rem]">
                    <Input
                      id={`${fieldId}-label`}
                      label="項目名"
                      value={draft.label}
                      maxLength={100}
                      error={fieldErrors.label}
                      onChange={(event) => updateDraft(draft.key, { label: event.target.value })}
                    />
                    <Select
                      id={`${fieldId}-type`}
                      label="入力形式"
                      value={draft.type}
                      onChange={(event) => updateDraft(draft.key, { type: event.target.value as FieldType })}
                    >
                      {FIELD_TYPES.map((type) => (
                        <option key={type} value={type}>
                          {FIELD_TYPE_LABEL[type]}
                        </option>
                      ))}
                    </Select>
                  </div>

                  {hasOptions(draft.type) ? (
                    <Textarea
                      id={`${fieldId}-options`}
                      label="選択肢（1行に1つ）"
                      rows={3}
                      value={draft.optionsText}
                      error={fieldErrors.options}
                      onChange={(event) => updateDraft(draft.key, { optionsText: event.target.value })}
                    />
                  ) : null}

                  <Checkbox
                    label="必須項目にする"
                    checked={draft.required}
                    onChange={(event) => updateDraft(draft.key, { required: event.target.checked })}
                  />
                </Card>
              </li>
            );
          })}
        </ol>

        {errors.fields ? <p className="text-sm text-danger">{errors.fields}</p> : null}

        <Button variant="secondary" onClick={addDraft}>
          <span className="flex items-center justify-center gap-1.5">
            <PlusIcon className="size-4" />
            項目を追加
          </span>
        </Button>

        <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <Button variant="secondary" className="sm:w-auto" onClick={() => router.back()}>
            キャンセル
          </Button>
          <Button className="sm:w-auto" disabled={isPending} onClick={handleSubmit}>
            {isPending ? "保存中..." : isEdit ? "変更を保存" : "フォームを作成"}
          </Button>
        </div>
      </section>

      <section className="flex flex-col gap-3 lg:sticky lg:top-6 lg:self-start" aria-label="プレビュー">
        <h3 className="text-sm font-semibold text-muted-foreground">プレビュー</h3>
        <Card className="justify-start gap-5">
          <div className="flex flex-col gap-1">
            <p className="text-lg font-bold break-words">{title.trim() || "（フォーム名未設定）"}</p>
            {description.trim() ? <p className="text-sm text-muted-foreground">{description}</p> : null}
          </div>
          {drafts.map((draft) => {
            const options = parseOptions(draft.optionsText);
            return (
              <FieldInput
                key={draft.key}
                id={`${idPrefix}-${draft.key}-preview`}
                field={{ label: draft.label, type: draft.type, required: draft.required, options }}
                value={previewValues[draft.key] ?? (draft.type === "checkbox" ? [] : "")}
                onChange={(value) => setPreviewValues((current) => ({ ...current, [draft.key]: value }))}
              />
            );
          })}
        </Card>
      </section>
    </div>
  );
}
