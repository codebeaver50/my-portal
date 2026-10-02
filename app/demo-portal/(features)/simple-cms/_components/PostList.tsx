"use client";

import Link from "next/link";
import { Badge, Button, Card, CardHeader, CardTitle, LinkButton, useToast } from "@/components/ui";
import { useDeletePost, usePosts } from "../_lib/useCms";
import type { Post } from "../_lib/types";
import { EditIcon, TrashIcon } from "./icons";

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("ja-JP", { year: "numeric", month: "long", day: "numeric" });
}

export function PostList({ initialPosts, isAdmin = false }: { initialPosts: Post[]; isAdmin?: boolean }) {
  const { data: posts = initialPosts } = usePosts();
  const deleteMutation = useDeletePost();
  const { showToast } = useToast();

  function handleDelete(post: Post) {
    if (!window.confirm(`「${post.title}」を削除しますか？`)) return;

    deleteMutation.mutate(post.id, {
      onSuccess: () => showToast("記事を削除しました"),
      onError: (error) => {
        showToast(error instanceof Error ? error.message : "削除に失敗しました", "error");
      },
    });
  }

  if (posts.length === 0) {
    return <p className="text-sm text-muted-foreground">まだ記事がありません。最初の記事を投稿してみましょう。</p>;
  }

  return (
    <div className="flex flex-col gap-4">
      {posts.map((post) => (
        <Card key={post.id} className="transition-colors hover:border-accent">
          <CardHeader>
            <div className="flex items-start justify-between gap-2">
              <div className="flex flex-wrap items-center gap-2">
                {post.category ? <Badge>{post.category.name}</Badge> : null}
                <span className="text-xs text-muted-foreground">{formatDate(post.createdAt)}</span>
              </div>
              {isAdmin ? (
                <div className="flex shrink-0 gap-2">
                  <LinkButton
                    href={`/demo-portal/simple-cms/edit/${post.slug}`}
                    variant="secondary"
                    className="w-auto px-2.5 py-2"
                    aria-label={`${post.title}を編集`}
                  >
                    <EditIcon className="size-4" />
                  </LinkButton>
                  <Button
                    variant="secondary"
                    className="w-auto px-2.5 py-2"
                    onClick={() => handleDelete(post)}
                    disabled={deleteMutation.isPending}
                    aria-label={`${post.title}を削除`}
                  >
                    <TrashIcon className="size-4" />
                  </Button>
                </div>
              ) : null}
            </div>
            <Link href={`/demo-portal/simple-cms/${post.slug}`} className="hover:underline">
              <CardTitle>{post.title}</CardTitle>
            </Link>
          </CardHeader>
        </Card>
      ))}
    </div>
  );
}
