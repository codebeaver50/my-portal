import type { ReactNode, SelectHTMLAttributes } from "react";
import { cn } from "./cn";

type SelectProps = SelectHTMLAttributes<HTMLSelectElement> & {
  label?: string;
  error?: string;
  children: ReactNode;
};

export function Select({ label, error, className, id, children, ...props }: SelectProps) {
  return (
    <label className="flex flex-col gap-1.5 text-sm" htmlFor={id}>
      {label ? <span className="font-medium">{label}</span> : null}
      <span className="relative flex">
        {/* ブラウザ標準の矢印は右端に寄りすぎるため、appearance-none で消して余白を取った矢印を重ねる */}
        <select
          id={id}
          className={cn(
            "w-full appearance-none rounded-lg border border-border bg-card py-2 pr-9 pl-3 text-sm text-foreground outline-none focus:border-accent disabled:cursor-not-allowed disabled:opacity-60",
            error ? "border-danger" : undefined,
            className,
          )}
          {...props}
        >
          {children}
        </select>
        <svg
          aria-hidden="true"
          viewBox="0 0 16 16"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.75"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="pointer-events-none absolute top-1/2 right-3 size-4 -translate-y-1/2 text-muted-foreground"
        >
          <path d="m4 6 4 4 4-4" />
        </svg>
      </span>
      {error ? <span className="text-xs text-danger">{error}</span> : null}
    </label>
  );
}
