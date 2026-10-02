"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createCategory,
  createPost,
  deleteCategory,
  deletePost,
  fetchCategories,
  fetchPosts,
  updateCategory,
  updatePost,
} from "./api";
import type { CategoryInput, PostInput } from "./types";

const CATEGORIES_KEY = ["simple-cms", "categories"];
const POSTS_KEY = ["simple-cms", "posts"];

export function useCategories() {
  return useQuery({
    queryKey: CATEGORIES_KEY,
    queryFn: () => fetchCategories(),
    staleTime: 60 * 1000,
  });
}

export function usePosts() {
  return useQuery({
    queryKey: POSTS_KEY,
    queryFn: () => fetchPosts(),
  });
}

export function useCreatePost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: PostInput) => createPost(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: POSTS_KEY });
    },
  });
}

export function useUpdatePost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: PostInput }) => updatePost(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: POSTS_KEY });
    },
  });
}

export function useDeletePost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deletePost(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: POSTS_KEY });
    },
  });
}

export function useCreateCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CategoryInput) => createCategory(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CATEGORIES_KEY });
    },
  });
}

export function useUpdateCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: CategoryInput }) => updateCategory(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CATEGORIES_KEY });
    },
  });
}

export function useDeleteCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteCategory(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CATEGORIES_KEY });
    },
  });
}
