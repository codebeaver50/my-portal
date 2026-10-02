"use client";

import { EditorContent, useEditor } from "@tiptap/react";
import { useEffect } from "react";
import { getTiptapExtensions } from "../_lib/tiptapExtensions";
import { PROSE_CLASSNAME } from "../_lib/proseClassName";
import { TiptapToolbar } from "./TiptapToolbar";
import { cn } from "@/components/ui";

type TiptapEditorProps = {
  content: string;
  onChange: (json: string, isEmpty: boolean) => void;
  error?: string;
};

export function TiptapEditor({ content, onChange, error }: TiptapEditorProps) {
  const editor = useEditor({
    // Next.jsのSSRとの水和不一致を避けるため、初回描画はクライアント側でのみ行う。
    immediatelyRender: false,
    extensions: getTiptapExtensions(),
    content: content ? JSON.parse(content) : "",
    editorProps: {
      attributes: {
        class: cn(PROSE_CLASSNAME, "min-h-48 rounded-b-lg border border-border bg-card px-4 py-3 outline-none"),
      },
    },
    onUpdate: ({ editor }) => {
      onChange(JSON.stringify(editor.getJSON()), editor.isEmpty);
    },
  });

  // 記事編集ページのような初期コンテンツの非同期読み込みに対応する。
  useEffect(() => {
    if (!editor) return;
    const next = content ? JSON.parse(content) : "";
    if (JSON.stringify(editor.getJSON()) !== JSON.stringify(next)) {
      editor.commands.setContent(next);
    }
    // contentの変更時のみ同期したいので、editorは依存配列に含めない。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [content]);

  return (
    <div className="flex flex-col gap-1.5 text-sm">
      <span className="font-medium">本文</span>
      <TiptapToolbar editor={editor} />
      <EditorContent editor={editor} />
      {error ? <span className="text-xs text-danger">{error}</span> : null}
    </div>
  );
}
