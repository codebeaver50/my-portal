import type { Metadata } from "next";
import { PostForm } from "../../_components/PostForm";

export const metadata: Metadata = {
  title: "記事を投稿 | Simple CMS | My Portal",
};

export default function NewPostPage() {
  return <PostForm />;
}
