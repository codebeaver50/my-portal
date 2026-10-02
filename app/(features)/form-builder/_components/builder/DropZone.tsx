"use client";

import { useRef } from "react";
import { useDrop } from "react-aria";
import { cn } from "@/components/ui";
import { getDropOperation } from "./dnd";
import type { DropPosition } from "../../_lib/layout";

export type DropZoneVariant = "row" | "column-start" | "column-end" | "fill" | "end";

type DropZoneProps = {
  position: DropPosition;
  // スクリーンリーダーで読み上げる、落とし先の説明。
  label: string;
  // ドラッグ中、かつこの位置に置ける場合のみ true。
  isEnabled: boolean;
  // row: 行と行のあいだ（横線）、column-start / column-end: 項目の左／右（縦線）、
  // fill: 行の右側の余白全体（行の末尾に追加）、end: フォーム末尾の常設の枠。
  variant: DropZoneVariant;
  onDrop: (position: DropPosition) => void;
};

// DropZone はドラッグした項目・パーツの落とし先。マウス・タッチでは線の付近にドラッグし、
// キーボードでは Tab で落とし先を移動して Enter で落とす（React Aria の useDrop）。
// 置けない位置やドラッグしていない間は、読み上げ・クリックの対象から外す。
export function DropZone({ position, label, isEnabled, variant, onDrop }: DropZoneProps) {
  const ref = useRef<HTMLDivElement>(null);
  const { dropProps, isDropTarget } = useDrop({
    ref,
    isDisabled: !isEnabled,
    getDropOperation,
    onDrop: () => onDrop(position),
  });

  const zoneProps = {
    ...dropProps,
    ref,
    role: "button",
    tabIndex: isEnabled ? -1 : undefined,
    "aria-label": label,
    "aria-hidden": !isEnabled || undefined,
  };

  if (variant === "end") {
    return (
      <div
        {...zoneProps}
        className={cn(
          "flex min-h-20 items-center justify-center rounded-lg border-2 border-dashed px-4 text-center text-sm text-muted-foreground outline-none transition-colors",
          isDropTarget ? "border-accent bg-accent-bg text-accent" : isEnabled ? "border-accent/50" : "border-border",
        )}
      >
        パーツをここにドラッグして項目を追加
      </div>
    );
  }

  if (variant === "fill") {
    // 余白の左端（直前の項目との間の余白の中央）に挿入位置の縦線を出し、余白全体を薄く強調する。
    return (
      <div
        {...zoneProps}
        className={cn("absolute inset-y-0 -left-7 right-0 z-10 outline-none", !isEnabled && "pointer-events-none")}
      >
        <div
          className={cn(
            "absolute inset-y-0 left-7 right-0 rounded-lg border-2 border-dashed transition-colors",
            isDropTarget ? "border-accent bg-accent-bg" : isEnabled ? "border-accent/25" : "border-transparent",
          )}
        />
        <div
          className={cn(
            "absolute inset-y-0 left-[18px] w-1 rounded-full transition-colors",
            isDropTarget ? "bg-accent" : "bg-transparent",
          )}
        />
      </div>
    );
  }

  const isRow = variant === "row";
  return (
    <div
      {...zoneProps}
      className={cn(
        "absolute z-10 flex items-center justify-center outline-none",
        isRow ? "inset-x-0 -inset-y-3" : "inset-y-0 w-10",
        variant === "column-start" && "-left-7",
        variant === "column-end" && "-right-7",
        !isEnabled && "pointer-events-none",
      )}
    >
      <div
        className={cn(
          "rounded-full transition-colors",
          isRow ? "h-1 w-full" : "h-full w-1",
          isDropTarget ? "bg-accent" : isEnabled ? "bg-accent/25" : "bg-transparent",
        )}
      />
    </div>
  );
}
