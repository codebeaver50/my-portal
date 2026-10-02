// フォームのレイアウト（行ごとに横並びの項目）を扱う純粋関数。
// API上は項目を表示順の平らな配列＋行番号で持ち、構築画面では行ごとの配列で扱う。

// Go側（formbuilder/service の maxFieldsPerRow / gridColumns / minFieldWidth）と揃える。
export const MAX_FIELDS_PER_ROW = 3;
// 項目の幅は1行を GRID_COLUMNS 等分した単位で表す。行の合計は GRID_COLUMNS 以内で、右側に余白が残ってもよい。
export const GRID_COLUMNS = 12;
export const MIN_FIELD_WIDTH = 3;

// 項目を置ける位置。newRow は rowIndex 番目の行の前に新しい行として挿入（末尾なら行数と同じ値）、
// inRow は rowIndex 番目の行の columnIndex 番目の前に横並びで挿入（末尾なら行の項目数と同じ値）。
export type DropPosition =
  | { kind: "newRow"; rowIndex: number }
  | { kind: "inRow"; rowIndex: number; columnIndex: number };

type Keyed = { key: string };
type Sized = Keyed & { width: number };

function totalWidth(items: readonly { width: number }[]): number {
  return items.reduce((sum, item) => sum + item.width, 0);
}

// fitRow は key の項目が行に収まるよう幅を調整する。他の項目の幅はなるべく保ち、
// 残りの幅が足りなければ key の項目を残りの幅に縮め、それも最小幅未満なら行を均等割りにする。
function fitRow<T extends Sized>(row: T[], key: string): T[] {
  const othersWidth = totalWidth(row.filter((item) => item.key !== key));
  const item = row.find((candidate) => candidate.key === key);
  if (!item || othersWidth + item.width <= GRID_COLUMNS) return row;

  const remaining = GRID_COLUMNS - othersWidth;
  if (remaining >= MIN_FIELD_WIDTH) {
    return row.map((candidate) => (candidate.key === key ? { ...candidate, width: remaining } : candidate));
  }
  const equal = Math.floor(GRID_COLUMNS / row.length);
  return row.map((candidate) => ({ ...candidate, width: equal }));
}

function fitRowContaining<T extends Sized>(rows: T[][], key: string): T[][] {
  return rows.map((row) => (row.some((item) => item.key === key) ? fitRow(row, key) : row));
}

// toRows は行番号つきの平らな配列を行ごとの配列にまとめる。配列は表示順に並んでいる前提。
export function toRows<T extends { row: number }>(fields: readonly T[]): T[][] {
  const rows: T[][] = [];
  fields.forEach((field, index) => {
    if (index > 0 && field.row === fields[index - 1].row) {
      rows[rows.length - 1].push(field);
    } else {
      rows.push([field]);
    }
  });
  return rows;
}

function insertAt<T>(rows: (T | null)[][], item: T, position: DropPosition): (T | null)[][] {
  if (position.kind === "newRow") {
    return [...rows.slice(0, position.rowIndex), [item], ...rows.slice(position.rowIndex)];
  }
  return rows.map((row, rowIndex) =>
    rowIndex === position.rowIndex
      ? [...row.slice(0, position.columnIndex), item, ...row.slice(position.columnIndex)]
      : row,
  );
}

function compact<T>(rows: (T | null)[][]): T[][] {
  return rows.map((row) => row.filter((item): item is T => item !== null)).filter((row) => row.length > 0);
}

// insertItem は item を position に挿入する。既存の行に入れて幅が足りなければ fitRow で調整する。
export function insertItem<T extends Sized>(rows: T[][], item: T, position: DropPosition): T[][] {
  return fitRowContaining(compact(insertAt(rows, item, position)), item.key);
}

