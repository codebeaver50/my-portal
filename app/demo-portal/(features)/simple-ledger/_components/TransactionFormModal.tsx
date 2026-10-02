"use client";

import { useState } from "react";
import { Button, Input, Modal, Select, Textarea, useToast } from "@/components/ui";
import { useChartOfAccounts, useCreateTransaction, useUpdateTransaction } from "../_lib/useTransactions";
import { transactionSchema } from "../_lib/schema";
import { TrashIcon } from "./icons";
import type { ChartOfAccount, EntryType, Transaction } from "../_lib/types";

type LineInput = {
  chartOfAccountsId: string;
  type: EntryType;
  amount: string;
  description: string;
};

const EMPTY_LINE = (type: EntryType): LineInput => ({ chartOfAccountsId: "", type, amount: "", description: "" });

function toLineInputs(transaction?: Transaction): LineInput[] {
  if (!transaction?.journalEntries?.length) {
    return [EMPTY_LINE("debit"), EMPTY_LINE("credit")];
  }
  return transaction.journalEntries.map((entry) => ({
    chartOfAccountsId: String(entry.chartOfAccountsId),
    type: entry.type,
    amount: String(entry.amount),
    description: entry.description,
  }));
}

const ACCOUNT_TYPE_LABEL: Record<ChartOfAccount["type"], string> = {
  asset: "資産",
  liability: "負債",
  equity: "純資産",
  revenue: "収益",
  expense: "費用",
};

function groupAccountsByType(accounts: ChartOfAccount[]) {
  const groups = new Map<ChartOfAccount["type"], ChartOfAccount[]>();
  for (const account of accounts) {
    const list = groups.get(account.type) ?? [];
    list.push(account);
    groups.set(account.type, list);
  }
  return Array.from(groups.entries());
}

type TransactionFormModalProps = {
  open: boolean;
  onClose: () => void;
  transaction?: Transaction;
};

