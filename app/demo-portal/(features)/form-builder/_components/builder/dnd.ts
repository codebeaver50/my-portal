import type { DragTypes, DropOperation } from "react-aria";
import type { FieldType } from "../../_lib/types";

// ドラッグ中のデータの種類。ブラウザがドラッグデータの種類名を小文字にするため小文字で定義する。
export const PART_DRAG_TYPE = "application/x-form-builder-part";
export const FIELD_DRAG_TYPE = "application/x-form-builder-field";

// Dragging はドラッグ中の対象。パーツ（新規項目）か、配置済み項目の移動か。
export type Dragging = { kind: "part"; type: FieldType } | { kind: "field"; key: string };

// このフォーム構築画面のドラッグ以外（ファイルや他ページのテキスト等）は受け付けない。
export function getDropOperation(types: DragTypes, allowedOperations: DropOperation[]): DropOperation {
  if (!types.has(PART_DRAG_TYPE) && !types.has(FIELD_DRAG_TYPE)) return "cancel";
  return allowedOperations[0] ?? "cancel";
}
