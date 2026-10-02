import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { Badge } from "@/components/ui";
import { PostContent } from "../_components/PostContent";
import { fetchPostBySlug } from "../_lib/api";

export const dynamic = "force-dynamic";

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("ja-JP", { year: "numeric", month: "long", day: "numeric" });
}

export async function generateMetadata({ params }: PageProps<"/demo-portal/simple-cms/[slug]">): Promise<Metadata> {
  const { slug } = await params;
  const post = await fetchPostBySlug(slug, { cache: "no-store" }).catch(() => null);
  if (!post) {
    return { title: "記事が見つかりません | My Portal" };
  }
  return { title: `${post.title} | Simple CMS | My Portal` };
}

export default async function SimpleCmsPostPage({ params }: PageProps<"/demo-portal/simple-cms/[slug]">) {
  const { slug } = await params;
  const post = await fetchPostBySlug(slug, { cache: "no-store" });

  if (!post) {
    notFound();
  }

  return (
    <article className="flex flex-col gap-6">
      <Link href="/demo-portal/simple-cms" className="text-sm text-muted-foreground hover:text-foreground">
        ← 記事一覧に戻る
      </Link>

      <header className="flex flex-col gap-3">
        <div className="flex flex-wrap items-center gap-2">
          {post.category ? <Badge>{post.category.name}</Badge> : null}
          <span className="text-xs text-muted-foreground">{formatDate(post.createdAt)}</span>
        </div>
        <h1 className="text-2xl font-extrabold tracking-tight">{post.title}</h1>
      </header>

      <PostContent content={post.content} />
    </article>
  );
}
