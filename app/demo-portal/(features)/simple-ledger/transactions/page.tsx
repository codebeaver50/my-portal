import type { Metadata } from "next";
import { TransactionsView } from "../_components/TransactionsView";

export const metadata: Metadata = {
  title: "取引一覧 | 複式簿記シンプル家計簿",
  description: "取引の作成・編集・削除ができる一覧ページ",
};

export default function SimpleLedgerTransactionsPage() {
  return <TransactionsView />;
}
