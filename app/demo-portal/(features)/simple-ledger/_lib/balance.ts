import type { AccountType, Transaction } from "./types";

export type BalanceSummary = {
  totalIncome: number;
  totalExpense: number;
  balance: number;
};

// 収益は貸方、費用は借方で記録される（複式簿記のルール）ため、その組み合わせ
// だけを集計対象にする。資産間の振替（現金→普通預金 等）は収支に含めない。
export function calculateBalance(transactions: Transaction[]): BalanceSummary {
  let totalIncome = 0;
  let totalExpense = 0;

  for (const transaction of transactions) {
    for (const entry of transaction.journalEntries ?? []) {
      const accountType = entry.chartOfAccounts?.type;
      if (accountType === "revenue" && entry.type === "credit") {
        totalIncome += entry.amount;
      }
      if (accountType === "expense" && entry.type === "debit") {
        totalExpense += entry.amount;
      }
    }
  }

  return { totalIncome, totalExpense, balance: totalIncome - totalExpense };
}

export type MonthlyBalancePoint = {
  month: string;
  income: number;
  expense: number;
  balance: number;
};

export function calculateMonthlyBalance(transactions: Transaction[]): MonthlyBalancePoint[] {
  const byMonth = new Map<string, { income: number; expense: number }>();

  for (const transaction of transactions) {
    const month = transaction.date.slice(0, 7); // YYYY-MM
    const current = byMonth.get(month) ?? { income: 0, expense: 0 };

    for (const entry of transaction.journalEntries ?? []) {
      const accountType = entry.chartOfAccounts?.type;
      if (accountType === "revenue" && entry.type === "credit") {
        current.income += entry.amount;
      }
      if (accountType === "expense" && entry.type === "debit") {
        current.expense += entry.amount;
      }
    }

    byMonth.set(month, current);
  }

  return Array.from(byMonth.entries())
    .map(([month, { income, expense }]) => ({ month, income, expense, balance: income - expense }))
    .sort((a, b) => a.month.localeCompare(b.month));
}

export type CategoryAmount = {
  name: string;
  value: number;
};

// 種別（収入/支出）ごとに、実際に使われた勘定科目名で合計金額を集計する。
// calculateBalanceと同じ「収益は貸方、費用は借方」というルールに揃えている
// ため、資産間の振替（現金→普通預金 等）はどちらにも含まれない。
export function calculateCategoryAmounts(
  transactions: Transaction[],
  side: "income" | "expense",
): CategoryAmount[] {
  const targetAccountType: AccountType = side === "income" ? "revenue" : "expense";
  const targetEntryType = side === "income" ? "credit" : "debit";

  const totals = new Map<string, number>();

  for (const transaction of transactions) {
    for (const entry of transaction.journalEntries ?? []) {
      if (entry.chartOfAccounts?.type !== targetAccountType || entry.type !== targetEntryType) {
        continue;
      }
      const name = entry.chartOfAccounts.name;
      totals.set(name, (totals.get(name) ?? 0) + entry.amount);
    }
  }

  return Array.from(totals.entries())
    .map(([name, value]) => ({ name, value }))
    .sort((a, b) => b.value - a.value);
}
