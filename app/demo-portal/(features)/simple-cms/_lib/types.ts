export type Category = {
  id: number;
  name: string;
  slug: string;
};

// content: Tiptapのドキュメントを表すProseMirror JSON文字列（JSON.parseして
// @tiptap/react / @tiptap/htmlに渡す）。
export type Post = {
  id: number;
  categoryId: number;
  category?: Category;
  title: string;
  slug: string;
  content: string;
  createdAt: string;
  updatedAt: string;
};

export type CategoryInput = {
  name: string;
  slug: string;
};

export type PostInput = {
  categoryId: number;
  title: string;
  slug: string;
  content: string;
};
