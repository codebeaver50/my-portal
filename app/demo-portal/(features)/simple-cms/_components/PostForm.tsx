"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Button, Input, Select, useToast } from "@/components/ui";
import { useCategories, useCreatePost, useUpdatePost } from "../_lib/useCms";
import { generateDefaultPostSlug, postSchema } from "../_lib/schema";
import { TiptapEditor } from "./TiptapEditor";
import type { Post } from "../_lib/types";

const EMPTY_DOC = JSON.stringify({ type: "doc", content: [{ type: "paragraph" }] });

export function PostForm({ post }: { post?: Post }) {
  const router = useRouter();
  const { data: categories = [] } = useCategories();
  const createMutation = useCreatePost();
  const updateMutation = useUpdatePost();
  const { showToast } = useToast();
  const isEditing = post !== undefined;

  const [title, setTitle] = useState(post?.title ?? "");
  const [slug, setSlug] = useState(post?.slug ?? "");
  const [categoryId, setCategoryId] = useState(post ? String(post.categoryId) : "");
  const [content, setContent] = useState(post?.content ?? EMPTY_DOC);
  const [contentIsEmpty, setContentIsEmpty] = useState(!post);
  const [error, setError] = useState<string | null>(null);

  const isPending = createMutation.isPending || updateMutation.isPending;

  useEffect(() => {
    // サーバーとクライアントで時刻がずれてハイドレーション不一致になるのを
    // 避けるため、初期値はマウント後にクライアント側でのみ設定する。
    if (!isEditing) {
      // サーバーにはない現在時刻（外部システム）と同期するための正当なeffect。
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setSlug((current) => current || generateDefaultPostSlug());
    }
    // 初回マウント時のみ実行する。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  function handleSubmit() {
    setError(null);

    if (contentIsEmpty) {
      setError("本文を入力してください");
      return;
    }

    const parsed = postSchema.safeParse({
      categoryId: Number(categoryId),
      title,
      slug,
      content,
    });

    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? "入力内容を確認してください");
      return;
    }

    const onSuccess = (savedPost: Post) => {
      showToast(isEditing ? "記事を更新しました" : "記事を保存しました");
      router.push(`/demo-portal/simple-cms/${savedPost.slug}`);
    };
    const onError = (mutationError: unknown) => {
      showToast(mutationError instanceof Error ? mutationError.message : "保存に失敗しました", "error");
    };

    if (isEditing) {
      updateMutation.mutate({ id: post.id, input: parsed.data }, { onSuccess, onError });
    } else {
      createMutation.mutate(parsed.data, { onSuccess, onError });
    }
  }

  return (
    <div className="flex flex-col gap-4">
      {categories.length === 0 ? (
        <p className="rounded-lg bg-warning-bg px-3 py-2 text-xs text-warning">
          カテゴリが1件もありません。先に
          <Link href="/demo-portal/simple-cms/categories" className="underline">
            カテゴリ管理ページ
          </Link>
          でカテゴリを作成してください。
        </p>
      ) : null}

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Input
          label="タイトル"
          value={title}
          maxLength={255}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="記事のタイトル"
        />
        <Select label="カテゴリ" value={categoryId} onChange={(e) => setCategoryId(e.target.value)}>
          <option value="">選択...</option>
          {categories.map((category) => (
            <option key={category.id} value={category.id}>
              {category.name}
            </option>
          ))}
        </Select>
      </div>

      <Input
        label="URLスラッグ"
        value={slug}
        maxLength={255}
        onChange={(e) => setSlug(e.target.value)}
        placeholder="my-post"
      />
      <p className="-mt-3 text-xs text-muted-foreground">
        公開URL: /simple-cms/{slug || "my-post"}（小文字英数字とハイフンのみ）
      </p>

      <TiptapEditor
        content={content}
        onChange={(json, isEmpty) => {
          setContent(json);
          setContentIsEmpty(isEmpty);
        }}
      />

      {error ? <p className="text-sm text-danger">{error}</p> : null}

      <Button className="w-auto" onClick={handleSubmit} disabled={isPending || categories.length === 0}>
        {isPending ? "保存中..." : isEditing ? "更新する" : "記事を公開"}
      </Button>
    </div>
  );
}
