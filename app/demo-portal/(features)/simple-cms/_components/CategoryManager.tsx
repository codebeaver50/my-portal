"use client";

import { useState } from "react";
import { Button, Table, TableCell, TableHeaderCell, TableRow, useToast } from "@/components/ui";
import { useCategories, useDeleteCategory } from "../_lib/useCms";
import { CategoryFormModal } from "./CategoryFormModal";
import { EditIcon, PlusIcon, TrashIcon } from "./icons";
import type { Category } from "../_lib/types";

export function CategoryManager({ initialCategories }: { initialCategories: Category[] }) {
  const { data: categories = initialCategories } = useCategories();
  const deleteMutation = useDeleteCategory();
  const { showToast } = useToast();

  const [modalOpen, setModalOpen] = useState(false);
  const [editingCategory, setEditingCategory] = useState<Category | undefined>(undefined);

  function openCreateModal() {
    setEditingCategory(undefined);
    setModalOpen(true);
  }

  function openEditModal(category: Category) {
    setEditingCategory(category);
    setModalOpen(true);
  }

  function handleDelete(category: Category) {
    if (!window.confirm(`「${category.name}」を削除しますか？`)) return;

    deleteMutation.mutate(category.id, {
      onSuccess: () => showToast("カテゴリを削除しました"),
      onError: (error) => {
        showToast(error instanceof Error ? error.message : "削除に失敗しました", "error");
      },
    });
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex justify-end">
        <Button className="w-auto" onClick={openCreateModal}>
          <span className="inline-flex items-center gap-1.5">
            <PlusIcon className="size-4" />
            カテゴリを追加
          </span>
        </Button>
      </div>

      {categories.length === 0 ? (
        <p className="text-sm text-muted-foreground">まだカテゴリがありません。最初のカテゴリを追加してください。</p>
      ) : (
        <Table>
          <thead>
            <TableRow>
              <TableHeaderCell>カテゴリ名</TableHeaderCell>
              <TableHeaderCell>URLスラッグ</TableHeaderCell>
              <TableHeaderCell aria-label="操作" />
            </TableRow>
          </thead>
          <tbody>
            {categories.map((category) => (
              <TableRow key={category.id}>
                <TableCell className="font-medium">{category.name}</TableCell>
                <TableCell className="text-muted-foreground">{category.slug}</TableCell>
                <TableCell>
                  <div className="flex justify-end gap-2">
                    <Button
                      variant="secondary"
                      className="w-auto px-2.5 py-2"
                      onClick={() => openEditModal(category)}
                      aria-label={`${category.name}を編集`}
                    >
                      <EditIcon className="size-4" />
                    </Button>
                    <Button
                      variant="secondary"
                      className="w-auto px-2.5 py-2"
                      onClick={() => handleDelete(category)}
                      disabled={deleteMutation.isPending}
                      aria-label={`${category.name}を削除`}
                    >
                      <TrashIcon className="size-4" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </tbody>
        </Table>
      )}

      <CategoryFormModal
        key={editingCategory?.id ?? "create"}
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        category={editingCategory}
      />
    </div>
  );
}
