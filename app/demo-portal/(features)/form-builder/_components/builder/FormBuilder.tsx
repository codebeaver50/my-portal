"use client";

import { useCallback, useId, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { Button, Card, Input, Textarea, useToast } from "@/components/ui";
import { draftRowsFromForm, newDraft, toFieldInputs } from "../../_lib/draft";
import { insertItem, moveItem, removeItem, resizeItem, stepPosition } from "../../_lib/layout";
import { MAX_FIELDS, formSchema } from "../../_lib/schema";
import { useCreateForm, useUpdateForm } from "../../_lib/useForms";
import { BuilderCanvas } from "./BuilderCanvas";
import { FieldSettingsModal } from "./FieldSettingsModal";
import { PartsPalette } from "./PartsPalette";
import type { FieldDraft } from "../../_lib/draft";
import type { DropPosition } from "../../_lib/layout";
import type { FieldType, Form } from "../../_lib/types";
import type { FieldErrors } from "./CanvasField";
import type { Dragging } from "./dnd";

type BuilderErrors = {
  title?: string;
  description?: string;
  fields?: string;
  byField: Record<string, FieldErrors>;
};

const EMPTY_ERRORS: BuilderErrors = { byField: {} };

function withoutFieldErrors(errors: BuilderErrors, key: string): BuilderErrors {
  if (!errors.byField[key]) return errors;
  const byField = { ...errors.byField };
  delete byField[key];
  return { ...errors, byField };
}

type FormBuilderProps = {
  form?: Form;
};

// FormBuilder はフォームの構築画面。左のパーツをキャンバスへドラッグして項目を追加し、
// 配置済みの項目もドラッグで並べ替え・横並びにできる。
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
  const [rows, setRows] = useState<FieldDraft[][]>(() => (form ? draftRowsFromForm(form) : []));
  const [errors, setErrors] = useState<BuilderErrors>(EMPTY_ERRORS);
  const [dragging, setDragging] = useState<Dragging | null>(null);
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [focusKey, setFocusKey] = useState<string | null>(null);

  const isPending = createMutation.isPending || updateMutation.isPending;
  const fieldCount = rows.flat().length;
  const isFull = fieldCount >= MAX_FIELDS;

  const clearFocusKey = useCallback(() => setFocusKey(null), []);
  const handleDragEnd = useCallback(() => setDragging(null), []);

  function addField(type: FieldType, position: DropPosition) {
    const key = newKey();
    setRows((current) => insertItem(current, newDraft(key, type), position));
    setErrors((current) => ({ ...current, fields: undefined }));
    setFocusKey(key);
  }

  function handleDrop(position: DropPosition) {
    if (!dragging) return;
    // 移動した項目が別の行へ移ると要素が作り直され、元の要素に dragend が届かないことがあるため、
    // ドロップした時点でドラッグ状態を終える。
    setDragging(null);
    if (dragging.kind === "part") {
      addField(dragging.type, position);
    } else {
      setRows((current) => moveItem(current, dragging.key, position));
      setFocusKey(dragging.key);
    }
  }

  function handleRemove(key: string) {
    setRows((current) => removeItem(current, key));
    setErrors((current) => withoutFieldErrors(current, key));
  }

  function handleResize(key: string, width: number) {
    setRows((current) => resizeItem(current, key, width));
  }

  function handleSaveField(draft: FieldDraft) {
    setRows((current) => current.map((row) => row.map((item) => (item.key === draft.key ? draft : item))));
    setErrors((current) => withoutFieldErrors(current, draft.key));
    setEditingKey(null);
    setFocusKey(draft.key);
  }

  function handleMoveField(key: string, direction: -1 | 1) {
    setRows((current) => {
      const position = stepPosition(current, key, direction);
      return position ? moveItem(current, key, position) : current;
    });
  }

  function handleSubmit() {
    const drafts = rows.flat();
    const parsed = formSchema.safeParse({ title, description, fields: toFieldInputs(rows) });

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
      router.push("/demo-portal/form-builder");
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

  const editing = (() => {
    if (editingKey === null) return null;
    for (const [rowIndex, row] of rows.entries()) {
      const columnIndex = row.findIndex((draft) => draft.key === editingKey);
      if (columnIndex >= 0) return { draft: row[columnIndex], rowIndex, columnIndex };
    }
    return null;
  })();

  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-[15rem_minmax(0,1fr)]">
      <section className="flex min-w-0 flex-col gap-2 lg:sticky lg:top-6 lg:self-start" aria-labelledby={`${idPrefix}-parts`}>
        <h3 id={`${idPrefix}-parts`} className="text-sm font-semibold text-muted-foreground">
          パーツ
        </h3>
        <PartsPalette
          isDisabled={isFull}
          onDragStart={setDragging}
          onDragEnd={handleDragEnd}
          onAdd={(type) => addField(type, { kind: "newRow", rowIndex: rows.length })}
        />
        <p className="text-xs text-muted-foreground">
          {isFull
            ? `項目は${MAX_FIELDS}個までです`
            : "右のフォームへドラッグして配置します。キーボードでは Enter で掴み、Tab で移動先を選んで Enter で置けます。"}
        </p>
      </section>

      <section className="flex min-w-0 flex-col gap-4" aria-label="フォーム">
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

        <Card className="justify-start gap-0 p-3 sm:p-6">
          <BuilderCanvas
            rows={rows}
            dragging={dragging}
            errorsByKey={errors.byField}
            focusKey={focusKey}
            onFocused={clearFocusKey}
            onDragStart={setDragging}
            onDragEnd={handleDragEnd}
            onDrop={handleDrop}
            onEdit={setEditingKey}
            onRemove={handleRemove}
            onResize={handleResize}
          />
        </Card>

        {errors.fields ? <p className="text-sm text-danger">{errors.fields}</p> : null}

        <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <Button variant="secondary" className="sm:w-auto" onClick={() => router.back()}>
            キャンセル
          </Button>
          <Button className="sm:w-auto" disabled={isPending} onClick={handleSubmit}>
            {isPending ? "保存中..." : isEdit ? "変更を保存" : "フォームを作成"}
          </Button>
        </div>
      </section>

      {editing ? (
        <FieldSettingsModal
          key={editing.draft.key}
          draft={editing.draft}
          initialErrors={errors.byField[editing.draft.key]}
          positionText={`${editing.rowIndex + 1}行目・${editing.columnIndex + 1}列目（この行の項目数: ${rows[editing.rowIndex].length}）`}
          canMoveBackward={stepPosition(rows, editing.draft.key, -1) !== null}
          canMoveForward={stepPosition(rows, editing.draft.key, 1) !== null}
          onMove={(direction) => handleMoveField(editing.draft.key, direction)}
          onSave={handleSaveField}
          onClose={() => {
            setEditingKey(null);
            setFocusKey(editing.draft.key);
          }}
        />
      ) : null}
    </div>
  );
}
