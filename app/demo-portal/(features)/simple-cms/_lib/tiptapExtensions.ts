import StarterKit from "@tiptap/starter-kit";

// 編集用エディタと閲覧用（読み取り専用）エディタの両方で同じ拡張構成を使う。
// 構成がずれると、保存済みのJSONを閲覧側で正しく描画できなくなる。
export function getTiptapExtensions() {
  return [
    StarterKit.configure({
      heading: { levels: [1, 2, 3] },
      link: {
        openOnClick: false,
        autolink: true,
        HTMLAttributes: { rel: "noopener noreferrer nofollow", class: "underline" },
      },
    }),
  ];
}
