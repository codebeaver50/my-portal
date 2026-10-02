import type { Metadata } from "next";
import { PageContainer, PageHeader } from "@/components/ui";
import { BookTable } from "./_components/BookTable";
import { fetchBookRows } from "./_lib/api";
import type { BookRowsPage } from "./_lib/api";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "書籍データベース | My Portal",
  description: "IT関連書籍リスト",
};

async function loadInitialBooks(): Promise<BookRowsPage> {
  try {
    return await fetchBookRows(0, { cache: "no-store" });
  } catch (error) {
    // Go APIが未起動/未接続でもポータルトップ経由でページ自体は開けるようにする
    console.error("book-database: failed to fetch initial books", error);
    return { rows: [], hasMore: false };
  }
}

export default async function BookDatabasePage() {
  const { rows, hasMore } = await loadInitialBooks();

  return (
    <PageContainer>
      <PageHeader title="IT関連書籍リスト" />
      <BookTable initialBooks={rows} initialHasMore={hasMore} />
    </PageContainer>
  );
}
