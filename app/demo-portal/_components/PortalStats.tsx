import type { Feature } from "../../_lib/features";

export function PortalStats({ features }: { features: Feature[] }) {
  const activeCount = features.filter((feature) => feature.status === "available").length;
  const pad = (value: number) => value.toString().padStart(2, "0");

  return (
    <div className="flex divide-x divide-border rounded-xl border border-border bg-card shadow-sm">
      <div className="flex flex-col gap-1 px-5 py-3">
        <span className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">
          Active Apps
        </span>
        <span className="text-lg font-bold">
          {pad(activeCount)} {"//"} {pad(features.length)}
        </span>
      </div>
      <div className="flex flex-col gap-1 px-5 py-3">
        <span className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">Status</span>
        <span className="flex items-center gap-1.5 text-lg font-bold text-success">
          <span className="size-2 rounded-full bg-success" aria-hidden />
          Online
        </span>
      </div>
    </div>
  );
}
