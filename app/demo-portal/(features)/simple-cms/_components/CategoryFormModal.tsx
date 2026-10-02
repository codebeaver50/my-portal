"use client";

import { useState } from "react";
import { Button, Input, Modal, useToast } from "@/components/ui";
import { useCreateCategory, useUpdateCategory } from "../_lib/useCms";
import { categorySchema } from "../_lib/schema";
import type { Category } from "../_lib/types";

type CategoryFormModalProps = {
  open: boolean;
  onClose: () => void;
  category?: Category;
};

export function CategoryFormModal({ open, onClose, category }: CategoryFormModalProps) {
  const isEdit = Boolean(category);
  const createMutation = useCreateCategory();
  const updateMutation = useUpdateCategory();
  const { showToast } = useToast();

  const [name, setName] = useState(category?.name ?? "");
  const [slug, setSlug] = useState(category?.slug ?? "");
  const [error, setError] = useState<string | null>(null);

  const isPending = createMutation.isPending || updateMutation.isPending;

  function resetAndClose() {
    setName(category?.name ?? "");
    setSlug(category?.slug ?? "");
    setError(null);
    onClose();
  }

  function handleSubmit() {
    setError(null);

    const parsed = categorySchema.safeParse({ name, slug });
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? "入力内容を確認してください");
      return;
    }

    const onSuccess = () => {
      showToast(isEdit ? "カテゴリを更新しました" : "カテゴリを作成しました");
      resetAndClose();
    };
    const onError = (mutationError: unknown) => {
      showToast(mutationError instanceof Error ? mutationError.message : "保存に失敗しました", "error");
    };

    if (isEdit && category) {
      updateMutation.mutate({ id: category.id, input: parsed.data }, { onSuccess, onError });
    } else {
      createMutation.mutate(parsed.data, { onSuccess, onError });
    }
  }

  return (
    <Modal
      isOpen={open}
      onClose={resetAndClose}
      title={isEdit ? "カテゴリを編集" : "カテゴリを追加"}
      footer={
        <>
          <Button variant="secondary" className="w-auto" onClick={resetAndClose}>
            キャンセル
          </Button>
          <Button className="w-auto" onClick={handleSubmit} disabled={isPending}>
            {isPending ? "保存中..." : isEdit ? "更新" : "保存"}
          </Button>
        </>
      }
    >
      <div className="flex flex-col gap-4">
        <Input label="カテゴリ名" value={name} maxLength={100} onChange={(e) => setName(e.target.value)} />
        <Input label="URLスラッグ" value={slug} maxLength={100} onChange={(e) => setSlug(e.target.value)} placeholder="news" />
        {error ? <p className="text-sm text-danger">{error}</p> : null}
      </div>
    </Modal>
  );
}
