"use client";

import { useState } from "react";
import {
  Button,
  Card,
  LinkButton,
  Modal,
  Table,
  TableCell,
  TableHeaderCell,
  TableRow,
  useToast,
} from "@/components/ui";
import { useDeleteRecord, useRecords } from "../_lib/useForms";
import { TrashIcon } from "./icons";
import type { Form, FormField, FormRecord, RecordValue } from "../_lib/types";

const PAGE_SIZE = 20;

// 短い値の列は折り返さず、長文になりうる列だけ折り返す。
const CELL_CLASS: Record<FormField["type"], string> = {
  text: "break-words",
  textarea: "max-w-xs whitespace-pre-wrap break-words",
  number: "whitespace-nowrap text-right tabular-nums",
  email: "whitespace-nowrap",
  date: "whitespace-nowrap",
  select: "whitespace-nowrap",
  radio: "whitespace-nowrap",
  checkbox: "break-words",
};

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatValue(field: FormField, value: RecordValue | undefined): string {
  if (value === undefined) return "—";
  if (Array.isArray(value)) return value.length > 0 ? value.join("、") : "—";
  if (field.type === "number" && typeof value === "number") return value.toLocaleString("ja-JP");
  return String(value);
}

export function RecordsTable({ form }: { form: Form }) {
  const [page, setPage] = useState(1);
  const { data, isPending, isError, error, isPlaceholderData } = useRecords(form.id, page, PAGE_SIZE);
  const [deleting, setDeleting] = useState<FormRecord | null>(null);
  const deleteMutation = useDeleteRecord(form.id);
  const { showToast } = useToast();

  function handleDelete() {
    if (!deleting) return;
    deleteMutation.mutate(deleting.id, {
      onSuccess: () => {
        showToast("回答を削除しました");
        setDeleting(null);
        // 最終ページの最後の1件を削除した場合は前のページに戻る。
        if (data && data.records.length === 1 && page > 1) {
          setPage(page - 1);
        }
      },
      onError: (mutationError) => {
        showToast(mutationError.message, "error");
      },
    });
  }

  if (isPending) {
    return <p className="py-10 text-center text-sm text-muted-foreground">読み込み中...</p>;
  }
  if (isError) {
    return <p className="py-10 text-center text-sm text-danger">{error.message}</p>;
  }

  if (data.total === 0) {
    return (
      <Card className="items-center gap-4 py-10 text-center">
        <p className="text-sm text-muted-foreground">まだ回答がありません。</p>
        <LinkButton href={`/demo-portal/form-builder/${form.id}`} className="w-auto">
          最初の回答を送信する
        </LinkButton>
      </Card>
    );
  }

  const from = (data.page - 1) * data.pageSize + 1;
  const to = from + data.records.length - 1;

  return (
    <div className="flex flex-col gap-4">
      <Card className={isPlaceholderData ? "p-0 opacity-60" : "p-0"}>
        <Table>
          <thead>
            <tr>
              <TableHeaderCell className="whitespace-nowrap">回答日時</TableHeaderCell>
              {form.fields.map((field) => (
                <TableHeaderCell key={field.id} className="min-w-32">
                  {field.label}
                </TableHeaderCell>
              ))}
              <TableHeaderCell>
                <span className="sr-only">操作</span>
              </TableHeaderCell>
            </tr>
          </thead>
          <tbody>
            {data.records.map((record) => (
              <TableRow key={record.id}>
                <TableCell className="whitespace-nowrap text-muted-foreground">
                  {formatDateTime(record.createdAt)}
                </TableCell>
                {form.fields.map((field) => (
                  <TableCell
                    key={field.id}
                    className={CELL_CLASS[field.type]}
                  >
                    {formatValue(field, record.values[String(field.id)])}
                  </TableCell>
                ))}
                <TableCell className="text-right">
                  <Button
                    variant="secondary"
                    className="w-auto px-2 py-1.5 text-danger"
                    aria-label="回答を削除"
                    onClick={() => setDeleting(record)}
                  >
                    <TrashIcon className="size-4" />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </tbody>
        </Table>
      </Card>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-muted-foreground">
          {data.total}件中 {from}〜{to}件を表示
        </p>
        <div className="flex gap-2">
          <Button
            variant="secondary"
            className="w-auto"
            disabled={page === 1 || isPlaceholderData}
            onClick={() => setPage(page - 1)}
          >
            前へ
          </Button>
          <Button
            variant="secondary"
            className="w-auto"
            disabled={!data.hasNextPage || isPlaceholderData}
            onClick={() => setPage(page + 1)}
          >
            次へ
          </Button>
        </div>
      </div>

      <Modal
        isOpen={deleting !== null}
        onClose={() => setDeleting(null)}
        title="回答を削除しますか？"
        description={
          deleting ? `${formatDateTime(deleting.createdAt)} の回答を削除します。この操作は取り消せません。` : undefined
        }
        footer={
          <>
            <Button variant="secondary" className="w-auto" onClick={() => setDeleting(null)}>
              キャンセル
            </Button>
            <Button variant="danger" className="w-auto" disabled={deleteMutation.isPending} onClick={handleDelete}>
              削除する
            </Button>
          </>
        }
      >
        {null}
      </Modal>
    </div>
  );
}
