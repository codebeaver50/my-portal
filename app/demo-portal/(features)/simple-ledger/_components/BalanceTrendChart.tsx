"use client";

import { Card } from "@/components/ui";
import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { useTrendColors } from "../_lib/chartColors";
import type { MonthlyBalancePoint } from "../_lib/balance";

export function BalanceTrendChart({ data }: { data: MonthlyBalancePoint[] }) {
  const colors = useTrendColors();

  return (
    <Card>
      <h3 className="text-sm font-semibold">月別収支推移</h3>
      {data.length === 0 ? (
        <p className="py-8 text-center text-sm text-muted-foreground">データがありません</p>
      ) : (
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={data} margin={{ left: 8, right: 16, top: 8 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
            <XAxis dataKey="month" stroke="var(--muted-foreground)" fontSize={12} />
            <YAxis stroke="var(--muted-foreground)" fontSize={12} />
            <Tooltip
              contentStyle={{ background: "var(--card)", border: "1px solid var(--border)", borderRadius: 8 }}
              formatter={(value) => `${Number(value ?? 0).toLocaleString()}円`}
            />
            <Legend />
            <Line type="monotone" dataKey="income" name="収入" stroke={colors.income} strokeWidth={2} dot={false} />
            <Line
              type="monotone"
              dataKey="expense"
              name="支出"
              stroke={colors.expense}
              strokeWidth={2}
              strokeDasharray="6 3"
              dot={false}
            />
            <Line
              type="monotone"
              dataKey="balance"
              name="収支"
              stroke={colors.balance}
              strokeWidth={2}
              strokeDasharray="2 3"
              dot={false}
            />
          </LineChart>
        </ResponsiveContainer>
      )}
    </Card>
  );
}
