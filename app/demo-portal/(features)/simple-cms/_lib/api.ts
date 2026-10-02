import type { Category, CategoryInput, Post, PostInput } from "./types";

// Next.js and the Go API run in separate Docker containers, so both
// server- and client-side calls go through this same public base URL
// (Caddy reverse-proxies /api/* on that host to the api container).
export function getApiBaseUrl(): string {
  return process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
}

async function parseErrorMessage(response: Response, fallback: string): Promise<string> {
  const body = await response.json().catch(() => null);
  return (body && typeof body.error === "string" && body.error) || fallback;
}

export async function fetchCategories(init?: RequestInit): Promise<Category[]> {
  const response = await fetch(`${getApiBaseUrl()}/api/simple-cms/categories`, init);
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "カテゴリの取得に失敗しました"));
  }
  const data: { categories: Category[] } = await response.json();
  return data.categories;
}

export async function createCategory(input: CategoryInput): Promise<Category> {
  const response = await fetch(`${getApiBaseUrl()}/api/simple-cms/categories`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "カテゴリの作成に失敗しました"));
  }
  return response.json();
}

export async function updateCategory(id: number, input: CategoryInput): Promise<Category> {
  const response = await fetch(`${getApiBaseUrl()}/api/simple-cms/categories/${id}`, {
    method: "PUT",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "カテゴリの更新に失敗しました"));
  }
  return response.json();
}

export async function deleteCategory(id: number): Promise<void> {
  const response = await fetch(`${getApiBaseUrl()}/api/simple-cms/categories/${id}`, {
    method: "DELETE",
    credentials: "include",
  });
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "カテゴリの削除に失敗しました"));
  }
}

export async function fetchPosts(init?: RequestInit): Promise<Post[]> {
  const response = await fetch(`${getApiBaseUrl()}/api/simple-cms/posts`, init);
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "記事の取得に失敗しました"));
  }
  const data: { posts: Post[] } = await response.json();
  return data.posts;
}

export async function fetchPostBySlug(slug: string, init?: RequestInit): Promise<Post | null> {
  const response = await fetch(`${getApiBaseUrl()}/api/simple-cms/posts/${slug}`, init);
  if (response.status === 404) {
    return null;
  }
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "記事の取得に失敗しました"));
  }
  return response.json();
}

export async function createPost(input: PostInput): Promise<Post> {
  const response = await fetch(`${getApiBaseUrl()}/api/simple-cms/posts`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "記事の作成に失敗しました"));
  }
  return response.json();
}

export async function updatePost(id: number, input: PostInput): Promise<Post> {
  const response = await fetch(`${getApiBaseUrl()}/api/simple-cms/posts/${id}`, {
    method: "PUT",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "記事の更新に失敗しました"));
  }
  return response.json();
}

export async function deletePost(id: number): Promise<void> {
  const response = await fetch(`${getApiBaseUrl()}/api/simple-cms/posts/${id}`, {
    method: "DELETE",
    credentials: "include",
  });
  if (!response.ok) {
    throw new Error(await parseErrorMessage(response, "記事の削除に失敗しました"));
  }
}
