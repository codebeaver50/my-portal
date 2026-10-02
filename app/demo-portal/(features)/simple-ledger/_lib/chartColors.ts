"use client";

import { useEffect, useState } from "react";

// dataviz skillの検証済みデフォルト配色（8色、CVD安全な固定順）。
// カテゴリ別グラフ（勘定科目名など、系列数が可変のもの）に使う。
export const CATEGORICAL_PALETTE_LIGHT = [
  "#2a78d6",
  "#eb6834",
  "#1baf7a",
  "#eda100",
  "#e87ba4",
  "#008300",
  "#4a3aa7",
  "#e34948",
];

export const CATEGORICAL_PALETTE_DARK = [
  "#3987e5",
  "#d95926",
  "#199e70",
  "#c98500",
  "#d55181",
  "#008300",
  "#9085e9",
  "#e66767",
];

// 収入/支出/収支の3系列は、app/globals.cssの --success/--danger/--accent と
// 同じ意味づけ（緑=収入・赤=支出・紫=収支）を保つため、そのライト/ダーク値を
// そのまま定数化している。緑と赤は色覚多様性の観点で隣接色として弱いため、
// チャート側では線種（実線・破線・点線）も併用して見分けられるようにする。
export const SEMANTIC_TREND_COLORS_LIGHT = { income: "#16a34a", expense: "#dc2626", balance: "#4f46e5" };
export const SEMANTIC_TREND_COLORS_DARK = { income: "#4ade80", expense: "#f87171", balance: "#818cf8" };

export function useTrendColors() {
  return useIsDarkMode() ? SEMANTIC_TREND_COLORS_DARK : SEMANTIC_TREND_COLORS_LIGHT;
}

function prefersDark(): boolean {
  return typeof window !== "undefined" && window.matchMedia("(prefers-color-scheme: dark)").matches;
}

export function useIsDarkMode(): boolean {
  const [isDark, setIsDark] = useState(prefersDark);

  useEffect(() => {
    const query = window.matchMedia("(prefers-color-scheme: dark)");
    const listener = (event: MediaQueryListEvent) => setIsDark(event.matches);
    query.addEventListener("change", listener);
    return () => query.removeEventListener("change", listener);
  }, []);

  return isDark;
}

export function useCategoricalPalette(): string[] {
  return useIsDarkMode() ? CATEGORICAL_PALETTE_DARK : CATEGORICAL_PALETTE_LIGHT;
}
