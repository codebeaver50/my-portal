"use client";

import { useRef, useState } from "react";
import { mergeProps, useFocusRing, useMove } from "react-aria";
import { cn } from "@/components/ui";
import { GRID_COLUMNS, MIN_FIELD_WIDTH } from "../../_lib/layout";

// キャンバスの行の列間の余白（BuilderCanvas の gap-x-4）。ドラッグ量を列数に換算するのに使う。
const COLUMN_GAP_PX = 16;

type WidthHandleProps = {
  // スクリーンリーダーで読み上げる項目名（例:「氏名」）。
  fieldName: string;
  width: number;
  maxWidth: number;
  onResize: (width: number) => void;
};

// WidthHandle は項目の右端に置く幅変更のつまみ。マウス・タッチでは左右にドラッグし、
// キーボードではフォーカスして ←/→ で1単位（1/12）ずつ変える（React Aria の useMove）。
export function WidthHandle({ fieldName, width, maxWidth, onResize }: WidthHandleProps) {
  const ref = useRef<HTMLDivElement>(null);
  const start = useRef({ width, deltaX: 0, columnPx: 1 });
  const [isResizing, setIsResizing] = useState(false);
  const { focusProps, isFocusVisible } = useFocusRing();

  const { moveProps } = useMove({
    onMoveStart() {
      const row = ref.current?.closest("[data-layout-row]");
      const rowWidth = row?.getBoundingClientRect().width ?? GRID_COLUMNS;
      start.current = { width, deltaX: 0, columnPx: (rowWidth + COLUMN_GAP_PX) / GRID_COLUMNS };
      setIsResizing(true);
    },
    onMove(event) {
      if (event.pointerType === "keyboard") {
        if (event.deltaX !== 0) onResize(width + Math.sign(event.deltaX));
        return;
      }
      start.current.deltaX += event.deltaX;
      onResize(start.current.width + Math.round(start.current.deltaX / start.current.columnPx));
    },
    onMoveEnd() {
      setIsResizing(false);
    },
  });

  const percent = Math.round((width / GRID_COLUMNS) * 100);

  return (
    <div
      {...mergeProps(moveProps, focusProps)}
      ref={ref}
      role="separator"
      tabIndex={0}
      aria-orientation="vertical"
      aria-label={`${fieldName}の幅`}
      aria-valuenow={width}
      aria-valuemin={MIN_FIELD_WIDTH}
      aria-valuemax={maxWidth}
      aria-valuetext={`${width}/${GRID_COLUMNS}（${percent}%）`}
      className="group absolute inset-y-0 -right-3 z-[5] flex w-4 cursor-col-resize touch-none justify-center outline-none"
    >
      <div
        className={cn(
          "my-auto h-10 w-1 rounded-full transition-colors",
          isResizing || isFocusVisible ? "bg-accent" : "bg-border group-hover:bg-accent",
        )}
      />
      {isResizing || isFocusVisible ? (
        <span className="pointer-events-none absolute -top-7 rounded bg-foreground px-1.5 py-0.5 text-xs whitespace-nowrap text-background">
          幅 {width}/{GRID_COLUMNS}
        </span>
      ) : null}
    </div>
  );
}
