import type { InputHTMLAttributes } from "react";
import { cn } from "./cn";

type ChoiceProps = Omit<InputHTMLAttributes<HTMLInputElement>, "type"> & {
  label: string;
};

function Choice({ type, label, className, ...props }: ChoiceProps & { type: "checkbox" | "radio" }) {
  return (
    <label
      className={cn(
        "inline-flex cursor-pointer items-center gap-2 text-sm has-disabled:cursor-not-allowed has-disabled:opacity-60",
        className,
      )}
    >
      <input type={type} className="size-4 shrink-0 accent-accent" {...props} />
      <span>{label}</span>
    </label>
  );
}

export function Checkbox(props: ChoiceProps) {
  return <Choice type="checkbox" {...props} />;
}

export function Radio(props: ChoiceProps) {
  return <Choice type="radio" {...props} />;
}
