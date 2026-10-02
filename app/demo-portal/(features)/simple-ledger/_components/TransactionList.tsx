"use client";

import { Badge, Button, Card, Table, TableCell, TableHeaderCell, TableRow, useToast } from "@/components/ui";
import { useDeleteTransaction } from "../_lib/useTransactions";
import { EditIcon, TrashIcon } from "./icons";
import type { Transaction } from "../_lib/types";

function firstEntry(transaction: Transaction, type: "debit" | "credit") {
  return transaction.journalEntries?.find((entry) => entry.type === type);
}

type TransactionListProps = {
  transactions: Transaction[];
  onEdit: (transaction: Transaction) => void;
};

export function TransactionList({ transactions, onEdit }: TransactionListProps) {
  const deleteMutation = useDeleteTransaction();
  const { showToast } = useToast();

  function handleDelete(transaction: Transaction) {
    if (!window.confirm(`「${transaction.description || "この取引"}」を削除しますか？`)) return;
    deleteMutation.mutate(transaction.id, {
      onSuccess: () => showToast("取引を削除しました"),
      onError: (error) => showToast(error instanceof Error ? error.message : "削除に失敗しました", "error"),
    });
  }

  return (
    <Card className="p-0">
      <Table>
        <thead>
          <tr>
            <TableHeaderCell>取引日</TableHeaderCell>
            <TableHeaderCell>借方</TableHeaderCell>
            <TableHeaderCell>貸方</TableHeaderCell>
            <TableHeaderCell>説明</TableHeaderCell>
            <TableHeaderCell>操作</TableHeaderCell>
          </tr>
        </thead>
        <tbody>
          {transactions.length === 0 ? (
            <TableRow>
              <TableCell colSpan={5} className="text-center text-muted-foreground">
                取引がありません
              </TableCell>
            </TableRow>
          ) : (
            transactions.map((transaction) => {
              const debit = firstEntry(transaction, "debit");
              const credit = firstEntry(transaction, "credit");
              return (
                <TableRow key={transaction.id}>
                  <TableCell className="whitespace-nowrap text-sm text-muted-foreground">
                    {transaction.date}
                  </TableCell>
                  <TableCell>
                    {debit ? (
                      <div className="flex flex-col gap-0.5">
                        <span className="text-sm font-medium">{debit.chartOfAccounts?.name}</span>
                        <span className="text-sm text-muted-foreground">{debit.amount.toLocaleString()}円</span>
                      </div>
                    ) : (
                      "-"
                    )}
                  </TableCell>
                  <TableCell>
                    {credit ? (
                      <div className="flex flex-col gap-0.5">
                        <span className="text-sm font-medium">{credit.chartOfAccounts?.name}</span>
                        <span className="text-sm text-muted-foreground">{credit.amount.toLocaleString()}円</span>
                      </div>
                    ) : (
                      "-"
                    )}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {transaction.isCorrection ? <Badge variant="warning">訂正</Badge> : null}
                      <span className="text-sm">{transaction.description}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      <Button
                        variant="secondary"
                        className="w-auto px-2.5 py-2"
                        aria-label="編集"
                        onClick={() => onEdit(transaction)}
                      >
                        <EditIcon className="size-4" />
                      </Button>
                      <Button
                        variant="secondary"
                        className="w-auto px-2.5 py-2"
                        aria-label="削除"
                        onClick={() => handleDelete(transaction)}
                      >
                        <TrashIcon className="size-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              );
            })
          )}
        </tbody>
      </Table>
    </Card>
  );
}
