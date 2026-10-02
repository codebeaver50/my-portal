"use client";

import { useEffect, useMemo, useState } from "react";
import { useInView } from "react-intersection-observer";
import { Button } from "@/components/ui";
import { useInfiniteTransactions } from "../_lib/useTransactions";
import { useDebounce } from "../_lib/useDebounce";
import { TransactionFilterBar } from "./TransactionFilterBar";
import { TransactionList } from "./TransactionList";
import { TransactionFormModal } from "./TransactionFormModal";
import { PlusIcon } from "./icons";
import type { Transaction, TransactionCategoryFilter } from "../_lib/types";

function matchesCategory(transaction: Transaction, category: TransactionCategoryFilter): boolean {
  if (category === "all") return true;
  return (transaction.journalEntries ?? []).some((entry) => {
    if (category === "income") {
      return entry.type === "credit" && entry.chartOfAccounts?.type === "revenue";
    }
    return entry.type === "debit" && entry.chartOfAccounts?.type === "expense";
  });
}

export function TransactionsView() {
  const [searchValue, setSearchValue] = useState("");
  const debouncedKeyword = useDebounce(searchValue, 500);
  const [categoryValue, setCategoryValue] = useState<TransactionCategoryFilter>("all");
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [editingTransaction, setEditingTransaction] = useState<Transaction | null>(null);

  const { data, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteTransactions(debouncedKeyword);
  const { ref: sentinelRef, inView } = useInView();

  useEffect(() => {
    if (inView && hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [inView, hasNextPage, isFetchingNextPage, fetchNextPage]);

  const transactions = useMemo(() => {
    const all = data?.pages.flatMap((page) => page.transactions) ?? [];
    return all.filter((transaction) => matchesCategory(transaction, categoryValue));
  }, [data, categoryValue]);

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">取引一覧</h2>
        <Button className="w-auto" onClick={() => setIsAddOpen(true)}>
          <span className="flex items-center gap-1.5">
            <PlusIcon className="size-4" />
            取引を追加
          </span>
        </Button>
      </div>

      <TransactionFilterBar
        searchValue={searchValue}
        onSearchChange={setSearchValue}
        categoryValue={categoryValue}
        onCategoryChange={setCategoryValue}
      />

      <TransactionList transactions={transactions} onEdit={setEditingTransaction} />

      <div ref={sentinelRef} className="flex justify-center py-4 text-sm text-muted-foreground">
        {isFetchingNextPage ? "読み込み中..." : !hasNextPage && transactions.length > 0 ? "すべて読み込みました" : null}
      </div>

      <TransactionFormModal open={isAddOpen} onClose={() => setIsAddOpen(false)} />
      {editingTransaction ? (
        <TransactionFormModal
          key={editingTransaction.id}
          open
          onClose={() => setEditingTransaction(null)}
          transaction={editingTransaction}
        />
      ) : null}
    </div>
  );
}
