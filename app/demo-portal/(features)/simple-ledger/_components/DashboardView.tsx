"use client";

import { useMemo, useState } from "react";
import { Button } from "@/components/ui";
import { useAllTransactions } from "../_lib/useTransactions";
import { calculateBalance, calculateMonthlyBalance } from "../_lib/balance";
import { SummaryCard } from "./SummaryCard";
import { BalanceTrendChart } from "./BalanceTrendChart";
import { RecentTransactionList } from "./RecentTransactionList";
import { TransactionFormModal } from "./TransactionFormModal";
import { PlusIcon } from "./icons";

export function DashboardView() {
  const { data: transactions = [], isLoading } = useAllTransactions();
  const [isAddOpen, setIsAddOpen] = useState(false);

  const { totalIncome, totalExpense, balance } = useMemo(() => calculateBalance(transactions), [transactions]);
  const monthly = useMemo(() => calculateMonthlyBalance(transactions), [transactions]);

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">ダッシュボード</h2>
        <Button className="w-auto" onClick={() => setIsAddOpen(true)}>
          <span className="flex items-center gap-1.5">
            <PlusIcon className="size-4" />
            取引を追加
          </span>
        </Button>
      </div>

      {isLoading ? (
        <p className="text-sm text-muted-foreground">読み込み中...</p>
      ) : (
        <>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <SummaryCard variant="income" value={totalIncome} />
            <SummaryCard variant="expense" value={totalExpense} />
            <SummaryCard variant="balance" value={balance} />
          </div>
          <BalanceTrendChart data={monthly} />
          <RecentTransactionList transactions={transactions} />
        </>
      )}

      <TransactionFormModal open={isAddOpen} onClose={() => setIsAddOpen(false)} />
    </div>
  );
}
