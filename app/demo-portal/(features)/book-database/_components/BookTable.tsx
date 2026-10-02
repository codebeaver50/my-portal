"use client";

import { Table, TableCell, TableHeaderCell, TableRow } from "@/components/ui";
import { getApiBaseUrl } from "../_lib/api";
import type { BookRow } from "../_lib/types";
import { NoImage } from "./NoImage";
import { useFetchBooks } from "./useFetchBooks";

type BookTableProps = {
  initialBooks: BookRow[];
  initialHasMore: boolean;
};

export function BookTable({ initialBooks, initialHasMore }: BookTableProps) {
  const { books, hasNextPage, sentinelRef } = useFetchBooks(initialBooks, initialHasMore);

  if (books.length === 0) {
    return <p className="py-12 text-center text-sm text-foreground/60">書籍がありません</p>;
  }

  return (
    <>
      <Table className="hidden md:table">
        <thead>
          <tr>
            <TableHeaderCell className="w-48 text-center">発行月</TableHeaderCell>
            <TableHeaderCell className="w-30" aria-hidden="true" />
            <TableHeaderCell>書籍タイトル / 著者</TableHeaderCell>
            <TableHeaderCell className="w-48 text-center">出版社</TableHeaderCell>
          </tr>
        </thead>
        <tbody>
          {books.map((book) => (
            <TableRow key={book.id}>
              <TableCell className="text-center tabular-nums text-foreground/70">
                {formatPublishedDate(book.publishedDate)}
              </TableCell>
              <TableCell className="text-center">
                {book.coverImageUrl ? (
                  <a
                    href={`https://www.amazon.co.jp/s?k=${encodeURIComponent(book.title)}`}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <img
                      className="h-32 w-22 object-cover"
                      src={`${getApiBaseUrl()}/api/books/${book.id}/cover`}
                      alt=""
                    />
                  </a>
                ) : (
                  <NoImage />
                )}
              </TableCell>
              <TableCell>
                <div className="text-[15px] font-semibold leading-snug">
                  {book.title}
                  {book.subtitle ? ` ${book.subtitle}` : ""}
                </div>
                {book.authors.length > 0 && (
                  <div className="mt-1 text-[13px] text-foreground/60">{book.authors.join(" / ")}</div>
                )}
              </TableCell>
              <TableCell className="text-center text-foreground/70">{book.publisherName ?? "-"}</TableCell>
            </TableRow>
          ))}
        </tbody>
      </Table>

      <ul className="divide-y divide-border/60 md:hidden">
        {books.map((book) => (
          <li key={book.id} className="flex gap-4 py-5">
            {book.coverImageUrl ? (
              <a
                className="shrink-0"
                href={`https://www.amazon.co.jp/s?k=${encodeURIComponent(book.title)}`}
                target="_blank"
                rel="noopener noreferrer"
              >
                <img
                  className="h-24 w-16 object-cover"
                  src={`${getApiBaseUrl()}/api/books/${book.id}/cover`}
                  alt=""
                />
              </a>
            ) : (
              <NoImage className="h-24 w-16 shrink-0" />
            )}
            <div className="min-w-0 flex-1">
              <div className="text-[13px] tabular-nums text-foreground/60">
                {formatPublishedDate(book.publishedDate)}
                {book.publisherName ? ` ・ ${book.publisherName}` : ""}
              </div>
              <div className="mt-1 text-[15px] font-semibold leading-snug">
                {book.title}
                {book.subtitle ? ` ${book.subtitle}` : ""}
              </div>
              {book.authors.length > 0 && (
                <div className="mt-1 text-[13px] text-foreground/60">{book.authors.join(" / ")}</div>
              )}
            </div>
          </li>
        ))}
      </ul>

      {hasNextPage && (
        <div ref={sentinelRef} className="py-5 text-center text-[13px] text-foreground/60">
          読み込み中...
        </div>
      )}
    </>
  );
}

function formatPublishedDate(date: Date | null): string {
  if (!date) {
    return "-";
  }
  const month = String(date.getUTCMonth() + 1).padStart(2, "0");
  return `${date.getUTCFullYear()}/${month}`;
}