// moveItem は key の項目を position へ移動する。position は移動前のレイアウト上の位置で、
// 移動元が空いた結果として空になった行は取り除く。移動した項目の幅は保ち、移動先の行に収まらなければ
// fitRow で調整する。
export function moveItem<T extends Sized>(rows: T[][], key: string, position: DropPosition): T[][] {
  const item = rows.flat().find((candidate) => candidate.key === key);
  if (!item) return rows;
  // 移動元を null にして位置（インデックス）を保ったまま挿入し、最後に詰める。
  const vacated = rows.map((row) => row.map((candidate) => (candidate.key === key ? null : candidate)));
  return fitRowContaining(compact(insertAt(vacated, item, position)), key);
}

export function removeItem<T extends Keyed>(rows: T[][], key: string): T[][] {
  return rows.map((row) => row.filter((item) => item.key !== key)).filter((row) => row.length > 0);
}

export function layoutSignature(rows: readonly Keyed[][]): string {
  return rows.map((row) => row.map((item) => item.key).join(",")).join("|");
}

// canDrop は position に置けるかどうかを返す。draggedKey は配置済み項目の移動時に指定し、
// 満員の行でも同じ行の中での並べ替えは許可する。移動しても配置が変わらない位置は置けない扱いにする。
export function canDrop<T extends Sized>(rows: T[][], position: DropPosition, draggedKey?: string): boolean {
  if (position.kind === "inRow") {
    const row = rows[position.rowIndex] ?? [];
    const others = row.filter((item) => item.key !== draggedKey);
    if (others.length >= MAX_FIELDS_PER_ROW) return false;
  }
  if (draggedKey === undefined) return true;
  return layoutSignature(moveItem(rows, draggedKey, position)) !== layoutSignature(rows);
}

// allPositions は置ける位置を読み順（上の行から、行の中は左から）に列挙する。
export function allPositions(rows: readonly unknown[][]): DropPosition[] {
  const positions: DropPosition[] = [];
  for (let rowIndex = 0; rowIndex <= rows.length; rowIndex++) {
    positions.push({ kind: "newRow", rowIndex });
    if (rowIndex === rows.length) break;
    for (let columnIndex = 0; columnIndex <= rows[rowIndex].length; columnIndex++) {
      positions.push({ kind: "inRow", rowIndex, columnIndex });
    }
  }
  return positions;
}

// stepPosition はドラッグ操作の代わりに、項目を読み順で1つ前（-1）／後ろ（1）の位置へ動かす
// ときの移動先を返す。移動先がなければ null。
export function stepPosition<T extends Sized>(rows: T[][], key: string, direction: -1 | 1): DropPosition | null {
  const positions = allPositions(rows);
  const current = layoutSignature(rows);
  const results = positions.map((position) => {
    if (!canDrop(rows, position, key)) {
      // 配置が変わらない位置（＝現在地）か、満員で置けない位置かを区別する。
      return layoutSignature(moveItem(rows, key, position)) === current ? "current" : "blocked";
    }
    return "movable";
  });

  if (direction === -1) {
    const first = results.indexOf("current");
    for (let index = first - 1; index >= 0; index--) {
      if (results[index] === "movable") return positions[index];
    }
  } else {
    const last = results.lastIndexOf("current");
    for (let index = last + 1; index < positions.length; index++) {
      if (results[index] === "movable") return positions[index];
    }
  }
  return null;
}

// maxWidth は key の項目に設定できる最大の幅（行の残りの幅＋自分の幅）を返す。
export function maxWidth<T extends Sized>(rows: T[][], key: string): number {
  const row = rows.find((candidate) => candidate.some((item) => item.key === key)) ?? [];
  return GRID_COLUMNS - totalWidth(row.filter((item) => item.key !== key));
}

// resizeItem は key の項目の幅を、最小幅〜行に収まる最大幅の範囲で変更する。
export function resizeItem<T extends Sized>(rows: T[][], key: string, width: number): T[][] {
  const clamped = Math.min(Math.max(Math.round(width), MIN_FIELD_WIDTH), maxWidth(rows, key));
  return rows.map((row) => row.map((item) => (item.key === key && item.width !== clamped ? { ...item, width: clamped } : item)));
}
