import type { ChartOfAccount, Transaction, TransactionInput } from "./types";

// Next.js and the Go API run in separate Docker containers, so both
// server- and client-side calls go through this same public base URL
// (Caddy reverse-proxies /api/* on that host to the api container).
export function getApiBaseUrl(): string {
  return process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
}

async function parseErrorMessage(response: Response, fallback: string): Promise<string> {
  const body = await response.json().catch(() => null);
  return (body && typeof body.error === "string" && body.error) || fallback;
}

export async function fetchChartOfAccounts(init?: RequestInit): Promise<ChartOfAccount[]> {
  const response = await fetch(`${getApiBaseUrl()}/api/simple-ledger/chart-of-accounts`, init);
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "勘定科目の取得に失敗しました"));
  }
  const data: { accounts: ChartOfAccount[] } = await response.json();
  return data.accounts;
}

export async function fetchTransactions(init?: RequestInit): Promise<Transaction[]> {
  const response = await fetch(`${getApiBaseUrl()}/api/simple-ledger/transactions`, init);
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "取引の取得に失敗しました"));
  }
  const data: { transactions: Transaction[] } = await response.json();
  return data.transactions;
}

export type TransactionsPage = {
  transactions: Transaction[];
  total: number;
  page: number;
  pageSize: number;
  hasNextPage: boolean;
};

export async function fetchTransactionsPaginated(
  page: number,
  pageSize: number,
  keyword?: string,
  init?: RequestInit,
): Promise<TransactionsPage> {
  const params = new URLSearchParams({ page: String(page), pageSize: String(pageSize) });
  if (keyword) params.set("keyword", keyword);

  const response = await fetch(
    `${getApiBaseUrl()}/api/simple-ledger/transactions/paginated?${params.toString()}`,
    init,
  );
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "取引の取得に失敗しました"));
  }
  return response.json();
}

export async function createTransaction(input: TransactionInput): Promise<Transaction> {
  const response = await fetch(`${getApiBaseUrl()}/api/simple-ledger/transactions`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "取引の作成に失敗しました"));
  }
  return response.json();
}

export async function updateTransaction(id: number, input: TransactionInput): Promise<Transaction> {
  const response = await fetch(`${getApiBaseUrl()}/api/simple-ledger/transactions/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "取引の更新に失敗しました"));
  }
  return response.json();
}

export async function deleteTransaction(id: number): Promise<void> {
  const response = await fetch(`${getApiBaseUrl()}/api/simple-ledger/transactions/${id}`, {
    method: "DELETE",
  });
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "取引の削除に失敗しました"));
  }
}
