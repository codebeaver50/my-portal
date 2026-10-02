"use client";

import { Card } from "@/components/ui";
import { Cell, Legend, Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";
import { useCategoricalPalette } from "../_lib/chartColors";
import type { CategoryAmount } from "../_lib/balance";

export function CategoryPieChart({ title, data }: { title: string; data: CategoryAmount[] }) {
  const palette = useCategoricalPalette();

  return (
    <Card>
      <h3 className="text-sm font-semibold">{title}</h3>
      {data.length === 0 ? (
        <p className="py-8 text-center text-sm text-muted-foreground">データがありません</p>
      ) : (
        <ResponsiveContainer width="100%" height={280}>
          <PieChart>
            <Pie data={data} dataKey="value" nameKey="name" outerRadius={90} labelLine={false}>
              {data.map((entry, index) => (
                <Cell key={entry.name} fill={palette[index % palette.length]} />
              ))}
            </Pie>
            <Tooltip
              contentStyle={{ background: "var(--card)", border: "1px solid var(--border)", borderRadius: 8 }}
              formatter={(value) => `${Number(value ?? 0).toLocaleString()}円`}
            />
            <Legend />
          </PieChart>
        </ResponsiveContainer>
      )}
    </Card>
  );
}
