import type { Metadata } from "next";
import { checkIsAdmin } from "@/app/_lib/adminAuth";
import { AdminHeader } from "./_components/AdminHeader";
import { PostList } from "./_components/PostList";
import { fetchPosts } from "./_lib/api";
import type { Post } from "./_lib/types";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Simple CMS | My Portal",
  description: "カテゴリ分けとリッチテキスト編集ができる記事投稿デモの一覧",
};

async function loadInitialPosts(): Promise<Post[]> {
  try {
    return await fetchPosts({ cache: "no-store" });
  } catch (error) {
    // Go APIが未起動/未接続でもポータルトップ経由でページ自体は開けるようにする
    console.error("simple-cms: failed to fetch initial posts", error);
    return [];
  }
}

export default async function SimpleCmsPage() {
  const [posts, isAdmin] = await Promise.all([loadInitialPosts(), checkIsAdmin()]);

  return (
    <>
      {isAdmin ? <AdminHeader /> : null}
      <PostList initialPosts={posts} isAdmin={isAdmin} />
    </>
  );
}
