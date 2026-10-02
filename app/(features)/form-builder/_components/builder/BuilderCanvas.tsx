"use client";

import { cn } from "@/components/ui";
import { GRID_COLUMNS, canDrop, maxWidth } from "../../_lib/layout";
import { COL_SPAN_CLASS } from "../layoutClasses";
import { CanvasField } from "./CanvasField";
import { DropZone } from "./DropZone";
import { WidthHandle } from "./WidthHandle";
import type { FieldDraft } from "../../_lib/draft";
import type { DropPosition } from "../../_lib/layout";
import type { FieldErrors } from "./CanvasField";
import type { Dragging } from "./dnd";
import type { DropZoneVariant } from "./DropZone";

function fieldName(draft: FieldDraft): string {
  return `「${draft.label.trim() || "項目名未設定"}」`;
}

// positionLabel は落とし先をスクリーンリーダーで読み上げる文言。
function positionLabel(rows: FieldDraft[][], position: DropPosition): string {
  if (position.kind === "newRow") {
    const { rowIndex } = position;
    if (rows.length === 0) return "フォームに追加";
    if (rowIndex === 0) return "1行目の上に新しい行として挿入";
    if (rowIndex === rows.length) return "最後の行の下に新しい行として挿入";
    return `${rowIndex}行目と${rowIndex + 1}行目のあいだに新しい行として挿入`;
  }
  const row = rows[position.rowIndex];
  const line = `${position.rowIndex + 1}行目`;
  return position.columnIndex === 0
    ? `${line}の${fieldName(row[0])}の左に挿入`
    : `${line}の${fieldName(row[position.columnIndex - 1])}の右に挿入`;
}

type BuilderCanvasProps = {
  rows: FieldDraft[][];
  dragging: Dragging | null;
  errorsByKey: Record<string, FieldErrors>;
  focusKey: string | null;
  onFocused: () => void;
  onDragStart: (dragging: Dragging) => void;
  onDragEnd: () => void;
  onDrop: (position: DropPosition) => void;
  onEdit: (key: string) => void;
  onRemove: (key: string) => void;
  onResize: (key: string, width: number) => void;
};

// BuilderCanvas は行ごとに項目を横並びで表示し、行と行のあいだ・行の中の項目の前後に落とし先を置く。
// 行は12列のグリッドで、各項目は幅（width）の分だけ列を占める。右端のつまみで幅を変えられる。
export function BuilderCanvas({
  rows,
  dragging,
  errorsByKey,
  focusKey,
  onFocused,
  onDragStart,
  onDragEnd,
  onDrop,
  onEdit,
  onRemove,
  onResize,
}: BuilderCanvasProps) {
  const draggedKey = dragging?.kind === "field" ? dragging.key : undefined;
  const zone = (position: DropPosition, variant: DropZoneVariant) => (
    <DropZone
      position={position}
      variant={variant}
      label={positionLabel(rows, position)}
      isEnabled={dragging !== null && canDrop(rows, position, draggedKey)}
      onDrop={onDrop}
    />
  );

  return (
    <div className="flex flex-col">
      {rows.map((row, rowIndex) => {
        // 行の右側の余白（列数）。余白があれば余白全体を「行の末尾に追加」の落とし先にする。
        const remaining = GRID_COLUMNS - row.reduce((sum, draft) => sum + draft.width, 0);
        const endPosition: DropPosition = { kind: "inRow", rowIndex, columnIndex: row.length };
        return (
          <div key={row[0].key} className="flex flex-col">
            <div className="relative h-4">{zone({ kind: "newRow", rowIndex }, "row")}</div>
            <div data-layout-row className="grid grid-cols-12 gap-x-4">
              {row.map((draft, columnIndex) => (
                <div key={draft.key} className={cn("relative min-w-0", COL_SPAN_CLASS[draft.width])}>
                  {zone({ kind: "inRow", rowIndex, columnIndex }, "column-start")}
                  <CanvasField
                    draft={draft}
                    errors={errorsByKey[draft.key]}
                    shouldFocus={focusKey === draft.key}
                    onFocused={onFocused}
                    onDragStart={onDragStart}
                    onDragEnd={onDragEnd}
                    onEdit={onEdit}
                    onRemove={onRemove}
                  />
                  <WidthHandle
                    fieldName={fieldName(draft)}
                    width={draft.width}
                    maxWidth={maxWidth(rows, draft.key)}
                    onResize={(width) => onResize(draft.key, width)}
                  />
                  {columnIndex === row.length - 1 && remaining === 0 ? zone(endPosition, "column-end") : null}
                </div>
              ))}
              {remaining > 0 ? (
                <div className={cn("relative min-h-12", COL_SPAN_CLASS[remaining])}>{zone(endPosition, "fill")}</div>
              ) : null}
            </div>
          </div>
        );
      })}
      <div className={rows.length > 0 ? "pt-4" : undefined}>{zone({ kind: "newRow", rowIndex: rows.length }, "end")}</div>
    </div>
  );
}
