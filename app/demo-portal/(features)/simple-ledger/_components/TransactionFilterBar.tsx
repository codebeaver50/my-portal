import { Card, Input, Select } from "@/components/ui";
import type { TransactionCategoryFilter } from "../_lib/types";

type TransactionFilterBarProps = {
  searchValue: string;
  onSearchChange: (value: string) => void;
  categoryValue: TransactionCategoryFilter;
  onCategoryChange: (value: TransactionCategoryFilter) => void;
};

export function TransactionFilterBar({
  searchValue,
  onSearchChange,
  categoryValue,
  onCategoryChange,
}: TransactionFilterBarProps) {
  return (
    <Card className="flex-row flex-wrap items-end gap-4">
      <div className="min-w-[12rem] flex-1">
        <Input
          label="検索"
          value={searchValue}
          onChange={(e) => onSearchChange(e.target.value)}
          placeholder="説明で検索..."
        />
      </div>
      <div className="w-40">
        <Select
          label="カテゴリ"
          value={categoryValue}
          onChange={(e) => onCategoryChange(e.target.value as TransactionCategoryFilter)}
        >
          <option value="all">すべて</option>
          <option value="income">収入</option>
          <option value="expense">支出</option>
        </Select>
      </div>
    </Card>
  );
}
