"use client";

import { EditorContent, useEditor } from "@tiptap/react";
import { getTiptapExtensions } from "../_lib/tiptapExtensions";
import { PROSE_CLASSNAME } from "../_lib/proseClassName";

export function PostContent({ content }: { content: string }) {
  const editor = useEditor({
    immediatelyRender: false,
    editable: false,
    extensions: getTiptapExtensions(),
    content: JSON.parse(content),
    editorProps: {
      attributes: { class: PROSE_CLASSNAME },
    },
  });

  return <EditorContent editor={editor} />;
}
