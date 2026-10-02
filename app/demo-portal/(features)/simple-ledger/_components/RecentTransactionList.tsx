import { Card, Table, TableCell, TableHeaderCell, TableRow } from "@/components/ui";
import type { Transaction } from "../_lib/types";

function firstEntry(transaction: Transaction, type: "debit" | "credit") {
  return transaction.journalEntries?.find((entry) => entry.type === type);
}

export function RecentTransactionList({ transactions }: { transactions: Transaction[] }) {
  const recent = transactions.slice(0, 5);

  return (
    <Card className="p-0">
      <h3 className="p-6 pb-0 text-sm font-semibold">最近の取引</h3>
      <Table>
        <thead>
          <tr>
            <TableHeaderCell>日付</TableHeaderCell>
            <TableHeaderCell>借方</TableHeaderCell>
            <TableHeaderCell>貸方</TableHeaderCell>
            <TableHeaderCell>説明</TableHeaderCell>
          </tr>
        </thead>
        <tbody>
          {recent.length === 0 ? (
            <TableRow>
              <TableCell colSpan={4} className="text-center text-muted-foreground">
                取引がありません
              </TableCell>
            </TableRow>
          ) : (
            recent.map((transaction) => {
              const debit = firstEntry(transaction, "debit");
              const credit = firstEntry(transaction, "credit");
              return (
                <TableRow key={transaction.id}>
                  <TableCell className="whitespace-nowrap text-sm text-muted-foreground">
                    {transaction.date}
                  </TableCell>
                  <TableCell className="text-sm">
                    {debit ? `${debit.chartOfAccounts?.name} ${debit.amount.toLocaleString()}円` : "-"}
                  </TableCell>
                  <TableCell className="text-sm">
                    {credit ? `${credit.chartOfAccounts?.name} ${credit.amount.toLocaleString()}円` : "-"}
                  </TableCell>
                  <TableCell className="text-sm">{transaction.description}</TableCell>
                </TableRow>
              );
            })
          )}
        </tbody>
      </Table>
    </Card>
  );
}
