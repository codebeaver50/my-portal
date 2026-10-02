// 項目の幅（1行を12等分した単位）ごとのTailwindのクラス。クラス名を検出できるよう静的に列挙する。
// 幅は3〜12（_lib/layout.ts の MIN_FIELD_WIDTH〜GRID_COLUMNS）。

export const COL_SPAN_CLASS: Record<number, string> = {
  3: "col-span-3",
  4: "col-span-4",
  5: "col-span-5",
  6: "col-span-6",
  7: "col-span-7",
  8: "col-span-8",
  9: "col-span-9",
  10: "col-span-10",
  11: "col-span-11",
  12: "col-span-12",
};

// 回答画面ではスマホ幅で縦に積み、sm以上で幅どおりに横並びにする。
export const SM_COL_SPAN_CLASS: Record<number, string> = {
  3: "sm:col-span-3",
  4: "sm:col-span-4",
  5: "sm:col-span-5",
  6: "sm:col-span-6",
  7: "sm:col-span-7",
  8: "sm:col-span-8",
  9: "sm:col-span-9",
  10: "sm:col-span-10",
  11: "sm:col-span-11",
  12: "sm:col-span-12",
};
