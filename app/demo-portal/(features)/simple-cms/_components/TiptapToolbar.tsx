"use client";

import type { Editor } from "@tiptap/react";
import { cn } from "@/components/ui";

type ToolbarButtonProps = {
  onClick: () => void;
  active?: boolean;
  disabled?: boolean;
  label: string;
  children: React.ReactNode;
};

function ToolbarButton({ onClick, active, disabled, label, children }: ToolbarButtonProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      aria-label={label}
      aria-pressed={active}
      className={cn(
        "rounded-md px-2.5 py-1.5 text-sm font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-40",
        active ? "bg-foreground text-background" : "text-foreground hover:bg-neutral-bg",
      )}
    >
      {children}
    </button>
  );
}

export function TiptapToolbar({ editor }: { editor: Editor | null }) {
  if (!editor) return null;

  const setLink = () => {
    const previousUrl = editor.getAttributes("link").href as string | undefined;
    const url = window.prompt("リンク先URL", previousUrl ?? "https://");
    if (url === null) return;
    if (url === "") {
      editor.chain().focus().extendMarkRange("link").unsetLink().run();
      return;
    }
    editor.chain().focus().extendMarkRange("link").setLink({ href: url }).run();
  };

  return (
    <div className="flex flex-wrap items-center gap-1 rounded-t-lg border border-b-0 border-border bg-neutral-bg/50 p-1.5">
      <ToolbarButton
        label="見出し1"
        active={editor.isActive("heading", { level: 1 })}
        onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()}
      >
        H1
      </ToolbarButton>
      <ToolbarButton
        label="見出し2"
        active={editor.isActive("heading", { level: 2 })}
        onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
      >
        H2
      </ToolbarButton>
      <ToolbarButton
        label="見出し3"
        active={editor.isActive("heading", { level: 3 })}
        onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}
      >
        H3
      </ToolbarButton>

      <span className="mx-1 h-5 w-px bg-border" aria-hidden />

      <ToolbarButton
        label="太字"
        active={editor.isActive("bold")}
        onClick={() => editor.chain().focus().toggleBold().run()}
      >
        <strong>B</strong>
      </ToolbarButton>
      <ToolbarButton
        label="斜体"
        active={editor.isActive("italic")}
        onClick={() => editor.chain().focus().toggleItalic().run()}
      >
        <em>I</em>
      </ToolbarButton>
      <ToolbarButton
        label="取り消し線"
        active={editor.isActive("strike")}
        onClick={() => editor.chain().focus().toggleStrike().run()}
      >
        <span className="line-through">S</span>
      </ToolbarButton>
      <ToolbarButton label="リンク" active={editor.isActive("link")} onClick={setLink}>
        Link
      </ToolbarButton>

      <span className="mx-1 h-5 w-px bg-border" aria-hidden />

      <ToolbarButton
        label="箇条書きリスト"
        active={editor.isActive("bulletList")}
        onClick={() => editor.chain().focus().toggleBulletList().run()}
      >
        •List
      </ToolbarButton>
      <ToolbarButton
        label="番号付きリスト"
        active={editor.isActive("orderedList")}
        onClick={() => editor.chain().focus().toggleOrderedList().run()}
      >
        1.List
      </ToolbarButton>
      <ToolbarButton
        label="引用"
        active={editor.isActive("blockquote")}
        onClick={() => editor.chain().focus().toggleBlockquote().run()}
      >
        &ldquo;&rdquo;
      </ToolbarButton>
      <ToolbarButton
        label="コードブロック"
        active={editor.isActive("codeBlock")}
        onClick={() => editor.chain().focus().toggleCodeBlock().run()}
      >
        {"</>"}
      </ToolbarButton>

      <span className="mx-1 h-5 w-px bg-border" aria-hidden />

      <ToolbarButton
        label="元に戻す"
        disabled={!editor.can().undo()}
        onClick={() => editor.chain().focus().undo().run()}
      >
        ↺
      </ToolbarButton>
      <ToolbarButton
        label="やり直す"
        disabled={!editor.can().redo()}
        onClick={() => editor.chain().focus().redo().run()}
      >
        ↻
      </ToolbarButton>
    </div>
  );
}
