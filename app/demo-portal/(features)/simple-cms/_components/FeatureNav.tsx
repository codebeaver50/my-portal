"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/components/ui";

const NAV_ITEMS = [
  { href: "/demo-portal/simple-cms", label: "記事一覧" },
  { href: "/demo-portal/simple-cms/categories", label: "カテゴリ管理" },
];

export function FeatureNav() {
  const pathname = usePathname();

  return (
    <nav className="flex gap-1 border-b border-border">
      {NAV_ITEMS.map((item) => {
        const isActive = pathname === item.href;
        return (
          <Link
            key={item.href}
            href={item.href}
            className={cn(
              "border-b-2 px-4 py-3 text-sm font-medium transition-colors",
              isActive
                ? "border-accent text-accent"
                : "border-transparent text-muted-foreground hover:text-foreground",
            )}
          >
            {item.label}
          </Link>
        );
      })}
    </nav>
  );
}
