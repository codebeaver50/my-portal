"use client";

import { useId, useState } from "react";
import { Button, Card, LinkButton, cn, useToast } from "@/components/ui";
import { ApiError } from "../_lib/api";
import { toRows } from "../_lib/layout";
import { SM_COL_SPAN_CLASS } from "./layoutClasses";
import { validateAnswer } from "../_lib/schema";
import { useCreateRecord } from "../_lib/useForms";
import { FieldInput } from "./FieldInput";
import type { AnswerDraft } from "../_lib/schema";
import type { Form } from "../_lib/types";

function emptyDraft(form: Form): AnswerDraft {
  return Object.fromEntries(form.fields.map((field) => [String(field.id), field.type === "checkbox" ? [] : ""]));
}

export function AnswerForm({ form }: { form: Form }) {
  const idPrefix = useId();
  const { showToast } = useToast();
  const createMutation = useCreateRecord(form.id);

  const [draft, setDraft] = useState<AnswerDraft>(() => emptyDraft(form));
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [isSubmitted, setIsSubmitted] = useState(false);

  function handleChange(key: string, value: string | string[]) {
    setDraft((current) => ({ ...current, [key]: value }));
    setFieldErrors((current) => {
      if (!(key in current)) return current;
      const next = { ...current };
      delete next[key];
      return next;
    });
  }

  function handleSubmit() {
    const result = validateAnswer(form.fields, draft);
    if (!result.success) {
      setFieldErrors(result.fieldErrors);
      showToast("入力内容を確認してください", "error");
      return;
    }

    setFieldErrors({});
    createMutation.mutate(
      { values: result.values },
      {
        onSuccess: () => {
          setIsSubmitted(true);
          setDraft(emptyDraft(form));
        },
        onError: (mutationError) => {
          if (mutationError instanceof ApiError) {
            setFieldErrors(mutationError.fieldErrors);
          }
          showToast(mutationError.message, "error");
        },
      },
    );
  }

  if (isSubmitted) {
    return (
      <Card className="items-center gap-5 py-10 text-center">
        <div className="flex flex-col gap-1">
          <p className="text-lg font-bold">回答を送信しました</p>
          <p className="text-sm text-muted-foreground">ご回答ありがとうございました。</p>
        </div>
        <div className="flex w-full flex-col gap-3 sm:w-auto sm:flex-row">
          <Button variant="secondary" className="sm:w-auto" onClick={() => setIsSubmitted(false)}>
            もう一度回答する
          </Button>
          <LinkButton href={`/demo-portal/form-builder/${form.id}/records`} className="sm:w-auto">
            回答データを見る
          </LinkButton>
        </div>
      </Card>
    );
  }

  return (
    <Card className="mx-auto w-full max-w-3xl justify-start gap-5">
      {toRows(form.fields).map((row) => (
        // スマホ幅では縦に積み、sm以上で12列のグリッドに幅どおり横並びにする。
        <div key={row[0].id} className="grid grid-cols-1 gap-5 sm:grid-cols-12">
          {row.map((field) => {
            const key = String(field.id);
            return (
              <div key={field.id} className={cn("min-w-0", SM_COL_SPAN_CLASS[field.width])}>
                <FieldInput
                  id={`${idPrefix}-${field.id}`}
                  field={field}
                  value={draft[key] ?? ""}
                  error={fieldErrors[key]}
                  onChange={(value) => handleChange(key, value)}
                />
              </div>
            );
          })}
        </div>
      ))}
      <p className="text-xs text-muted-foreground">
        <span className="text-danger">*</span> は必須項目です
      </p>
      <Button disabled={createMutation.isPending} onClick={handleSubmit}>
        {createMutation.isPending ? "送信中..." : "回答を送信"}
      </Button>
    </Card>
  );
}
