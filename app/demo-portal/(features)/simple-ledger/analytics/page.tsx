import type { Metadata } from "next";
import { AnalyticsView } from "../_components/AnalyticsView";

export const metadata: Metadata = {
  title: "分析 | 複式簿記シンプル家計簿",
  description: "カテゴリ別の収支分析ページ",
};

export default function SimpleLedgerAnalyticsPage() {
  return <AnalyticsView />;
}
