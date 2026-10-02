"use client";

import { useInfiniteQuery } from "@tanstack/react-query";
import { useEffect } from "react";
import { useInView } from "react-intersection-observer";
import { fetchBookRows } from "../_lib/api";
import type { BookRow } from "../_lib/types";

export function useFetchBooks(initialBooks: BookRow[], initialHasMore: boolean) {
  const { ref: sentinelRef, inView } = useInView();
  const { data, hasNextPage, isFetchingNextPage, fetchNextPage } = useInfiniteQuery({
    queryKey: ["books"],
    queryFn: ({ pageParam }) => fetchBookRows(pageParam),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.hasMore ? allPages.reduce((sum, page) => sum + page.rows.length, 0) : undefined,
    initialData: {
      pages: [{ rows: initialBooks, hasMore: initialHasMore }],
      pageParams: [0],
    },
  });

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
