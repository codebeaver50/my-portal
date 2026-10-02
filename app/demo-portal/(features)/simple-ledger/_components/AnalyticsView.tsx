"use client";

import { useMemo } from "react";
import { useAllTransactions } from "../_lib/useTransactions";
import { calculateBalance, calculateCategoryAmounts } from "../_lib/balance";
import { SummaryCard } from "./SummaryCard";
import { CategoryPieChart } from "./CategoryPieChart";
import { CategoryList } from "./CategoryList";

export function AnalyticsView() {
  const { data: transactions = [], isLoading } = useAllTransactions();

  const { totalIncome, totalExpense, balance } = useMemo(() => calculateBalance(transactions), [transactions]);
  const incomeCategories = useMemo(() => calculateCategoryAmounts(transactions, "income"), [transactions]);
  const expenseCategories = useMemo(() => calculateCategoryAmounts(transactions, "expense"), [transactions]);

  if (isLoading) {
    return <p className="text-sm text-muted-foreground">読み込み中...</p>;
  }

  return (
    <div className="flex flex-col gap-6">
      <h2 className="text-xl font-bold">分析</h2>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <SummaryCard variant="count" value={transactions.length} unit="件" />
        <SummaryCard variant="income" value={totalIncome} />
        <SummaryCard variant="expense" value={totalExpense} />
        <SummaryCard variant="balance" value={balance} />
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <CategoryPieChart title="カテゴリ別収入" data={incomeCategories} />
        <CategoryPieChart title="カテゴリ別支出" data={expenseCategories} />
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <CategoryList title="収入カテゴリ詳細" data={incomeCategories} />
        <CategoryList title="支出カテゴリ詳細" data={expenseCategories} />
      </div>
    </div>
  );
}
