"use client";

import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createTransaction,
  deleteTransaction,
  fetchChartOfAccounts,
  fetchTransactions,
  fetchTransactionsPaginated,
  updateTransaction,
} from "./api";
import type { TransactionInput } from "./types";

const TRANSACTIONS_KEY = ["simple-ledger", "transactions"];
const CHART_OF_ACCOUNTS_KEY = ["simple-ledger", "chart-of-accounts"];
const PAGE_SIZE = 30;

export function useChartOfAccounts() {
  return useQuery({
    queryKey: CHART_OF_ACCOUNTS_KEY,
    queryFn: () => fetchChartOfAccounts(),
    staleTime: 5 * 60 * 1000,
  });
}

export function useAllTransactions() {
  return useQuery({
    queryKey: TRANSACTIONS_KEY,
    queryFn: () => fetchTransactions(),
  });
}

export function useInfiniteTransactions(keyword?: string) {
  return useInfiniteQuery({
    queryKey: [...TRANSACTIONS_KEY, "infinite", keyword ?? ""],
    queryFn: ({ pageParam }) => fetchTransactionsPaginated(pageParam, PAGE_SIZE, keyword),
    initialPageParam: 1,
    getNextPageParam: (lastPage) => (lastPage.hasNextPage ? lastPage.page + 1 : undefined),
  });
}

export function useCreateTransaction() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: TransactionInput) => createTransaction(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: TRANSACTIONS_KEY });
    },
  });
}

export function useUpdateTransaction() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: TransactionInput }) => updateTransaction(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: TRANSACTIONS_KEY });
    },
  });
}

export function useDeleteTransaction() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteTransaction(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: TRANSACTIONS_KEY });
    },
  });
}
