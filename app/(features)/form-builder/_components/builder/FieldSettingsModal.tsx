"use client";

import { useId, useState } from "react";
import { Button, Checkbox, Input, Modal, Select, Textarea } from "@/components/ui";
import { FIELD_TYPE_LABEL, hasOptions } from "../../_lib/fieldTypes";
import { DEFAULT_OPTIONS_TEXT, toFieldInput } from "../../_lib/draft";
import { formFieldSchema } from "../../_lib/schema";
import { FIELD_TYPES } from "../../_lib/types";
import { ArrowDownIcon, ArrowUpIcon } from "../icons";
import type { FieldDraft } from "../../_lib/draft";
import type { FieldType } from "../../_lib/types";
import type { FieldErrors } from "./CanvasField";

type FieldSettingsModalProps = {
  draft: FieldDraft;
  initialErrors?: FieldErrors;
  // 現在の位置（例:「2行目・1列目」）。並び順ボタンで即時に変わる。
  positionText: string;
  canMoveBackward: boolean;
  canMoveForward: boolean;
  onMove: (direction: -1 | 1) => void;
  onSave: (draft: FieldDraft) => void;
  onClose: () => void;
};

// FieldSettingsModal は項目の設定ダイアログ。項目名・入力形式・選択肢・必須は「保存」で反映し、
// 並び順のボタンはドラッグ操作の代わりとして即時に反映する。
export function FieldSettingsModal({
  draft,
  initialErrors,
  positionText,
  canMoveBackward,
  canMoveForward,
  onMove,
  onSave,
  onClose,
}: FieldSettingsModalProps) {
  const idPrefix = useId();
  const [current, setCurrent] = useState(draft);
  const [errors, setErrors] = useState<FieldErrors>(initialErrors ?? {});

  function update(patch: Partial<FieldDraft>) {
    setCurrent((value) => ({ ...value, ...patch }));
  }

  function handleTypeChange(type: FieldType) {
    // 選択肢を持つ形式に切り替えたとき、選択肢が空なら初期値を入れておく。
    const optionsText = hasOptions(type) && current.optionsText.trim() === "" ? DEFAULT_OPTIONS_TEXT : current.optionsText;
    update({ type, optionsText });
    setErrors((value) => ({ ...value, options: undefined }));
  }

  function handleSave() {
    const parsed = formFieldSchema.safeParse(toFieldInput(current, 0));
    if (!parsed.success) {
      const next: FieldErrors = {};
      for (const issue of parsed.error.issues) {
        const [prop] = issue.path;
        if (prop === "label" || prop === "options") next[prop] ??= issue.message;
      }
      setErrors(next);
      return;
    }
    // 保存する値は入力どおり（前後の空白は送信時のバリデーションで除く）。
    onSave(current);
  }

  return (
    <Modal
      isOpen
      onClose={onClose}
      title="項目の設定"
      footer={
        <>
          <Button variant="secondary" className="w-auto" onClick={onClose}>
            キャンセル
          </Button>
          <Button className="w-auto" onClick={handleSave}>
            保存
          </Button>
        </>
      }
    >
      <Input
        id={`${idPrefix}-label`}
        label="項目名"
        value={current.label}
        maxLength={100}
        autoFocus
        error={errors.label}
        onChange={(event) => {
          update({ label: event.target.value });
          setErrors((value) => ({ ...value, label: undefined }));
        }}
      />
      <Select
        id={`${idPrefix}-type`}
        label="入力形式"
        value={current.type}
        onChange={(event) => handleTypeChange(event.target.value as FieldType)}
      >
        {FIELD_TYPES.map((type) => (
          <option key={type} value={type}>
            {FIELD_TYPE_LABEL[type]}
          </option>
        ))}
      </Select>
      {hasOptions(current.type) ? (
        <Textarea
          id={`${idPrefix}-options`}
          label="選択肢（1行に1つ）"
          rows={4}
          value={current.optionsText}
          error={errors.options}
          onChange={(event) => {
            update({ optionsText: event.target.value });
            setErrors((value) => ({ ...value, options: undefined }));
          }}
        />
      ) : null}
      <Checkbox
        label="必須項目にする"
        checked={current.required}
        onChange={(event) => update({ required: event.target.checked })}
      />

      <div className="flex flex-col gap-2 border-t border-border pt-4">
        <p className="text-sm font-medium">並び順</p>
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-sm text-muted-foreground" aria-live="polite">
            現在の位置: {positionText}
          </span>
          <div className="ml-auto flex gap-2">
            <Button variant="secondary" className="w-auto px-3 py-1.5" disabled={!canMoveBackward} onClick={() => onMove(-1)}>
              <span className="flex items-center gap-1">
                <ArrowUpIcon className="size-4" />
                前へ
              </span>
            </Button>
            <Button variant="secondary" className="w-auto px-3 py-1.5" disabled={!canMoveForward} onClick={() => onMove(1)}>
              <span className="flex items-center gap-1">
                <ArrowDownIcon className="size-4" />
                後ろへ
              </span>
            </Button>
          </div>
        </div>
        <p className="text-xs text-muted-foreground">
          キャンバス上の項目をドラッグしても並べ替えられます。「前へ」「後ろへ」は横並びの行にも出入りします。
        </p>
      </div>
    </Modal>
  );
}
