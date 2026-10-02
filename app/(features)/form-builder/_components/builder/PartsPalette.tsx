"use client";

import { useDrag } from "react-aria";
import { cn } from "@/components/ui";
import { FIELD_TYPE_LABEL } from "../../_lib/fieldTypes";
import { FIELD_TYPES } from "../../_lib/types";
import { GripIcon, PlusIcon } from "../icons";
import { PART_DRAG_TYPE } from "./dnd";
import type { FieldType } from "../../_lib/types";
import type { Dragging } from "./dnd";

type PaletteItemProps = {
  type: FieldType;
  isDisabled: boolean;
  onDragStart: (dragging: Dragging) => void;
  onDragEnd: () => void;
  onAdd: (type: FieldType) => void;
};

function PaletteItem({ type, isDisabled, onDragStart, onDragEnd, onAdd }: PaletteItemProps) {
  const label = FIELD_TYPE_LABEL[type];
  const { dragProps, isDragging } = useDrag({
    isDisabled,
    getItems: () => [{ [PART_DRAG_TYPE]: type, "text/plain": label }],
    getAllowedDropOperations: () => ["copy"],
    onDragStart: () => onDragStart({ kind: "part", type }),
    onDragEnd,
  });

  return (
    <li className="flex shrink-0 items-stretch gap-1">
      {/* マウス・タッチではドラッグ、キーボードでは Enter でドラッグを開始する（React Aria の useDrag）。 */}
      <div
        {...dragProps}
        role="button"
        tabIndex={isDisabled ? -1 : 0}
        aria-label={`${label}のパーツ`}
        aria-disabled={isDisabled || undefined}
        className={cn(
          "flex flex-1 items-center gap-2 rounded-lg border border-border bg-card px-3 py-2 text-sm whitespace-nowrap outline-none focus-visible:border-accent",
          isDisabled ? "cursor-not-allowed opacity-60" : "cursor-grab hover:border-accent",
          isDragging && "opacity-40",
        )}
      >
        <GripIcon className="size-4 shrink-0 text-muted-foreground" />
        {label}
      </div>
      <button
        type="button"
        aria-label={`${label}を末尾に追加`}
        disabled={isDisabled}
        onClick={() => onAdd(type)}
        className="rounded-lg border border-border px-2 text-muted-foreground hover:bg-neutral-bg disabled:cursor-not-allowed disabled:opacity-60"
      >
        <PlusIcon className="size-4" />
      </button>
    </li>
  );
}

type PartsPaletteProps = Omit<PaletteItemProps, "type">;

// PartsPalette は入力形式ごとのパーツ一覧。キャンバスへドラッグするか、＋ボタンで末尾に追加する。
export function PartsPalette(props: PartsPaletteProps) {
  return (
    <ul className="flex gap-2 overflow-x-auto pb-1 lg:flex-col lg:overflow-visible lg:pb-0" aria-label="パーツ">
      {FIELD_TYPES.map((type) => (
        <PaletteItem key={type} type={type} {...props} />
      ))}
    </ul>
  );
}
