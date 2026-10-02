import type { BookRow } from "./types";

// Next.js and the Go API run in separate Docker containers, so both
// server- and client-side calls go through this same public base URL
// (Caddy reverse-proxies /api/* on that host to the api container).
export function getApiBaseUrl(): string {
  return process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
}

type BooksApiResponse = {
  books: (Omit<BookRow, "publishedDate"> & { publishedDate: string | null })[];
  hasMore: boolean;
};

export type BookRowsPage = {
  rows: BookRow[];
  hasMore: boolean;
};

export async function fetchBookRows(skip: number, init?: RequestInit): Promise<BookRowsPage> {
  const response = await fetch(`${getApiBaseUrl()}/api/books?skip=${skip}`, init);
  if (!response.ok) {
    throw new Error(`failed to fetch books: ${response.status}`);
  }
  const data: BooksApiResponse = await response.json();

  return {
    rows: data.books.map((book) => ({
      ...book,
      publishedDate: book.publishedDate ? new Date(book.publishedDate) : null,
    })),
    hasMore: data.hasMore,
  };
}
