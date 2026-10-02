"use client";

import { useState } from "react";
import {
  Badge,
  Button,
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
  LinkButton,
  Modal,
  useToast,
} from "@/components/ui";
import { useDeleteForm, useForms } from "../_lib/useForms";
import { PlusIcon, TrashIcon } from "./icons";
import type { FormSummary } from "../_lib/types";

export function FormListView() {
  const { data: forms, isPending, isError, error } = useForms();
  const [deleting, setDeleting] = useState<FormSummary | null>(null);
  const deleteMutation = useDeleteForm();
  const { showToast } = useToast();

  function handleDelete() {
    if (!deleting) return;
    deleteMutation.mutate(deleting.id, {
      onSuccess: () => {
        showToast("フォームを削除しました");
        setDeleting(null);
      },
      onError: (mutationError) => {
        showToast(mutationError.message, "error");
      },
    });
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-xl font-bold">フォーム一覧</h2>
        <LinkButton href="/form-builder/new" className="w-auto">
          <span className="flex items-center gap-1.5">
            <PlusIcon className="size-4" />
            フォームを作成
          </span>
        </LinkButton>
      </div>

      {isPending ? (
        <p className="py-10 text-center text-sm text-muted-foreground">読み込み中...</p>
      ) : isError ? (
        <p className="py-10 text-center text-sm text-danger">{error.message}</p>
      ) : forms.length === 0 ? (
        <p className="py-10 text-center text-sm text-muted-foreground">
          フォームがまだありません。「フォームを作成」から追加してください。
        </p>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {forms.map((form) => (
            <Card key={form.id}>
              <CardHeader>
                <CardTitle className="break-words">{form.title}</CardTitle>
                {form.description ? <CardDescription>{form.description}</CardDescription> : null}
                <div className="flex flex-wrap gap-2">
                  <Badge variant="neutral" dot={false}>
                    項目 {form.fieldCount}
                  </Badge>
                  <Badge variant="accent" dot={false}>
                    回答 {form.recordCount}件
                  </Badge>
                </div>
              </CardHeader>
              <div className="flex flex-wrap gap-2">
                <LinkButton href={`/form-builder/${form.id}`} className="w-auto flex-1">
                  回答する
                </LinkButton>
                <LinkButton href={`/form-builder/${form.id}/records`} variant="secondary" className="w-auto flex-1">
                  回答データ
                </LinkButton>
                <LinkButton href={`/form-builder/${form.id}/edit`} variant="secondary" className="w-auto">
                  編集
                </LinkButton>
                <Button
                  variant="secondary"
                  className="w-auto text-danger"
                  aria-label={`${form.title}を削除`}
                  onClick={() => setDeleting(form)}
                >
                  <TrashIcon className="size-4" />
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}

      <Modal
        isOpen={deleting !== null}
        onClose={() => setDeleting(null)}
        title="フォームを削除しますか？"
        description={
          deleting
            ? `「${deleting.title}」と、集まった回答${deleting.recordCount}件がすべて削除されます。この操作は取り消せません。`
            : undefined
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
