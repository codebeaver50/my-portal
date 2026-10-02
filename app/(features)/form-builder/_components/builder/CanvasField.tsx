"use client";

import { useEffect, useId, useRef } from "react";
import type { ButtonHTMLAttributes } from "react";
import { useButton, useDrag } from "react-aria";
import { cn } from "@/components/ui";
import { FIELD_TYPE_LABEL } from "../../_lib/fieldTypes";
import { parseOptions } from "../../_lib/draft";
import { FieldInput } from "../FieldInput";
import { GripIcon, PencilIcon, TrashIcon } from "../icons";
import { FIELD_DRAG_TYPE } from "./dnd";
import type { FieldDraft } from "../../_lib/draft";
import type { Dragging } from "./dnd";

export type FieldErrors = { label?: string; options?: string };

type CanvasFieldProps = {
  draft: FieldDraft;
  errors?: FieldErrors;
  // ドロップ・並べ替えの直後に、移動ハンドルへフォーカスを戻すかどうか。
  shouldFocus: boolean;
  onFocused: () => void;
  onDragStart: (dragging: Dragging) => void;
  onDragEnd: () => void;
  onEdit: (key: string) => void;
  onRemove: (key: string) => void;
};

function IconButton({ className, ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      type="button"
      className={cn("rounded p-1.5 text-muted-foreground hover:bg-neutral-bg hover:text-foreground", className)}
      {...props}
    />
  );
}

// CanvasField はキャンバス上の配置済み項目。回答画面と同じ見た目の入力欄（操作不可）を表示し、
// カード全体のドラッグか、移動ハンドル（キーボードでは Enter）で並べ替える。
export function CanvasField({
  draft,
  errors,
  shouldFocus,
  onFocused,
  onDragStart,
  onDragEnd,
  onEdit,
  onRemove,
}: CanvasFieldProps) {
  const previewId = useId();
  const label = draft.label.trim() || "（項目名未設定）";
  const handleRef = useRef<HTMLButtonElement>(null);

  const { dragProps, dragButtonProps, isDragging } = useDrag({
    hasDragButton: true,
    getItems: () => [{ [FIELD_DRAG_TYPE]: draft.key, "text/plain": label }],
    getAllowedDropOperations: () => ["move"],
    onDragStart: () => onDragStart({ kind: "field", key: draft.key }),
    onDragEnd,
  });
  const { buttonProps } = useButton({ ...dragButtonProps, "aria-label": `「${label}」を移動` }, handleRef);

  useEffect(() => {
    if (!shouldFocus) return;
    handleRef.current?.focus();
    onFocused();
  }, [shouldFocus, onFocused]);

  const errorMessages = [errors?.label, errors?.options].filter(Boolean);

  return (
    <div
      {...dragProps}
      onClick={() => onEdit(draft.key)}
      className={cn(
        "flex h-full cursor-pointer flex-col gap-2 rounded-lg border bg-card p-2 transition-colors hover:border-accent sm:p-4",
        errorMessages.length > 0 ? "border-danger" : "border-border",
        isDragging && "opacity-40",
      )}
    >
      <div className="flex flex-wrap items-center gap-1" onClick={(event) => event.stopPropagation()}>
        <button
          {...buttonProps}
          ref={handleRef}
          className="cursor-grab rounded p-1 text-muted-foreground outline-none hover:bg-neutral-bg focus-visible:ring-2 focus-visible:ring-accent"
        >
          <GripIcon className="size-4" />
        </button>
        <span className="hidden min-w-0 flex-1 truncate text-xs text-muted-foreground sm:block">{FIELD_TYPE_LABEL[draft.type]}</span>
        <span className="flex-1 sm:hidden" />
        <IconButton aria-label={`「${label}」を設定`} onClick={() => onEdit(draft.key)}>
          <PencilIcon className="size-4" />
        </IconButton>
        <IconButton aria-label={`「${label}」を削除`} className="hover:text-danger" onClick={() => onRemove(draft.key)}>
          <TrashIcon className="size-4" />
        </IconButton>
      </div>

      {/* 構築画面では入力できない見本として表示する。 */}
      <div inert className="pointer-events-none">
        <FieldInput
          id={previewId}
          field={{ label: draft.label, type: draft.type, required: draft.required, options: parseOptions(draft.optionsText) }}
          value={draft.type === "checkbox" ? [] : ""}
          onChange={() => {}}
        />
      </div>

      {errorMessages.map((message) => (
        <p key={message} className="text-xs text-danger">
          {message}
        </p>
      ))}
    </div>
  );
}
