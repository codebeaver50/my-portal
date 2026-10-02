"use client";

import { Card } from "@/components/ui";
import { useCategoricalPalette } from "../_lib/chartColors";
import type { CategoryAmount } from "../_lib/balance";

export function CategoryList({ title, data }: { title: string; data: CategoryAmount[] }) {
  const palette = useCategoricalPalette();

  return (
    <Card>
      <h3 className="text-sm font-semibold">{title}</h3>
      {data.length === 0 ? (
        <p className="py-8 text-center text-sm text-muted-foreground">データがありません</p>
      ) : (
        <ul className="flex flex-col gap-2.5">
          {data.map((item, index) => (
            <li key={item.name} className="flex items-center justify-between text-sm">
              <span className="flex items-center gap-2">
                <span
                  className="size-3 rounded-full"
                  style={{ backgroundColor: palette[index % palette.length] }}
                  aria-hidden
                />
                {item.name}
              </span>
              <span className="font-medium">{item.value.toLocaleString()}円</span>
            </li>
          ))}
        </ul>
      )}
    </Card>
  );
}
