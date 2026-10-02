import type { ButtonHTMLAttributes, ReactNode } from "react";
import Link from "next/link";
import { cn } from "./cn";

type ButtonVariant = "primary" | "secondary" | "accent" | "danger";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
};

const variantClasses: Record<ButtonVariant, string> = {
  primary:
    "bg-foreground text-background hover:opacity-90 disabled:bg-neutral-bg disabled:text-muted-foreground disabled:opacity-100",
  secondary:
    "border border-border text-foreground hover:bg-neutral-bg disabled:text-muted-foreground disabled:hover:bg-transparent",
  accent:
    "bg-accent text-accent-foreground hover:opacity-90 disabled:bg-neutral-bg disabled:text-muted-foreground disabled:opacity-100",
  danger:
    "bg-danger text-background hover:opacity-90 disabled:bg-neutral-bg disabled:text-muted-foreground disabled:opacity-100",
};

const baseClasses =
  "w-full rounded-lg px-4 py-2.5 text-center text-sm font-semibold transition-colors disabled:cursor-not-allowed";

export function Button({
  variant = "primary",
  className,
  ...props
}: ButtonProps) {
  return (
    <button
      className={cn(baseClasses, variantClasses[variant], className)}
      {...props}
    />
  );
}

type LinkButtonProps = {
  href: string;
  variant?: ButtonVariant;
  className?: string;
  children: ReactNode;
  "aria-label"?: string;
  target?: string;
  rel?: string;
};

export function LinkButton({
  href,
  variant = "primary",
  className,
  children,
  ...rest
}: LinkButtonProps) {
  return (
    <Link
      href={href}
      className={cn(baseClasses, variantClasses[variant], className)}
      {...rest}
    >
      {children}
    </Link>
  );
}