export function TransactionFormModal({ open, onClose, transaction }: TransactionFormModalProps) {
  const isEdit = Boolean(transaction);
  const { data: accounts = [] } = useChartOfAccounts();
  const createMutation = useCreateTransaction();
  const updateMutation = useUpdateTransaction();
  const { showToast } = useToast();

  const [date, setDate] = useState(transaction?.date ?? new Date().toISOString().slice(0, 10));
  const [description, setDescription] = useState(transaction?.description ?? "");
  const [lines, setLines] = useState<LineInput[]>(() => toLineInputs(transaction));
  const [correctionNote, setCorrectionNote] = useState("");
  const [error, setError] = useState<string | null>(null);

  const isPending = createMutation.isPending || updateMutation.isPending;

  const debitTotal = lines
    .filter((line) => line.type === "debit")
    .reduce((sum, line) => sum + (Number(line.amount) || 0), 0);
  const creditTotal = lines
    .filter((line) => line.type === "credit")
    .reduce((sum, line) => sum + (Number(line.amount) || 0), 0);
  const isBalanced = lines.length >= 2 && debitTotal === creditTotal && debitTotal > 0;

  function updateLine(index: number, patch: Partial<LineInput>) {
    setLines((current) => current.map((line, i) => (i === index ? { ...line, ...patch } : line)));
  }

  function removeLine(index: number) {
    setLines((current) => current.filter((_, i) => i !== index));
  }

  function addLine(type: EntryType) {
    setLines((current) => [...current, EMPTY_LINE(type)]);
  }

  function resetAndClose() {
    setDate(transaction?.date ?? new Date().toISOString().slice(0, 10));
    setDescription(transaction?.description ?? "");
    setLines(toLineInputs(transaction));
    setCorrectionNote("");
    setError(null);
    onClose();
  }

  function handleSubmit() {
    setError(null);

    const parsed = transactionSchema.safeParse({
      date,
      description,
      journalEntries: lines
        .filter((line) => line.chartOfAccountsId && line.amount)
        .map((line) => ({
          chartOfAccountsId: Number(line.chartOfAccountsId),
          type: line.type,
          amount: Number(line.amount),
          description: line.description,
        })),
    });

    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? "入力内容を確認してください");
      return;
    }

    const input = { ...parsed.data, correctionNote: correctionNote || undefined };
    const onSuccess = () => {
      showToast(isEdit ? "取引を更新しました" : "取引を保存しました");
      resetAndClose();
    };
    const onError = (mutationError: unknown) => {
      showToast(mutationError instanceof Error ? mutationError.message : "保存に失敗しました", "error");
    };

    if (isEdit && transaction) {
      updateMutation.mutate({ id: transaction.id, input }, { onSuccess, onError });
    } else {
      createMutation.mutate(input, { onSuccess, onError });
    }
  }

  return (
    <Modal
      isOpen={open}
      onClose={resetAndClose}
      size="large"
      title={isEdit ? "取引を編集" : "取引を追加"}
      description="複式簿記のルールに従い、借方・貸方の明細を入力してください。"
      footer={
        <>
          <Button variant="secondary" className="w-auto" onClick={resetAndClose}>
            キャンセル
          </Button>
          <Button className="w-auto" onClick={handleSubmit} disabled={isPending || !isBalanced}>
            {isPending ? "保存中..." : isEdit ? "更新" : "保存"}
          </Button>
        </>
      }
    >
      <div className="flex flex-col gap-4">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Input label="日付" type="date" value={date} onChange={(e) => setDate(e.target.value)} required />
          <Input
            label="説明（任意）"
            value={description}
            maxLength={100}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="この取引について"
          />
        </div>

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <span className="text-sm font-semibold">仕訳明細</span>
            <span className="text-xs text-muted-foreground">
              借方: {debitTotal.toLocaleString()}円 / 貸方: {creditTotal.toLocaleString()}円
            </span>
          </div>

          {!isBalanced && (debitTotal > 0 || creditTotal > 0) ? (
            <p className="rounded-lg bg-warning-bg px-3 py-2 text-xs text-warning">
              借方合計と貸方合計が一致していません
            </p>
          ) : null}

          <div className="flex flex-col gap-2">
            {lines.map((line, index) => (
              <div key={index} className="grid grid-cols-1 items-end gap-2 sm:grid-cols-[6rem_1fr_7rem_1fr_auto]">
                <Select
                  label={index === 0 ? "種別" : undefined}
                  value={line.type}
                  onChange={(e) => updateLine(index, { type: e.target.value as EntryType })}
                >
                  <option value="debit">借方</option>
                  <option value="credit">貸方</option>
                </Select>
                <Select
                  label={index === 0 ? "勘定科目" : undefined}
                  value={line.chartOfAccountsId}
                  onChange={(e) => updateLine(index, { chartOfAccountsId: e.target.value })}
                >
                  <option value="">選択...</option>
                  {groupAccountsByType(accounts).map(([type, group]) => (
                    <optgroup key={type} label={ACCOUNT_TYPE_LABEL[type]}>
                      {group.map((account) => (
                        <option key={account.id} value={account.id}>
                          {account.name}
                        </option>
                      ))}
                    </optgroup>
                  ))}
                </Select>
                <Input
                  label={index === 0 ? "金額" : undefined}
                  type="number"
                  min={0}
                  value={line.amount}
                  onChange={(e) => updateLine(index, { amount: e.target.value })}
                  placeholder="0"
                />
                <Input
                  label={index === 0 ? "摘要（任意）" : undefined}
                  value={line.description}
                  onChange={(e) => updateLine(index, { description: e.target.value })}
                  placeholder="摘要"
                />
                <Button
                  type="button"
                  variant="secondary"
                  className="w-auto px-2.5 py-2"
                  onClick={() => removeLine(index)}
                  disabled={lines.length <= 2}
                  aria-label="この明細を削除"
                >
                  <TrashIcon className="size-4" />
                </Button>
              </div>
            ))}
          </div>

          <div className="flex gap-2">
            <Button type="button" variant="secondary" className="w-auto" onClick={() => addLine("debit")}>
              借方明細を追加
            </Button>
            <Button type="button" variant="secondary" className="w-auto" onClick={() => addLine("credit")}>
              貸方明細を追加
            </Button>
          </div>
        </div>

        {isEdit ? (
          <Textarea
            label="訂正理由（任意）"
            value={correctionNote}
            onChange={(e) => setCorrectionNote(e.target.value)}
            placeholder="入力すると、元の取引を残したまま新しい訂正取引として記録します"
            maxLength={255}
            rows={2}
          />
        ) : null}

        {error ? <p className="text-sm text-danger">{error}</p> : null}
      </div>
    </Modal>
  );
}
