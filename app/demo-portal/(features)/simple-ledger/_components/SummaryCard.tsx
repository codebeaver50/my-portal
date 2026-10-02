import { Card } from "@/components/ui";
import { TrendingDownIcon, TrendingUpIcon, WalletIcon } from "./icons";

type SummaryCardVariant = "income" | "expense" | "balance" | "count";

const VARIANT_CONFIG: Record<
  SummaryCardVariant,
  { label: string; icon?: typeof TrendingUpIcon; color: string; background: string }
> = {
  income: { label: "総収入", icon: TrendingUpIcon, color: "text-success", background: "bg-success-bg" },
  expense: { label: "総支出", icon: TrendingDownIcon, color: "text-danger", background: "bg-danger-bg" },
  balance: { label: "収支", icon: WalletIcon, color: "text-accent", background: "bg-neutral-bg" },
  count: { label: "総取引数", color: "text-foreground", background: "bg-neutral-bg" },
};

type SummaryCardProps = {
  variant: SummaryCardVariant;
  value: number;
  unit?: string;
};

export function SummaryCard({ variant, value, unit = "円" }: SummaryCardProps) {
  const config = VARIANT_CONFIG[variant];
  const Icon = config.icon;

  return (
    <Card className="gap-3">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-muted-foreground">{config.label}</span>
        {Icon ? (
          <span className={`rounded-md p-2 ${config.background}`}>
            <Icon className={`size-5 ${config.color}`} />
          </span>
        ) : null}
      </div>
      <span className={`text-2xl font-bold ${config.color}`}>
        {value.toLocaleString()} {unit}
      </span>
    </Card>
  );
}
