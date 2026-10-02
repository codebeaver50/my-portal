import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { PostForm } from "../../../_components/PostForm";
import { fetchPostBySlug } from "../../../_lib/api";

export const dynamic = "force-dynamic";

export async function generateMetadata({ params }: PageProps<"/demo-portal/simple-cms/edit/[slug]">): Promise<Metadata> {
  const { slug } = await params;
  const post = await fetchPostBySlug(slug, { cache: "no-store" }).catch(() => null);
  if (!post) {
    return { title: "記事が見つかりません | My Portal" };
  }
  return { title: `${post.title}を編集 | Simple CMS | My Portal` };
}

export default async function EditPostPage({ params }: PageProps<"/demo-portal/simple-cms/edit/[slug]">) {
  const { slug } = await params;
  const post = await fetchPostBySlug(slug, { cache: "no-store" });

  if (!post) {
    notFound();
  }

  return <PostForm post={post} />;
}
