"use client";

import { useInfiniteQuery } from "@tanstack/react-query";
import { useEffect } from "react";
import { useInView } from "react-intersection-observer";
import { fetchBookRows } from "./api";
import type { BookRow } from "./types";

export function useFetchBooks(initialBooks: BookRow[], initialHasMore: boolean) {
  const { ref: sentinelRef, inView } = useInView();
  const { data, hasNextPage, isFetchingNextPage, fetchNextPage } = useInfiniteQuery({
    queryKey: ["book-database", "books"],
    queryFn: ({ pageParam }) => fetchBookRows(pageParam),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.hasMore ? allPages.reduce((sum, page) => sum + page.rows.length, 0) : undefined,
    initialData: {
      pages: [{ rows: initialBooks, hasMore: initialHasMore }],
      pageParams: [0],
    },
  });

  // 一覧の末尾の要素が画面に入ったこと（IntersectionObserver）に応じて、次のページを取得する。
  useEffect(() => {
    if (inView && hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [inView, hasNextPage, isFetchingNextPage, fetchNextPage]);

  return {
    books: data.pages.flatMap((page) => page.rows),
    hasNextPage,
    sentinelRef,
  };
}
