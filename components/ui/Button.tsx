import type { ButtonHTMLAttributes, ReactNode } from "react";
import Link from "next/link";
import { cn } from "./cn";
import { Tooltip } from "./Tooltip";

type ButtonVariant = "primary" | "secondary" | "accent" | "danger";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
  // 指定すると押せないボタンにし、理由をツールチップで出す（権限がなく保存できない場合等）。
  // disabled と違ってフォーカスでき、キーボード・スクリーンリーダーでも理由が伝わる。
  blockedReason?: string;
};

const variantClasses: Record<ButtonVariant, string> = {
  primary:
    "bg-foreground text-background hover:opacity-90 disabled:bg-neutral-bg disabled:text-muted-foreground disabled:opacity-100 aria-disabled:bg-neutral-bg aria-disabled:text-muted-foreground aria-disabled:opacity-100",
  secondary:
    // 文字色を薄くするだけだと、アイコンだけのボタン（削除等）では押せないことが分からないため、
    // 枠線を消してグレーの塗りにし、白地に枠線の押せるボタンと形で見分けられるようにする（primary 等の無効時と同じ塗り）
    "border border-border text-foreground hover:bg-neutral-bg disabled:border-transparent disabled:bg-neutral-bg disabled:text-muted-foreground/60 aria-disabled:border-transparent aria-disabled:bg-neutral-bg aria-disabled:text-muted-foreground/60",
  accent:
    "bg-accent text-accent-foreground hover:opacity-90 disabled:bg-neutral-bg disabled:text-muted-foreground disabled:opacity-100 aria-disabled:bg-neutral-bg aria-disabled:text-muted-foreground aria-disabled:opacity-100",
  danger:
    "bg-danger text-background hover:opacity-90 disabled:bg-neutral-bg disabled:text-muted-foreground disabled:opacity-100 aria-disabled:bg-neutral-bg aria-disabled:text-muted-foreground aria-disabled:opacity-100",
};

const baseClasses =
  "w-full rounded-lg px-4 py-2.5 text-center text-sm font-semibold transition-colors disabled:cursor-not-allowed aria-disabled:cursor-not-allowed";

export function Button({
  variant = "primary",
  className,
  blockedReason,
  ...props
}: ButtonProps) {
  if (blockedReason) {
    return (
      <Tooltip content={blockedReason}>
        <button
          className={cn(baseClasses, variantClasses[variant], className)}
          {...props}
          // disabled にするとフォーカスできず、理由を読めなくなるため外す
          disabled={false}
          aria-disabled="true"
          // 押しても処理（onClick・フォームの送信）を行わない
          onClick={(event) => event.preventDefault()}
        />
      </Tooltip>
    );
  }

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
