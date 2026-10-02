"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/components/ui";

type TabNavProps = {
  items: { href: string; label: string }[];
};

export function TabNav({ items }: TabNavProps) {
  const pathname = usePathname();

  return (
    <nav className="flex gap-1 overflow-x-auto border-b border-border">
      {items.map((item) => {
        const isActive = pathname === item.href;
        return (
          <Link
            key={item.href}
            href={item.href}
            className={cn(
              "border-b-2 px-4 py-3 text-sm font-medium whitespace-nowrap transition-colors",
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
