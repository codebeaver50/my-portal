import { cn } from "@/components/ui/cn";

type NoImageProps = {
  className?: string;
};

export function NoImage({ className }: NoImageProps) {
  return (
    <div
      className={cn(
        "inline-flex h-32 w-22 items-center justify-center bg-black/10 text-xs text-foreground/60 dark:bg-white/10",
        className,
      )}
    >
      NoImage
    </div>
  );
}
